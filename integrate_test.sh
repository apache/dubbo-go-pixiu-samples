#!/usr/bin/env bash
#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -Eeuo pipefail

if [[ $# -lt 1 ]]; then
  echo "Provide a sample directory, for example: ./integrate_test.sh dubbogo/simple/traffic" >&2
  exit 2
fi

P_DIR=$(cd "$1" && pwd)
PIXIU_CONFIG=${PIXIU_CONFIG:-conf.yaml}
TEST_TAGS=${TEST_TAGS:-integration}
TEST_RUN=${TEST_RUN:-}
PROJECT_NAME=$(basename "$P_DIR")
MODULE_FILE=$(cd "$P_DIR" && go env GOMOD)
SAMPLES_ROOT=$(dirname "$MODULE_FILE")

if [[ -n "${PIXIU_DIR:-}" ]]; then
  PIXIU_DIR=$(cd "$PIXIU_DIR" && pwd)
elif [[ -d "$SAMPLES_ROOT/../cmd/pixiu" ]]; then
  PIXIU_DIR=$(cd "$SAMPLES_ROOT/.." && pwd)
elif [[ -d "$SAMPLES_ROOT/cmd/pixiu" ]]; then
  PIXIU_DIR=$SAMPLES_ROOT
else
  echo "Could not locate the Pixiu source checkout; set PIXIU_DIR to its repository root" >&2
  exit 1
fi

if [[ "$P_DIR" == "$SAMPLES_ROOT/dubbogo/simple/traffic" && "${PIXIU_CI158_LOCK_HELD:-}" != "1" ]]; then
  python3 - "$0" "$@" <<'PY'
import fcntl
import os
import subprocess
import sys
from pathlib import Path

lock_path = Path("/private/tmp/pixiu-ci158-integration.lock")
if not lock_path.parent.exists():
    lock_path = Path("/tmp/pixiu-ci158-integration.lock")
with lock_path.open("a+") as lock_file:
    fcntl.flock(lock_file, fcntl.LOCK_EX)
    child_env = os.environ.copy()
    child_env["PIXIU_CI158_LOCK_HELD"] = "1"
    # The shell keeps the lock through cleanup even if this wrapper exits first.
    result = subprocess.run(
        ["bash", sys.argv[1], *sys.argv[2:]],
        env=child_env,
        pass_fds=(lock_file.fileno(),),
    )
    sys.exit(result.returncode)
PY
  exit $?
fi

PIXIU_SHA=$(git -C "$PIXIU_DIR" rev-parse HEAD 2>/dev/null || echo "unavailable")
SAMPLES_SHA=$(git -C "$SAMPLES_ROOT" rev-parse HEAD 2>/dev/null || echo "unavailable")
printf 'Samples checkout: %s (%s)\nPixiu checkout: %s (%s)\n' "$SAMPLES_ROOT" "$SAMPLES_SHA" "$PIXIU_DIR" "$PIXIU_SHA"

if [[ "$P_DIR" == "$SAMPLES_ROOT/dubbogo/simple/traffic" ]]; then
  RUN_DIR=$(mktemp -d "${TMPDIR:-/tmp}/pixiu-traffic-integration.XXXXXX")
  SERVER_PID=
  PIXIU_PID=
  SERVER_BIN=
  GATEWAY_BIN=

  stop_owned_process() {
    local pid=$1 expected=$2 actual
    [[ -n "$pid" ]] || return 0
    if ! kill -0 "$pid" 2>/dev/null; then
      return 0
    fi
    actual=$(ps -p "$pid" -o command= 2>/dev/null || true)
    case "$actual" in
      "$expected"*) kill "$pid" 2>/dev/null || true ;;
      *) echo "Refusing to stop non-owned PID $pid ($actual)" >&2; return 1 ;;
    esac
    for _ in $(seq 1 40); do
      process_is_running "$pid" || break
      sleep 0.25
    done
    if process_is_running "$pid"; then
      actual=$(ps -p "$pid" -o command= 2>/dev/null || true)
      case "$actual" in
        "$expected"*) kill -KILL "$pid" 2>/dev/null || true ;;
        *) echo "Refusing to force-stop non-owned PID $pid ($actual)" >&2; return 1 ;;
      esac
    fi
    wait "$pid" 2>/dev/null || true
  }

  process_is_running() {
    local state
    kill -0 "$1" 2>/dev/null || return 1
    state=$(ps -p "$1" -o stat= 2>/dev/null) || return 1
    [[ "$state" != *Z* ]]
  }

  backend_health_url() {
    local addr=$1 host port
    host=${addr%:*}
    port=${addr##*:}
    host=${host#[}
    host=${host%]}
    if [[ -z "$host" || "$host" == "0.0.0.0" || "$host" == "::" ]]; then
      host=127.0.0.1
    fi
    if [[ "$host" == *:* ]]; then
      host="[$host]"
    fi
    printf 'http://%s:%s/health' "$host" "$port"
  }

  wait_for_backend() {
    local addr=$1 label=$2 url=$3 ready=0 body
    for _ in $(seq 1 80); do
      if ! process_is_running "$SERVER_PID"; then
        echo "Traffic backend exited during startup" >&2
        return 1
      fi
      body=$(curl --silent --show-error --fail --max-time 1 "$url" 2>/dev/null || true)
      if grep -q "\"server\": \"$label\"" <<<"$body"; then
        ready=1
        break
      fi
      sleep 0.25
    done
    if [[ $ready -ne 1 ]]; then
      echo "Traffic backend $label failed readiness check at $url (listen address: $addr)" >&2
      return 1
    fi
  }

  cleanup_traffic() {
    local result=$?
    trap - EXIT INT TERM
    set +e
    stop_owned_process "$PIXIU_PID" "$GATEWAY_BIN" || result=1
    stop_owned_process "$SERVER_PID" "$SERVER_BIN" || result=1
    if [[ $result -ne 0 ]]; then
      [[ ! -f "$RUN_DIR/server.log" ]] || cat "$RUN_DIR/server.log" >&2
      [[ ! -f "$RUN_DIR/pixiu.log" ]] || cat "$RUN_DIR/pixiu.log" >&2
      echo "Traffic integration failed; build artifacts and logs retained at $RUN_DIR" >&2
    else
      rm -rf "$RUN_DIR"
    fi
    exit "$result"
  }
  trap cleanup_traffic EXIT
  trap 'exit 130' INT
  trap 'exit 143' TERM

  GOOS=$(go env GOOS)
  GOARCH=$(go env GOARCH)
  BASE_DIR="$RUN_DIR/dist"
  OUT_DIR="$BASE_DIR/${GOOS}_${GOARCH}"
  MAKE=(make --no-print-directory "PROJECT_DIR=$P_DIR" "PIXIU_DIR=$PIXIU_DIR" "PROJECT_NAME=$PROJECT_NAME" "BASE_DIR=$BASE_DIR" "GOOS=$GOOS" "GOARCH=$GOARCH" "PIXIU_CONFIG=$PIXIU_CONFIG" "TRAFFIC_CONFIG_GENERATOR=go run ./dubbogo/simple/traffic/configure" "TRAFFIC_MODULE_DIR=$SAMPLES_ROOT" -f "$SAMPLES_ROOT/igt/Makefile")

  echo "> Rendering traffic config and building server"
  "${MAKE[@]}" config
  "${MAKE[@]}" build
  echo "> Building Pixiu from the checked-out source $PIXIU_SHA"
  "${MAKE[@]}" CGO=on buildPixiuBinary

  SERVER_BIN="$OUT_DIR/server/$PROJECT_NAME"
  GATEWAY_BIN="$OUT_DIR/dubbo-go-pixiu"
  if [[ ! -x "$SERVER_BIN" || ! -x "$GATEWAY_BIN" ]]; then
    echo "Expected traffic server and Pixiu binaries were not built" >&2
    exit 1
  fi

  "$SERVER_BIN" >"$RUN_DIR/server.log" 2>&1 &
  SERVER_PID=$!

  for backend in "v1:${TRAFFIC_V1_ADDR:-:1315}" "v2:${TRAFFIC_V2_ADDR:-:1316}" "v3:${TRAFFIC_V3_ADDR:-:1317}"; do
    label=${backend%%:*}
    addr=${backend#*:}
    wait_for_backend "$addr" "$label" "$(backend_health_url "$addr")"
  done

  "$GATEWAY_BIN" gateway start -l debug -c "$OUT_DIR/pixiuconf/conf.yaml" >"$RUN_DIR/pixiu.log" 2>&1 &
  PIXIU_PID=$!

  gateway_url=${TRAFFIC_GATEWAY_URL:-http://localhost:8888}
  gateway_probe_url=$(printf '%s' "$gateway_url" | sed 's#^http://localhost:#http://127.0.0.1:#')
  expected_backend=v2
  if [[ "$PIXIU_CONFIG" == "header-conf.yaml" ]]; then
    expected_backend=v1
  fi
  ready=0
  for _ in $(seq 1 80); do
    if ! process_is_running "$SERVER_PID"; then
      echo "Traffic backend exited during startup" >&2
      cat "$RUN_DIR/server.log" >&2
      exit 1
    fi
    if ! process_is_running "$PIXIU_PID"; then
      echo "Pixiu exited during startup" >&2
      exit 1
    fi
    if [[ "$PIXIU_CONFIG" == "header-conf.yaml" ]]; then
      curl --silent --show-error --fail --max-time 1 -H 'X-A: t1' -H 'X-B: t4' "$gateway_probe_url/user" >"$RUN_DIR/probe.out" 2>/dev/null || true
    else
      curl --silent --show-error --fail --max-time 1 "$gateway_probe_url/user" >"$RUN_DIR/probe.out" 2>/dev/null || true
    fi
    if grep -q "\"server\": \"$expected_backend\"" "$RUN_DIR/probe.out"; then
      ready=1
      break
    fi
    sleep 0.25
  done
  if [[ $ready -ne 1 ]]; then
    echo "Traffic gateway failed readiness check at $gateway_url/user" >&2
    curl --silent --show-error --include --max-time 1 "$gateway_probe_url/user" >&2 || true
    exit 1
  fi
  echo "> Gateway ready at $gateway_url (expected backend $expected_backend)"

  echo "> Running go test -count=1 -tags '$TEST_TAGS' -run '$TEST_RUN' ./dubbogo/simple/traffic/test/..."
  (cd "$SAMPLES_ROOT" && go test -count=1 -tags "$TEST_TAGS" -run "$TEST_RUN" -v ./dubbogo/simple/traffic/test/...)
  exit 0
fi

RUN_DIR=$(mktemp -d "${TMPDIR:-/tmp}/pixiu-samples-integration.XXXXXX")
BASE_DIR="$RUN_DIR/dist"
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
OUT_DIR="$BASE_DIR/${GOOS}_${GOARCH}"
make_args=(--no-print-directory "PROJECT_DIR=$P_DIR" "PIXIU_DIR=$PIXIU_DIR" "PROJECT_NAME=$PROJECT_NAME" "BASE_DIR=$BASE_DIR" "GOOS=$GOOS" "GOARCH=$GOARCH" -f "$SAMPLES_ROOT/igt/Makefile")

cleanup_sample() {
  local result=$?
  trap - EXIT INT TERM
  set +e
  if [[ $result -ne 0 && -f "$OUT_DIR/$PROJECT_NAME.log" ]]; then
    cp "$OUT_DIR/$PROJECT_NAME.log" "$RUN_DIR/server.log"
    cat "$RUN_DIR/server.log" >&2
  fi
  make "${make_args[@]}" clean >/dev/null || result=1
  make "${make_args[@]}" docker-down >/dev/null || result=1
  if [[ $result -ne 0 ]]; then
    echo "Sample integration failed; logs retained at $RUN_DIR" >&2
  else
    rm -rf "$RUN_DIR"
  fi
  exit "$result"
}
trap cleanup_sample EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

make "${make_args[@]}" config
make "${make_args[@]}" docker-up
make "${make_args[@]}" docker-health-check
make "${make_args[@]}" build
make "${make_args[@]}" start
sleep 2
make "${make_args[@]}" CGO=on buildPixiu
SERVER_PID_FILE="$OUT_DIR/server.pid"
PIXIU_PID_FILE="$OUT_DIR/pixiu.pid"
sleep 2
for pid_file in "$SERVER_PID_FILE" "$PIXIU_PID_FILE"; do
  if [[ ! -s "$pid_file" ]] || ! kill -0 "$(cat "$pid_file")" 2>/dev/null; then
    echo "integration service failed to stay alive (PID file: $pid_file)" >&2
    exit 1
  fi
done
make "${make_args[@]}" TEST_TAGS="$TEST_TAGS" TEST_RUN="$TEST_RUN" integration
