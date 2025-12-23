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

package test

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

const embeddedGatewayBaseURL = "http://localhost:8888"

func doEmbeddedRequest(t *testing.T, path string, headerVal *string) (int, string) {
	t.Helper()

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", embeddedGatewayBaseURL+path, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	if headerVal != nil {
		req.Header.Set("Test_header", *headerVal)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Surface clearer guidance when gateway is not running
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			t.Fatalf("request timeout: ensure embedded Pixiu gateway and backend are running: %v", err)
		}
		t.Fatalf("request failed: ensure embedded Pixiu gateway and backend are running: %v", err)
	}
	if resp == nil {
		t.Fatalf("response is nil")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return resp.StatusCode, string(body)
}

func TestEmbeddedUserServiceAllow(t *testing.T) {
	header := "1"
	status, body := doEmbeddedRequest(t, "/UserService", &header)

	assert.Equal(t, http.StatusOK, status)
	assert.True(t, strings.Contains(body, "pass"))
	assert.True(t, strings.Contains(body, "UserService"))
}

func TestEmbeddedUserServiceDeny(t *testing.T) {
	status, body := doEmbeddedRequest(t, "/UserService", nil)

	assert.Equal(t, http.StatusForbidden, status)
	assert.False(t, strings.Contains(body, "pass"))
	assert.False(t, strings.Contains(body, "UserService"))
}

func TestEmbeddedOtherServiceDeny(t *testing.T) {
	header := "1"
	status, body := doEmbeddedRequest(t, "/OtherService", &header)

	assert.Equal(t, http.StatusForbidden, status)
	assert.False(t, strings.Contains(body, "pass"))
	assert.False(t, strings.Contains(body, "OtherService"))
}
