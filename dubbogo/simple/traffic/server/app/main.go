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
)

func main() {
	go startServer(":1315", "v1")
	go startServer(":1316", "v2")
	go startServer(":1317", "v3")

	log.Println("All traffic servers started")
	select {}
}

func startServer(addr, label string) {
	mux := http.NewServeMux()
	routers := []string{"/user", "/user/pixiu", "/prefix", "/health"}
	for _, router := range routers {
		msg := router[strings.LastIndex(router, "/")+1:]
		mux.HandleFunc(router, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"server": "%s","message":"%s","status":200}`, label, msg)
		})
	}
	log.Printf("Starting %s server on %s...", label, addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
