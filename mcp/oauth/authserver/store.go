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
	"time"
)

// ClientInfo holds the static information about a client.
// For this demo, we are hardcoding the clients.
type ClientInfo struct {
	ID           string
	Secret       string // Not used in public clients, but good practice to have.
	RedirectURIs []string
	// token endpoint auth method (e.g. "none" for public clients)
	TokenEndpointAuthMethod string
	// client_id_issued_at (unix seconds)
	ClientIDIssuedAt int64
}

// AuthCodeInfo holds the information associated with an authorization code.
type AuthCodeInfo struct {
	ClientID            string
	RedirectURI         string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string
	Expiry              time.Time
}

var (
	// clients stores the registered clients in memory.
	clients = make(map[string]ClientInfo)
	// authCodes stores the authorization codes in memory.
	authCodes = make(map[string]AuthCodeInfo)
)

// initStore initializes the in-memory data store.
func initStore() {
	// Initialize with a sample client for tests and local demos.
	clients["sample-client"] = ClientInfo{
		ID:                      "sample-client",
		Secret:                  "secret",
		RedirectURIs:            []string{"http://localhost:8081/callback"},
		TokenEndpointAuthMethod: "none",
		ClientIDIssuedAt:        time.Now().Unix(),
	}
}
