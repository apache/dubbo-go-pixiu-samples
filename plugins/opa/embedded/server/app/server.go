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
	"strings"
)

type Resp struct {
	Message string `json:"message"`
	Result  string `json:"result"`
}

func main() {
	routers := []string{"/UserService", "/OtherService"}

	for _, rt := range routers {
		route := rt
		msg := route[strings.LastIndex(route, "/")+1:]

		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[backend] %s %s Headers=%v", r.Method, r.URL.Path, r.Header)

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Resp{
				Message: msg,
				Result:  "pass",
			})
		})
	}

	log.Println("Starting sample backend on :1314 ...")
	log.Fatal(http.ListenAndServe(":1314", nil))
}
