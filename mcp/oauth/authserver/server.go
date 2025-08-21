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
	"log"
	"net/http"
)

const (
	listenAddr = ":9000"
	// issuerBaseURL is the base URL for the OAuth issuer
	issuerBaseURL = "http://localhost:9000"
)

func main() {
	// Initialize data stores and JWT keys.
	initStore()
	initJWT()

	// Setup HTTP routes.
	http.HandleFunc("/register", handleDynamicClientRegistration)
	http.HandleFunc("/.well-known/oauth-authorization-server", handleMetadata)
	http.HandleFunc("/.well-known/jwks.json", handleJwks)
	http.HandleFunc("/oauth/authorize", handleAuthorize)
	http.HandleFunc("/oauth/token", handleToken)

	log.Printf("OAuth Authorization Server listening on %s", listenAddr)

	// Start the server.
	if err := http.ListenAndServe(listenAddr, corsMiddleware(http.DefaultServeMux)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
