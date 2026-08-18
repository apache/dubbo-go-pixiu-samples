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
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	trafficV1AddrEnv = "TRAFFIC_V1_ADDR"
	trafficV2AddrEnv = "TRAFFIC_V2_ADDR"
	trafficV3AddrEnv = "TRAFFIC_V3_ADDR"

	defaultTrafficV1Addr = ":1315"
	defaultTrafficV2Addr = ":1316"
	defaultTrafficV3Addr = ":1317"
)

type trafficServer struct {
	addr  string
	label string
}

func main() {
	servers := []trafficServer{
		{addr: envOrDefault(trafficV1AddrEnv, defaultTrafficV1Addr), label: "v1"},
		{addr: envOrDefault(trafficV2AddrEnv, defaultTrafficV2Addr), label: "v2"},
		{addr: envOrDefault(trafficV3AddrEnv, defaultTrafficV3Addr), label: "v3"},
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(servers))
	for _, server := range servers {
		wg.Add(1)
		go func(server trafficServer) {
			defer wg.Done()
			errCh <- startServer(server.addr, server.label)
		}(server)
	}

	log.Println("All traffic servers started")
	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			log.Fatalf("traffic server stopped: %v", err)
		}
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func startServer(addr, label string) error {
	mux := http.NewServeMux()
	routers := []string{"/user", "/user/pixiu", "/prefix", "/health"}
	for _, router := range routers {
		msg := router[strings.LastIndex(router, "/")+1:]
		mux.HandleFunc(router, func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, `{"server": "%s","message":"%s","status":200}`, label, msg)
		})
	}
	log.Printf("Starting %s server on %s...", label, addr)
	return http.ListenAndServe(addr, mux)
}
