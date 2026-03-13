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

PIXIU_URL="${PIXIU_URL:-http://127.0.0.1:18888}"
CONTROLLER_URL="${CONTROLLER_URL:-http://127.0.0.1:18081}"
ENGINE_A_URL="${ENGINE_A_URL:-http://127.0.0.1:18091}"
ENGINE_B_URL="${ENGINE_B_URL:-http://127.0.0.1:18092}"

export NO_PROXY="${NO_PROXY:-127.0.0.1,localhost}"
export no_proxy="${no_proxy:-127.0.0.1,localhost}"

body='{"model":"mock-model","messages":[{"role":"user","content":"route this same prompt"}]}'

echo "reset stats"
curl -fsS -X POST "${CONTROLLER_URL}/reset" >/dev/null
curl -fsS -X POST "${ENGINE_A_URL}/reset" >/dev/null
curl -fsS -X POST "${ENGINE_B_URL}/reset" >/dev/null

echo "request #1"
curl -fsS -H 'Content-Type: application/json' -X POST "${PIXIU_URL}/v1/chat/completions" -d "${body}"
echo ""

sleep 0.4
echo "request #2"
curl -fsS -H 'Content-Type: application/json' -X POST "${PIXIU_URL}/v1/chat/completions" -d "${body}"
echo ""

sleep 0.8
echo "controller stats"
curl -fsS "${CONTROLLER_URL}/stats"
echo ""

echo "engine-a stats"
curl -fsS "${ENGINE_A_URL}/stats"
echo ""

echo "engine-b stats"
curl -fsS "${ENGINE_B_URL}/stats"
echo ""
