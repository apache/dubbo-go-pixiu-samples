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
	"strings"
	"sync"
)

import (
	traffic "github.com/dubbo-go-pixiu/samples/dubbogo/simple/traffic"
)

type trafficServer struct {
	addr  string
	label string
}

func main() {
	config, err := traffic.LoadConfig()
	if err != nil {
		log.Fatalf("invalid traffic configuration: %v", err)
	}
	servers := []trafficServer{
		{addr: config.V1.ListenAddr, label: "v1"},
		{addr: config.V2.ListenAddr, label: "v2"},
		{addr: config.V3.ListenAddr, label: "v3"},
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

func startServer(addr, label string) error {
	mux := http.NewServeMux()
	routers := []string{"/user", "/user/pixiu", "/prefix", "/health"}
	for _, router := range routers {
		msg := router[strings.LastIndex(router, "/")+1:]
		mux.HandleFunc(router, func(w http.ResponseWriter, r *http.Request) {
			if _, err := fmt.Fprintf(w, `{"server": "%s","message":"%s","status":200}`, label, msg); err != nil {
				log.Printf("write %s response: %v", label, err)
			}
		})
	}
	log.Printf("Starting %s server on %s...", label, addr)
	return (&http.Server{Addr: addr, Handler: mux}).ListenAndServe()
}
