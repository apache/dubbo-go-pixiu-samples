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

var eventCounter uint64

func main() {
	addr := envOrDefault("LMCACHE_ADDR", ":18081")
	preferred := envOrDefault("PREFERRED_ENDPOINT_ID", "mock-llm-b")
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

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	log.Printf("[mock-controller] listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[mock-controller] server failed: %v", err)
	}
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

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
