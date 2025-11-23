#!/bin/bash
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

echo "Testing Pixiu Metric Sample..."
echo ""

echo "1. Testing user API..."
curl -X GET "http://localhost:8888/api/v1/test-dubbo/user/tc?age=18"
echo ""
echo ""

echo "2. Sending multiple requests to generate metrics..."
for i in {1..10}; do
  echo "Request $i..."
  curl -X GET "http://localhost:8888/api/v1/test-dubbo/user/user$i?age=$((20+i))"
  echo ""
  sleep 0.5
done
echo ""

echo "3. Fetching Prometheus metrics..."
echo "Metrics endpoint: http://localhost:9091/"
echo ""
curl http://localhost:9091/
echo ""
echo ""

echo "Done! You can also view metrics in:"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000 (admin/admin)"


