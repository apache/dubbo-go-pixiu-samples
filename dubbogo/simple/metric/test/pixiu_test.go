/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metric

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestMetric(t *testing.T) {
	// Wait for services to be ready
	time.Sleep(2 * time.Second)

	// Test user API multiple times to generate metrics
	for i := 0; i < 5; i++ {
		resp := verify(t, "http://localhost:8888/api/v1/test-dubbo/user/tc?age=18", http.StatusOK)
		assert.True(t, strings.Contains(resp, "tc"))
		time.Sleep(100 * time.Millisecond)
	}

	// Wait a bit for metrics to be collected
	time.Sleep(1 * time.Second)

	// Test metrics endpoint
	metricsResp := verify(t, "http://localhost:9091/", http.StatusOK)

	// Verify that metrics are being collected
	// The new metric filter should expose standard Prometheus metrics
	assert.True(t, strings.Contains(metricsResp, "pixiu") || strings.Contains(metricsResp, "http"), 
		"Metrics should contain pixiu or http metrics")

	t.Logf("Metrics response preview (first 500 chars): %s", 
		metricsResp[:min(500, len(metricsResp))])
}

func verify(t *testing.T, url string, status int) string {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, status, resp.StatusCode)
	assert.NotNil(t, resp)
	s, _ := io.ReadAll(resp.Body)
	return string(s)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


