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

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserHandlerGetUserById(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost,
		"/com.dubbogo.pixiu.TripleUserService/GetUserById",
		strings.NewReader(`["0001"]`))
	w := httptest.NewRecorder()
	user(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `"id":"0001"`) || !strings.Contains(body, `"name":"tc"`) {
		t.Fatalf("expected id=0001 name=tc, got %s", body)
	}
}

func TestUserHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost,
		"/com.dubbogo.pixiu.TripleUserService/GetUserById",
		strings.NewReader(`["9999"]`))
	w := httptest.NewRecorder()
	user(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "user not found") {
		t.Fatalf("expected user not found body, got %s", w.Body.String())
	}
}

func TestUserHandlerBadSchemaString(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost,
		"/com.dubbogo.pixiu.TripleUserService/GetUserById",
		strings.NewReader(`"0001"`))
	w := httptest.NewRecorder()
	user(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for plain-string body, got %d", w.Code)
	}
}

func TestUserHandlerEmptyArray(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost,
		"/com.dubbogo.pixiu.TripleUserService/GetUserById",
		strings.NewReader(`[]`))
	w := httptest.NewRecorder()
	user(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty array, got %d body=%s", w.Code, w.Body.String())
	}
	b, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(b), "missing id argument") {
		t.Fatalf("expected missing id argument, got %s", string(b))
	}
}

func TestUserHandlerCacheNotPolluted(t *testing.T) {
	before := len(cache.cacheMap)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost,
			"/com.dubbogo.pixiu.TripleUserService/GetUserById",
			strings.NewReader(`["miss-`+string(rune('A'+i))+`"]`))
		w := httptest.NewRecorder()
		user(w, req)
	}
	after := len(cache.cacheMap)
	if before != after {
		t.Fatalf("cache size changed: before=%d after=%d", before, after)
	}
}

func TestUserHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/com.dubbogo.pixiu.TripleUserService/GetUserById", nil)
	w := httptest.NewRecorder()
	user(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
