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
PIXIU_SOURCE="${PIXIU_SOURCE:-}"
START_PIXIU="${START_PIXIU:-1}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/go-build-cache}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/go-mod-cache}"
KEEP_WORK_DIR="${KEEP_WORK_DIR:-0}"
export NO_PROXY="${NO_PROXY:-127.0.0.1,localhost}"
export no_proxy="${no_proxy:-127.0.0.1,localhost}"

required_cmds=(curl jq envsubst awk sort)
for cmd in "${required_cmds[@]}"; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "missing required command: ${cmd}"
    exit 1
  fi
done

require_env() {
  local key="$1"
  if [[ -z "${!key:-}" ]]; then
    echo "missing required env: ${key}"
    exit 1
  fi
}

parse_url_scheme() {
  local url="$1"
  if [[ "${url}" == https://* ]]; then
    echo "https"
  else
    echo "http"
  fi
}

parse_url_host() {
  local url="$1"
  local tmp="${url#*://}"
  tmp="${tmp%%/*}"
  echo "${tmp%%:*}"
}

parse_url_port() {
  local url="$1"
  local tmp="${url#*://}"
  tmp="${tmp%%/*}"
  if [[ "${tmp}" == *:* ]]; then
    echo "${tmp##*:}"
    return
  fi
  if [[ "${url}" == https://* ]]; then
    echo "443"
  else
    echo "80"
  fi
}

require_env VLLM_ENDPOINT
require_env LMCACHE_ENDPOINT

export PIXIU_LISTEN_PORT="${PIXIU_LISTEN_PORT:-18889}"
export PIXIU_PROM_PORT="${PIXIU_PROM_PORT:-2223}"
export MODEL_NAME="${MODEL_NAME:-Qwen2.5-3B-Instruct}"
export HOT_CONTENT_THRESHOLD="${HOT_CONTENT_THRESHOLD:-3}"
export LOAD_THRESHOLD="${LOAD_THRESHOLD:-0.7}"
export MEMORY_THRESHOLD="${MEMORY_THRESHOLD:-0.8}"
export LLM_SCHEME="${LLM_SCHEME:-$(parse_url_scheme "${VLLM_ENDPOINT}")}" # override to https if needed

export ENGINE_ENDPOINT_A_ID="${ENGINE_ENDPOINT_A_ID:-engine-a}"
export ENGINE_ENDPOINT_B_ID="${ENGINE_ENDPOINT_B_ID:-engine-b}"
export ENGINE_ENDPOINT_A_HOST="${ENGINE_ENDPOINT_A_HOST:-$(parse_url_host "${VLLM_ENDPOINT}")}" 
export ENGINE_ENDPOINT_A_PORT="${ENGINE_ENDPOINT_A_PORT:-$(parse_url_port "${VLLM_ENDPOINT}")}" 
export ENGINE_ENDPOINT_B_HOST="${ENGINE_ENDPOINT_B_HOST:-${ENGINE_ENDPOINT_A_HOST}}"
export ENGINE_ENDPOINT_B_PORT="${ENGINE_ENDPOINT_B_PORT:-${ENGINE_ENDPOINT_A_PORT}}"

export PIN_INSTANCE_ID="${PIN_INSTANCE_ID:-${ENGINE_ENDPOINT_A_ID}}"
export PIN_LOCATION="${PIN_LOCATION:-lmcache}"
export COMPRESS_INSTANCE_ID="${COMPRESS_INSTANCE_ID:-${ENGINE_ENDPOINT_B_ID}}"
export COMPRESS_LOCATION="${COMPRESS_LOCATION:-lmcache}"
export EVICT_INSTANCE_ID="${EVICT_INSTANCE_ID:-${ENGINE_ENDPOINT_A_ID}}"

PIXIU_URL="${PIXIU_URL:-http://127.0.0.1:${PIXIU_LISTEN_PORT}}"
PROM_URL="${PROM_URL:-http://127.0.0.1:${PIXIU_PROM_PORT}/metrics}"

WORK_DIR="$(mktemp -d /tmp/kvcache-real-verify-XXXXXX)"
RENDERED_CONFIG="${WORK_DIR}/pixiu.rendered.yaml"
PIXIU_LOG="${WORK_DIR}/pixiu.log"

BASELINE_ROUNDS="${BASELINE_ROUNDS:-12}"
CACHED_ROUNDS="${CACHED_ROUNDS:-12}"
BASELINE_FILE="${WORK_DIR}/baseline.times"
CACHED_FILE="${WORK_DIR}/cached.times"

PIXIU_PID=""

cleanup() {
  set +e
  local exit_code=$?
  if [[ -n "${PIXIU_PID}" ]] && kill -0 "${PIXIU_PID}" >/dev/null 2>&1; then
    kill "${PIXIU_PID}" >/dev/null 2>&1
    wait "${PIXIU_PID}" >/dev/null 2>&1
  fi
  if [[ "${KEEP_WORK_DIR}" != "1" && "${exit_code}" -eq 0 && -n "${WORK_DIR:-}" && -d "${WORK_DIR}" ]]; then
    rm -rf "${WORK_DIR}"
  fi
  return "${exit_code}"
}
trap cleanup EXIT INT TERM

wait_for_pixiu() {
  local body
  body="$(jq -nc --arg model "${MODEL_NAME}" '{model:$model,prompt:"kvcache smoke"}')"
  for _ in $(seq 1 100); do
    local code
    code="$(curl -s -o /dev/null -w '%{http_code}' \
      -H 'Content-Type: application/json' \
      -X POST "${PIXIU_URL}/v1/chat/completions" \
      -d "${body}" || true)"
    if [[ "${code}" == "200" || "${code}" == "4"* || "${code}" == "5"* ]]; then
      return 0
    fi
    sleep 0.3
  done
  return 1
}

start_pixiu() {
  if command -v pixiu >/dev/null 2>&1; then
    pixiu gateway start -c "${RENDERED_CONFIG}" >"${PIXIU_LOG}" 2>&1 &
    PIXIU_PID="$!"
  elif [[ -d "${PIXIU_SOURCE}/cmd/pixiu" ]]; then
    (
      cd "${PIXIU_SOURCE}"
      env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" go run ./cmd/pixiu/*.go gateway start -c "${RENDERED_CONFIG}"
    ) >"${PIXIU_LOG}" 2>&1 &
    PIXIU_PID="$!"
  else
    echo "cannot find pixiu binary or source. set PIXIU_SOURCE or install pixiu in PATH"
    exit 1
  fi

  if ! wait_for_pixiu; then
    echo "pixiu not ready, log: ${PIXIU_LOG}"
    exit 1
  fi
}

render_config() {
  envsubst <"${SCRIPT_DIR}/pixiu/conf.yaml" >"${RENDERED_CONFIG}"
}

p95() {
  local file="$1"
  awk 'NF {print $1}' "${file}" | sort -n | awk '
    {
      arr[NR] = $1
    }
    END {
      if (NR == 0) {
        print "0"
        exit
      }
      idx = int((NR * 95 + 99) / 100)
      if (idx < 1) idx = 1
      if (idx > NR) idx = NR
      print arr[idx]
    }'
}

avg() {
  local file="$1"
  awk '{sum += $1; n += 1} END {if (n == 0) {print "0"} else {printf "%.6f", sum / n}}' "${file}"
}

run_load() {
  local mode="$1"
  local rounds="$2"
  local output_file="$3"
  local fixed_prompt="kvcache route preference probe"

  : >"${output_file}"
  for i in $(seq 1 "${rounds}"); do
    local prompt
    if [[ "${mode}" == "baseline" ]]; then
      prompt="${fixed_prompt} baseline-${i}-$(date +%s%N)"
    else
      prompt="${fixed_prompt}"
    fi

    local body
    body="$(jq -nc --arg model "${MODEL_NAME}" --arg prompt "${prompt}" '{model:$model,messages:[{role:"user",content:$prompt}]}')"

    local result
    local response_file="${WORK_DIR}/${mode}-${i}.response.out"
    result="$(curl -sS -o "${response_file}" -w '%{http_code} %{time_total}' \
      -H 'Content-Type: application/json' \
      -X POST "${PIXIU_URL}/v1/chat/completions" \
      -d "${body}")"

    local status
    status="${result%% *}"
    local timing
    timing="${result##* }"

    if [[ "${status}" != "200" ]]; then
      echo "request failed in ${mode} mode: status=${status}"
      cat "${response_file}"
      exit 1
    fi

    echo "${timing}" >>"${output_file}"
  done
}

lookup_probe() {
  local tokenize_body
  tokenize_body="$(jq -nc --arg model "${MODEL_NAME}" '{model:$model,prompt:"kvcache lookup probe"}')"

  local tokenize_resp_file="${WORK_DIR}/lookup-probe-tokenize.response.json"
  local tokenize_status
  tokenize_status="$(curl -sS -o "${tokenize_resp_file}" -w '%{http_code}' \
    -H 'Content-Type: application/json' \
    -X POST "${VLLM_ENDPOINT}/tokenize" \
    -d "${tokenize_body}")"
  if [[ "${tokenize_status}" != "200" ]]; then
    echo "lookup_probe_error: tokenize returned HTTP ${tokenize_status}"
    cat "${tokenize_resp_file}"
    exit 1
  fi

  local tokens_json
  tokens_json="$(jq -c 'if type == "object" then (.tokens // []) else [] end' "${tokenize_resp_file}" 2>/dev/null || echo '[]')"
  if [[ "${tokens_json}" == "[]" ]]; then
    echo "lookup_probe_error: tokenize response did not contain usable tokens"
    cat "${tokenize_resp_file}"
    exit 1
  fi

  local lookup_body
  lookup_body="$(jq -nc --argjson t "${tokens_json}" '{tokens:$t}')"
  local lookup_resp_file="${WORK_DIR}/lookup-probe-lookup.response.json"
  local lookup_status
  lookup_status="$(curl -sS -o "${lookup_resp_file}" -w '%{http_code}' \
    -H 'Content-Type: application/json' \
    -X POST "${LMCACHE_ENDPOINT}/lookup" \
    -d "${lookup_body}")"
  if [[ "${lookup_status}" != "200" ]]; then
    echo "lookup_probe_error: lookup returned HTTP ${lookup_status}"
    cat "${lookup_resp_file}"
    exit 1
  fi

  local preferred
  preferred="$(jq -r 'if type == "object" then ((.layout_info // {}) | to_entries | max_by(.value["1"]) | .key) else empty end // empty' "${lookup_resp_file}" 2>/dev/null || true)"
  if [[ -z "${preferred}" ]]; then
    echo "lookup_probe_error: cannot parse preferred endpoint from lookup response"
    cat "${lookup_resp_file}"
    exit 1
  fi

  echo "${preferred}"
}

calc_preferred_hit_rate() {
  local preferred_id="$1"
  local preferred_addr=""

  if [[ "${preferred_id}" == "${ENGINE_ENDPOINT_A_ID}" ]]; then
    preferred_addr="${ENGINE_ENDPOINT_A_HOST}:${ENGINE_ENDPOINT_A_PORT}"
  elif [[ "${preferred_id}" == "${ENGINE_ENDPOINT_B_ID}" ]]; then
    preferred_addr="${ENGINE_ENDPOINT_B_HOST}:${ENGINE_ENDPOINT_B_PORT}"
  else
    echo "0 0 ${preferred_id}"
    return
  fi

  local metrics
  metrics="$(curl -sS "${PROM_URL}" || true)"

  local preferred_hits
  preferred_hits="$(awk -v addr="${preferred_addr}" '
    /^pixiu_llm_upstream_requests_total/ && index($0, "endpoint_address=\"" addr "\"") > 0 {sum += $NF}
    END {printf "%.0f", sum}
  ' <<<"${metrics}")"

  local total_hits
  total_hits="$(awk -v addr_a="${ENGINE_ENDPOINT_A_HOST}:${ENGINE_ENDPOINT_A_PORT}" -v addr_b="${ENGINE_ENDPOINT_B_HOST}:${ENGINE_ENDPOINT_B_PORT}" '
    /^pixiu_llm_upstream_requests_total/ && (index($0, "endpoint_address=\"" addr_a "\"") > 0 || index($0, "endpoint_address=\"" addr_b "\"") > 0) {sum += $NF}
    END {printf "%.0f", sum}
  ' <<<"${metrics}")"

  echo "${preferred_hits} ${total_hits} ${preferred_addr}"
}

echo "[1/5] rendering pixiu config"
render_config
echo "rendered config: ${RENDERED_CONFIG}"

if [[ "${START_PIXIU}" == "1" ]]; then
  echo "[2/5] starting pixiu"
  start_pixiu
else
  echo "[2/5] using existing pixiu at ${PIXIU_URL}"
fi

echo "[3/5] probing lookup/preferred endpoint"
preferred_id="$(lookup_probe)"
echo "lookup preferred endpoint id: ${preferred_id}"

echo "[4/5] running latency workloads (baseline vs cached)"
run_load baseline "${BASELINE_ROUNDS}" "${BASELINE_FILE}"
run_load cached "${CACHED_ROUNDS}" "${CACHED_FILE}"

baseline_p95="$(p95 "${BASELINE_FILE}")"
cached_p95="$(p95 "${CACHED_FILE}")"
baseline_avg="$(avg "${BASELINE_FILE}")"
cached_avg="$(avg "${CACHED_FILE}")"

improvement_pct="$(awk -v b="${baseline_p95}" -v c="${cached_p95}" 'BEGIN { if (b <= 0) {print "0.00"} else {printf "%.2f", ((b-c)/b)*100 } }')"

echo "[5/5] computing preferred endpoint hit rate"
read -r preferred_hits total_hits preferred_addr < <(calc_preferred_hit_rate "${preferred_id}")
hit_rate="$(awk -v p="${preferred_hits}" -v t="${total_hits}" 'BEGIN { if (t <= 0) {print "0.00"} else {printf "%.2f", (p/t)*100 } }')"

cat <<REPORT

=== KVCache Real-Engine Verification Report ===
Pixiu URL: ${PIXIU_URL}
Prometheus URL: ${PROM_URL}
Preferred endpoint id (from lookup): ${preferred_id}
Preferred endpoint addr (mapped): ${preferred_addr}

lookup success: PASS (preferred endpoint parsed)
preferred endpoint hit: ${preferred_hits}/${total_hits} (${hit_rate}%)

latency baseline avg: ${baseline_avg}s
latency cached   avg: ${cached_avg}s
latency baseline p95: ${baseline_p95}s
latency cached   p95: ${cached_p95}s
p95 delta (baseline-cached)/baseline: ${improvement_pct}%

rendered config: ${RENDERED_CONFIG}
REPORT

if [[ "${total_hits}" == "0" ]]; then
  echo "WARN: no llm upstream metrics observed. verify prometheus endpoint and traffic labels."
fi

if [[ "${START_PIXIU}" == "1" ]]; then
  echo "pixiu log: ${PIXIU_LOG}"
fi
