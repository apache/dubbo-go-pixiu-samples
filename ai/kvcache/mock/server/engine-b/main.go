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

	chatCalls int
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

var eventCounter uint64

func main() {
	addr := envOrDefault("LLM_B_ADDR", ":18092")
	engineID := envOrDefault("LLM_B_ID", "mock-llm-b")
	stats := &engineStats{}

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
	log.Printf("[mock-engine-b] listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[mock-engine-b] server failed: %v", err)
	}
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
