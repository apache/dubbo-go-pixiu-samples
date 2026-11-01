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
	"net/http"
	"strings"
	"testing"
)

func TestCanaryGET(t *testing.T) {
	assertTrafficResponse(t, "v1", func(req *http.Request) {
		req.Header.Set("canary-by-header", "v1")
	})
}

func TestCanaryGET1(t *testing.T) {
	assertTrafficResponse(t, "v2", nil)
}

func TestCanaryGET2(t *testing.T) {
	assertTrafficResponse(t, "v3", func(req *http.Request) {
		req.Header.Set("canary-by-header", "v3")
	})
}

func assertTrafficResponse(t *testing.T, backend string, setHeaders func(*http.Request)) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, trafficURL(t, "/user"), nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if setHeaders != nil {
		setHeaders(req)
	}
	resp, err := trafficHTTPClient(t).Do(req)
	if err != nil {
		t.Fatalf("request gateway: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if !strings.Contains(string(body), `"server": "`+backend+`"`) {
		t.Fatalf("expected backend %s response, got %s", backend, body)
	}
}
