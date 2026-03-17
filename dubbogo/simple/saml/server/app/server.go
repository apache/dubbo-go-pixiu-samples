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
)

type appResponse struct {
	Message string `json:"message"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, appResponse{Message: "ok"})
	})

	http.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, appResponse{
			Message: "saml login success",
			Email:   r.Header.Get("X-User-Email"),
			Name:    r.Header.Get("X-User-Name"),
		})
	})

	log.Println("Starting sample server ...")
	log.Fatal(http.ListenAndServe(":1314", nil))
}

func writeJSON(w http.ResponseWriter, code int, body appResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
