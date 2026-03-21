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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type controllerStats struct {
	mu sync.Mutex

	lookupCalls   int
	lookupSuccess int
	lookupFailure int
	pinCalls      int
	compressCalls int
	evictCalls    int
}

type engineAStats struct {
	mu sync.Mutex

	tokenizeCalls int
	chatCalls     int
}

type engineBStats struct {
	mu sync.Mutex

	chatCalls int
}

type lookupRequest struct {
	Tokens []int `json:"tokens"`
}

type tokensRequest struct {
	Tokens []int `json:"tokens"`
}

type eventResp struct {
	EventID   string `json:"event_id"`
	NumTokens int    `json:"num_tokens"`
}

type llmRequest struct {
	Model string `json:"model"`
}

type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llmChoice struct {
	Index   int        `json:"index"`
	Message llmMessage `json:"message"`
}

type llmResponse struct {
	ID       string         `json:"id"`
	Object   string         `json:"object"`
	Model    string         `json:"model"`
	ServedBy string         `json:"served_by"`
	Choices  []llmChoice    `json:"choices"`
	Usage    map[string]int `json:"usage"`
}

type tokenizeResponse struct {
	Count  int   `json:"count"`
	Tokens []int `json:"tokens"`
	MaxLen int   `json:"max_model_len"`
}

var globalEventCounter uint64

func main() {
	controllerAddr := envOrDefault("LMCACHE_ADDR", ":18081")
	engineAAddr := envOrDefault("LLM_A_ADDR", ":18091")
	engineBAddr := envOrDefault("LLM_B_ADDR", ":18092")
	preferredEndpoint := envOrDefault("PREFERRED_ENDPOINT_ID", "mock-llm-b")
	engineAID := envOrDefault("LLM_A_ID", "mock-llm-a")
	engineBID := envOrDefault("LLM_B_ID", "mock-llm-b")
	responseDelay := envDurationMSOrDefault("MOCK_LLM_RESPONSE_DELAY_MS", 150)

	controller := buildControllerMux(preferredEndpoint)
	engineA := buildEngineAMux(engineAID, responseDelay)
	engineB := buildEngineBMux(engineBID, responseDelay)

	errCh := make(chan error, 3)
	go serve("mock-controller", controllerAddr, controller, errCh)
	go serve("mock-engine-a", engineAAddr, engineA, errCh)
	go serve("mock-engine-b", engineBAddr, engineB, errCh)

	err := <-errCh
	log.Fatalf("kvcache mock app exited: %v", err)
}

func serve(name string, addr string, handler http.Handler, errCh chan<- error) {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Printf("[%s] listening on %s", name, addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errCh <- fmt.Errorf("%s failed: %w", name, err)
	}
}

func buildControllerMux(preferred string) http.Handler {
	stats := &controllerStats{}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "component": "mock-controller"})
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, _ *http.Request) {
		stats.mu.Lock()
		defer stats.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{
			"lookup_calls":         stats.lookupCalls,
			"lookup_success":       stats.lookupSuccess,
			"lookup_failure":       stats.lookupFailure,
			"pin_calls":            stats.pinCalls,
			"compress_calls":       stats.compressCalls,
			"evict_calls":          stats.evictCalls,
			"preferred_endpoint":   preferred,
			"timestamp_unix_milli": time.Now().UnixMilli(),
		})
	})
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		stats.mu.Lock()
		stats.lookupCalls = 0
		stats.lookupSuccess = 0
		stats.lookupFailure = 0
		stats.pinCalls = 0
		stats.compressCalls = 0
		stats.evictCalls = 0
		stats.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("/lookup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var req lookupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			stats.mu.Lock()
			stats.lookupFailure++
			stats.mu.Unlock()
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		tokenCount := len(req.Tokens)
		if tokenCount == 0 {
			tokenCount = 1
		}
		layout := map[string]map[string]any{
			"mock-llm-a": {"0": "ram-a", "1": max(tokenCount/2, 1)},
			"mock-llm-b": {"0": "ram-b", "1": tokenCount + 3},
		}
		if preferred == "mock-llm-a" {
			layout["mock-llm-a"]["1"] = tokenCount + 3
			layout["mock-llm-b"]["1"] = max(tokenCount/2, 1)
		}

		stats.mu.Lock()
		stats.lookupCalls++
		stats.lookupSuccess++
		stats.mu.Unlock()

		writeJSON(w, http.StatusOK, map[string]any{
			"event_id":    nextEventID("lookup"),
			"layout_info": layout,
		})
	})
	mux.HandleFunc("/pin", func(w http.ResponseWriter, r *http.Request) {
		handleTokenEvent(stats, w, r, "pin")
	})
	mux.HandleFunc("/compress", func(w http.ResponseWriter, r *http.Request) {
		handleTokenEvent(stats, w, r, "compress")
	})
	mux.HandleFunc("/evict", func(w http.ResponseWriter, r *http.Request) {
		handleTokenEvent(stats, w, r, "evict")
	})

	return mux
}

