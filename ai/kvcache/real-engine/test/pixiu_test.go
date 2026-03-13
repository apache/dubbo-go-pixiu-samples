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

const defaultBYOEPixiuURL = "http://127.0.0.1:18889"

var testHTTPClient = &http.Client{Timeout: 3 * time.Second}

func envOrDefault(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func checkPixiuAvailable(url string) bool {
	payload := map[string]any{
		"model":  envOrDefault("MODEL_NAME", "Qwen2.5-3B-Instruct"),
		"prompt": "kvcache byoe availability probe",
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

func TestBYOEEnvironmentAndGatewayAvailability(t *testing.T) {
	if os.Getenv("VLLM_ENDPOINT") == "" || os.Getenv("LMCACHE_ENDPOINT") == "" {
		t.Skip("VLLM_ENDPOINT/LMCACHE_ENDPOINT not set; BYOE environment not configured")
	}

	pixiuURL := envOrDefault("PIXIU_URL", defaultBYOEPixiuURL)
	if !checkPixiuAvailable(pixiuURL) {
		t.Skip("pixiu gateway is unavailable; start pixiu with ai/kvcache/real-engine/pixiu/conf.yaml first")
	}
}

func TestBYOERequestSmoke(t *testing.T) {
	if os.Getenv("VLLM_ENDPOINT") == "" || os.Getenv("LMCACHE_ENDPOINT") == "" {
		t.Skip("VLLM_ENDPOINT/LMCACHE_ENDPOINT not set; BYOE environment not configured")
	}

	pixiuURL := envOrDefault("PIXIU_URL", defaultBYOEPixiuURL)
	if !checkPixiuAvailable(pixiuURL) {
		t.Skip("pixiu gateway is unavailable; start pixiu first")
	}

	payload := map[string]any{
		"model": envOrDefault("MODEL_NAME", "Qwen2.5-3B-Instruct"),
		"messages": []map[string]any{{
			"role":    "user",
			"content": "kvcache byoe smoke test",
		}},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, pixiuURL+"/v1/chat/completions", bytes.NewReader(data))
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
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
}
