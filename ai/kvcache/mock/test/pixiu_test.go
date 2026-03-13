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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	defaultPixiuURL      = "http://127.0.0.1:18888"
	defaultControllerURL = "http://127.0.0.1:18081"
	defaultEngineAURL    = "http://127.0.0.1:18091"
	defaultEngineBURL    = "http://127.0.0.1:18092"
)

var testHTTPClient = &http.Client{Timeout: 3 * time.Second}

func getEnvOrDefault(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func checkServiceAvailable(url string) bool {
	resp, err := testHTTPClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func checkPixiuAvailable(url string) bool {
	payload := map[string]any{
		"model":  "mock-model",
		"prompt": "kvcache availability probe",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	req, err := http.NewRequest(http.MethodPost, url+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 600
}

func postJSON(t *testing.T, url string, payload any) map[string]any {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("request status %d: %s", resp.StatusCode, string(body))
	}

	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal response failed: %v body=%s", err, string(body))
	}
	return out
}

func getJSON(t *testing.T, url string) map[string]any {
	t.Helper()
	resp, err := testHTTPClient.Get(url)
	if err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("get status %d: %s", resp.StatusCode, string(body))
	}

	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal response failed: %v body=%s", err, string(body))
	}
	return out
}

func toInt(t *testing.T, value any, key string) int {
	t.Helper()
	num, ok := value.(float64)
	if !ok {
		t.Fatalf("%s is not numeric: %#v", key, value)
	}
	return int(num)
}

func TestKVCacheMockRoutingFlow(t *testing.T) {
	pixiuURL := getEnvOrDefault("PIXIU_URL", defaultPixiuURL)
	controllerURL := getEnvOrDefault("CONTROLLER_URL", defaultControllerURL)
	engineAURL := getEnvOrDefault("ENGINE_A_URL", defaultEngineAURL)
	engineBURL := getEnvOrDefault("ENGINE_B_URL", defaultEngineBURL)

	if !checkServiceAvailable(controllerURL+"/health") ||
		!checkServiceAvailable(engineAURL+"/health") ||
		!checkServiceAvailable(engineBURL+"/health") ||
		!checkPixiuAvailable(pixiuURL) {
		t.Skip("required services are unavailable; start mock servers and pixiu first")
	}

	postJSON(t, controllerURL+"/reset", map[string]any{})
	postJSON(t, engineAURL+"/reset", map[string]any{})
	postJSON(t, engineBURL+"/reset", map[string]any{})

	payload := map[string]any{
		"model": "mock-model",
		"messages": []map[string]any{{
			"role":    "user",
			"content": "please route same prompt for kvcache test",
		}},
	}

	postJSON(t, pixiuURL+"/v1/chat/completions", payload)
	time.Sleep(600 * time.Millisecond)
	second := postJSON(t, pixiuURL+"/v1/chat/completions", payload)

	servedBy, _ := second["served_by"].(string)
	if servedBy != "mock-llm-b" {
		t.Fatalf("expected served_by mock-llm-b, got %q", servedBy)
	}

	time.Sleep(1 * time.Second)
	controllerStats := getJSON(t, controllerURL+"/stats")
	engineAStats := getJSON(t, engineAURL+"/stats")
	engineBStats := getJSON(t, engineBURL+"/stats")

	if got := toInt(t, engineAStats["tokenize_calls"], "tokenize_calls"); got < 1 {
		t.Fatalf("expected tokenize_calls >= 1, got %d", got)
	}
	if got := toInt(t, controllerStats["lookup_calls"], "lookup_calls"); got < 2 {
		t.Fatalf("expected lookup_calls >= 2, got %d", got)
	}
	if got := toInt(t, controllerStats["pin_calls"], "pin_calls"); got < 1 {
		t.Fatalf("expected pin_calls >= 1, got %d", got)
	}
	if got := toInt(t, engineBStats["chat_calls"], "chat_calls"); got < 1 {
		t.Fatalf("expected engine-b chat_calls >= 1, got %d", got)
	}
}
