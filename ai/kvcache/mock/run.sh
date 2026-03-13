#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PIXIU_CONFIG="${PIXIU_CONFIG:-${SCRIPT_DIR}/pixiu/conf.yaml}"
PIXIU_URL="${PIXIU_URL:-http://127.0.0.1:18888}"
LMCACHE_ADMIN="${LMCACHE_ADMIN:-http://127.0.0.1:18081}"
ENGINE_A_ADMIN="${ENGINE_A_ADMIN:-http://127.0.0.1:18091}"
ENGINE_B_ADMIN="${ENGINE_B_ADMIN:-http://127.0.0.1:18092}"
PIXIU_SOURCE="${PIXIU_SOURCE:-$(cd "${SCRIPT_DIR}/../../../.." && pwd)/dubbo-go-pixiu}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/go-build-cache}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/go-mod-cache}"
export NO_PROXY="${NO_PROXY:-127.0.0.1,localhost}"
export no_proxy="${no_proxy:-127.0.0.1,localhost}"

WORK_DIR="$(mktemp -d /tmp/kvcache-mock-XXXXXX)"
CONTROLLER_LOG="${WORK_DIR}/mock_controller.log"
ENGINE_A_LOG="${WORK_DIR}/mock_engine_a.log"
ENGINE_B_LOG="${WORK_DIR}/mock_engine_b.log"
PIXIU_LOG="${WORK_DIR}/pixiu.log"

CONTROLLER_PID=""
ENGINE_A_PID=""
ENGINE_B_PID=""
PIXIU_PID=""

REQ_BODY='{"model":"mock-model","messages":[{"role":"user","content":"please explain kv cache routing"}]}'

cleanup() {
  set +e
  if [[ -n "${PIXIU_PID}" ]] && kill -0 "${PIXIU_PID}" >/dev/null 2>&1; then
    kill "${PIXIU_PID}" >/dev/null 2>&1
    wait "${PIXIU_PID}" >/dev/null 2>&1
  fi
  if [[ -n "${CONTROLLER_PID}" ]] && kill -0 "${CONTROLLER_PID}" >/dev/null 2>&1; then
    kill "${CONTROLLER_PID}" >/dev/null 2>&1
    wait "${CONTROLLER_PID}" >/dev/null 2>&1
  fi
  if [[ -n "${ENGINE_A_PID}" ]] && kill -0 "${ENGINE_A_PID}" >/dev/null 2>&1; then
    kill "${ENGINE_A_PID}" >/dev/null 2>&1
    wait "${ENGINE_A_PID}" >/dev/null 2>&1
  fi
  if [[ -n "${ENGINE_B_PID}" ]] && kill -0 "${ENGINE_B_PID}" >/dev/null 2>&1; then
    kill "${ENGINE_B_PID}" >/dev/null 2>&1
    wait "${ENGINE_B_PID}" >/dev/null 2>&1
  fi
}
trap cleanup EXIT INT TERM

extract_num() {
  local json="$1"
  local key="$2"
  sed -n "s/.*\"${key}\":\([0-9][0-9]*\).*/\1/p" <<<"${json}"
}

extract_str() {
  local json="$1"
  local key="$2"
  sed -n "s/.*\"${key}\":\"\([^\"]*\)\".*/\1/p" <<<"${json}"
}

wait_for_health() {
  local url="$1"
  for _ in $(seq 1 80); do
    if curl -fsS "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.2
  done
  return 1
}

wait_for_pixiu() {
  for _ in $(seq 1 80); do
    local status
    status="$(curl -s -o "${WORK_DIR}/kvcache-pixiu-ready.out" -w '%{http_code}' \
      -H 'Content-Type: application/json' \
      -X POST "${PIXIU_URL}/v1/chat/completions" \
      -d "${REQ_BODY}" || true)"
    if [[ "${status}" == "200" || "${status}" == "4"* || "${status}" == "5"* ]]; then
      return 0
    fi
    sleep 0.2
  done
  return 1
}

start_mocks() {
  (
    cd "${SCRIPT_DIR}"
    env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" go run ./server/controller >"${CONTROLLER_LOG}" 2>&1
  ) &
  CONTROLLER_PID="$!"

  (
    cd "${SCRIPT_DIR}"
    env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" go run ./server/engine-a >"${ENGINE_A_LOG}" 2>&1
  ) &
  ENGINE_A_PID="$!"

  (
    cd "${SCRIPT_DIR}"
    env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" go run ./server/engine-b >"${ENGINE_B_LOG}" 2>&1
  ) &
  ENGINE_B_PID="$!"

  wait_for_health "${LMCACHE_ADMIN}/health"
  wait_for_health "${ENGINE_A_ADMIN}/health"
  wait_for_health "${ENGINE_B_ADMIN}/health"
}