func handleTokenEvent(stats *controllerStats, w http.ResponseWriter, r *http.Request, op string) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req tokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	stats.mu.Lock()
	switch op {
	case "pin":
		stats.pinCalls++
	case "compress":
		stats.compressCalls++
	case "evict":
		stats.evictCalls++
	}
	stats.mu.Unlock()
	writeJSON(w, http.StatusOK, eventResp{EventID: nextEventID(op), NumTokens: len(req.Tokens)})
}

func buildEngineAMux(engineID string, responseDelay time.Duration) http.Handler {
	stats := &engineAStats{}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "engine": engineID, "tokenize_enabled": true})
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, _ *http.Request) {
		stats.mu.Lock()
		defer stats.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{
			"tokenize_calls":       stats.tokenizeCalls,
			"chat_calls":           stats.chatCalls,
			"engine_id":            engineID,
			"timestamp_unix_milli": time.Now().UnixMilli(),
		})
	})
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		stats.mu.Lock()
		stats.tokenizeCalls = 0
		stats.chatCalls = 0
		stats.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("/tokenize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		prompt := extractPrompt(r)
		tokens := tokenizePrompt(prompt)

		stats.mu.Lock()
		stats.tokenizeCalls++
		stats.mu.Unlock()

		writeJSON(w, http.StatusOK, tokenizeResponse{Count: len(tokens), Tokens: tokens, MaxLen: 8192})
	})
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var req llmRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Model == "" {
			req.Model = "mock-model"
		}
		stats.mu.Lock()
		stats.chatCalls++
		stats.mu.Unlock()
		if responseDelay > 0 {
			time.Sleep(responseDelay)
		}

		resp := llmResponse{
			ID:       nextEventID("chatcmpl"),
			Object:   "chat.completion",
			Model:    req.Model,
			ServedBy: engineID,
			Choices: []llmChoice{{
				Index: 0,
				Message: llmMessage{
					Role:    "assistant",
					Content: fmt.Sprintf("mock response from %s", engineID),
				},
			}},
			Usage: map[string]int{"prompt_tokens": 8, "completion_tokens": 8, "total_tokens": 16},
		}
		writeJSON(w, http.StatusOK, resp)
	})

	return mux
}

func buildEngineBMux(engineID string, responseDelay time.Duration) http.Handler {
	stats := &engineBStats{}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "engine": engineID, "tokenize_enabled": false})
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, _ *http.Request) {
		stats.mu.Lock()
		defer stats.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{
			"chat_calls":           stats.chatCalls,
			"engine_id":            engineID,
			"timestamp_unix_milli": time.Now().UnixMilli(),
		})
	})
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		stats.mu.Lock()
		stats.chatCalls = 0
		stats.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("/tokenize", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "tokenize not available on this instance"})
	})
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var req llmRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Model == "" {
			req.Model = "mock-model"
		}

		stats.mu.Lock()
		stats.chatCalls++
		stats.mu.Unlock()
		if responseDelay > 0 {
			time.Sleep(responseDelay)
		}

		resp := llmResponse{
			ID:       nextEventID("chatcmpl"),
			Object:   "chat.completion",
			Model:    req.Model,
			ServedBy: engineID,
			Choices: []llmChoice{{
				Index: 0,
				Message: llmMessage{
					Role:    "assistant",
					Content: fmt.Sprintf("mock response from %s", engineID),
				},
			}},
			Usage: map[string]int{"prompt_tokens": 8, "completion_tokens": 8, "total_tokens": 16},
		}
		writeJSON(w, http.StatusOK, resp)
	})

	return mux
}

func extractPrompt(r *http.Request) string {
	if r == nil || r.Body == nil {
		return ""
	}
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return ""
	}
	if prompt, ok := payload["prompt"]; ok {
		switch v := prompt.(type) {
		case string:
			return strings.TrimSpace(v)
		case []any:
			parts := make([]string, 0, len(v))
			for _, item := range v {
				if str, ok := item.(string); ok {
					parts = append(parts, str)
				}
			}
			return strings.Join(parts, "\n")
		}
	}
	messages, ok := payload["messages"].([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(messages))
	for _, item := range messages {
		msgMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if content, ok := msgMap["content"].(string); ok {
			parts = append(parts, content)
		}
	}
	return strings.Join(parts, "\n")
}

func tokenizePrompt(prompt string) []int {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return []int{0}
	}
	words := strings.Fields(prompt)
	if len(words) == 0 {
		return []int{0}
	}
	tokens := make([]int, 0, len(words))
	for idx, word := range words {
		sum := 0
		for _, r := range word {
			sum += int(r)
		}
		tokens = append(tokens, (sum%997)+idx+1)
	}
	return tokens
}

func nextEventID(prefix string) string {
	n := atomic.AddUint64(&globalEventCounter, 1)
	return prefix + "-" + strconv.FormatUint(n, 10)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func envOrDefault(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}

func envDurationMSOrDefault(key string, fallbackMS int) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return time.Duration(fallbackMS) * time.Millisecond
	}
	ms, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil || ms < 0 {
		return time.Duration(fallbackMS) * time.Millisecond
	}
	return time.Duration(ms) * time.Millisecond
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
