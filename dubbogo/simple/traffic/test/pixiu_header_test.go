//go:build manual

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
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestHeaderGET1(t *testing.T) {
	assertHeaderResponse(t, http.StatusOK, "v1", func(req *http.Request) {
		req.Header.Set("X-A", "t1")
		req.Header.Set("X-B", "t4")
	})
}

func TestHeaderGET2(t *testing.T) {
	assertHeaderResponse(t, http.StatusOK, "v2", func(req *http.Request) {
		req.Header.Set("X-B", "t4")
		req.Header.Set("X-C", "t1")
	})
}

func TestHeaderGET3(t *testing.T) {
	assertHeaderResponse(t, http.StatusOK, "v3", func(req *http.Request) {
		req.Header.Set("REG", "tt")
	})
}

func TestHeaderGETMissingB(t *testing.T) {
	assertHeaderRouteNotFound(t, func(req *http.Request) {
		req.Header.Set("X-A", "t1")
	})
}

func TestHeaderGETWithoutHeaders(t *testing.T) {
	assertHeaderRouteNotFound(t, nil)
}

func TestHeaderGETWrongValues(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
	}{
		{name: "both values wrong", headers: map[string]string{"X-A": "wrong", "X-B": "wrong"}},
		{name: "A wrong", headers: map[string]string{"X-A": "wrong", "X-B": "t4"}},
		{name: "B wrong", headers: map[string]string{"X-A": "t1", "X-B": "wrong"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertHeaderRouteNotFound(t, func(req *http.Request) {
				for key, value := range test.headers {
					req.Header.Set(key, value)
				}
			})
		})
	}
}

func assertHeaderResponse(t *testing.T, status int, backend string, setHeaders func(*http.Request)) {
	t.Helper()
	resp, body := headerRequest(t, setHeaders)
	if resp.StatusCode != status {
		t.Fatalf("expected HTTP %d, got %d: %s", status, resp.StatusCode, body)
	}
	if backend != "" && !containsServer(body, backend) {
		t.Fatalf("expected backend %s response, got %s", backend, body)
	}
}

func assertHeaderRouteNotFound(t *testing.T, setHeaders func(*http.Request)) {
	t.Helper()
	resp, body := headerRequest(t, setHeaders)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected HTTP 404 for unmatched route, got %d: %s", resp.StatusCode, body)
	}
	var response struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode route-not-found response %q: %v", body, err)
	}
	if response.Status != http.StatusNotFound || response.Message != "Route not found" {
		t.Fatalf("expected Pixiu route-not-found response, got %s", body)
	}
}

func headerRequest(t *testing.T, setHeaders func(*http.Request)) (*http.Response, []byte) {
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return resp, body
}

func containsServer(body []byte, backend string) bool {
	var response struct {
		Server string `json:"server"`
	}
	return json.Unmarshal(body, &response) == nil && response.Server == backend
}