start_pixiu() {
  if command -v pixiu >/dev/null 2>&1; then
    pixiu gateway start -c "${PIXIU_CONFIG}" >"${PIXIU_LOG}" 2>&1 &
    PIXIU_PID="$!"
  elif [[ -d "${PIXIU_SOURCE}/cmd/pixiu" ]]; then
    (
      cd "${PIXIU_SOURCE}"
      env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" go run ./cmd/pixiu/*.go gateway start -c "${PIXIU_CONFIG}"
    ) >"${PIXIU_LOG}" 2>&1 &
    PIXIU_PID="$!"
  else
    echo "cannot find pixiu binary or source. set PIXIU_SOURCE or install pixiu in PATH"
    return 1
  fi

  if ! wait_for_pixiu; then
    echo "pixiu start failed, log: ${PIXIU_LOG}"
    return 1
  fi
}

echo "[1/4] starting mock controller + engines"
start_mocks

echo "[2/4] starting pixiu"
start_pixiu

curl -fsS -X POST "${LMCACHE_ADMIN}/reset" >/dev/null
curl -fsS -X POST "${ENGINE_A_ADMIN}/reset" >/dev/null
curl -fsS -X POST "${ENGINE_B_ADMIN}/reset" >/dev/null

echo "[3/4] sending warmup and routed requests"
resp1="$(curl -fsS -H 'Content-Type: application/json' -X POST "${PIXIU_URL}/v1/chat/completions" -d "${REQ_BODY}")"
sleep 0.6
resp2="$(curl -fsS -H 'Content-Type: application/json' -X POST "${PIXIU_URL}/v1/chat/completions" -d "${REQ_BODY}")"

sleep 1.0
controller_stats="$(curl -fsS "${LMCACHE_ADMIN}/stats")"
engine_a_stats="$(curl -fsS "${ENGINE_A_ADMIN}/stats")"
engine_b_stats="$(curl -fsS "${ENGINE_B_ADMIN}/stats")"

echo "[4/4] evaluating results"
second_served_by="$(extract_str "${resp2}" "served_by")"
tokenize_calls="$(extract_num "${engine_a_stats}" "tokenize_calls")"
lookup_calls="$(extract_num "${controller_stats}" "lookup_calls")"
pin_calls="$(extract_num "${controller_stats}" "pin_calls")"
llm_b_calls="$(extract_num "${engine_b_stats}" "chat_calls")"

fail=0
if [[ -z "${tokenize_calls}" || "${tokenize_calls}" -lt 1 ]]; then
  echo "FAIL: tokenize was not called on engine-a"
  fail=1
fi
if [[ -z "${lookup_calls}" || "${lookup_calls}" -lt 2 ]]; then
  echo "FAIL: lookup should be called at least twice (cache build + route hint)"
  fail=1
fi
if [[ -z "${pin_calls}" || "${pin_calls}" -lt 1 ]]; then
  echo "FAIL: pin was not called"
  fail=1
fi
if [[ "${second_served_by}" != "mock-llm-b" ]]; then
  echo "FAIL: expected routed request served_by=mock-llm-b, got '${second_served_by}'"
  fail=1
fi
if [[ -z "${llm_b_calls}" || "${llm_b_calls}" -lt 1 ]]; then
  echo "FAIL: preferred endpoint mock-llm-b received no traffic"
  fail=1
fi

echo ""
echo "response#1: ${resp1}"
echo "response#2: ${resp2}"
echo "controller_stats: ${controller_stats}"
echo "engine_a_stats: ${engine_a_stats}"
echo "engine_b_stats: ${engine_b_stats}"
echo "controller log: ${CONTROLLER_LOG}"
echo "engine-a log: ${ENGINE_A_LOG}"
echo "engine-b log: ${ENGINE_B_LOG}"
echo "pixiu log: ${PIXIU_LOG}"

if [[ "${fail}" -ne 0 ]]; then
  exit 1
fi

echo "PASS: kvcache mock sample validated with 3 isolated mocks (controller + two engines)."
