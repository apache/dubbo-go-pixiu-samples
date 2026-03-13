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

type engineStats struct {
	mu sync.Mutex

	tokenizeCalls int
	chatCalls     int
}

type llmRequest struct {
	Model string `json:"model"`
}

type llmResponse struct {
	ID       string         `json:"id"`
	Object   string         `json:"object"`
	Model    string         `json:"model"`
	ServedBy string         `json:"served_by"`
	Choices  []llmChoice    `json:"choices"`
	Usage    map[string]int `json:"usage"`
}

type llmChoice struct {
	Index   int        `json:"index"`
	Message llmMessage `json:"message"`
}

type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type tokenizeResponse struct {
	Count  int   `json:"count"`
	Tokens []int `json:"tokens"`
	MaxLen int   `json:"max_model_len"`
}

var eventCounter uint64

func main() {
	addr := envOrDefault("LLM_A_ADDR", ":18091")
	engineID := envOrDefault("LLM_A_ID", "mock-llm-a")
	stats := &engineStats{}

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

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	log.Printf("[mock-engine-a] listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[mock-engine-a] server failed: %v", err)
	}
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func nextEventID(prefix string) string {
	n := atomic.AddUint64(&eventCounter, 1)
	return prefix + "-" + strconv.FormatUint(n, 10)
}

func envOrDefault(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}
