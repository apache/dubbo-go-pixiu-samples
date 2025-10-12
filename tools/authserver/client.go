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
	"time"
)

// dynamicClientRegistrationRequest represents the RFC 7591 minimal client metadata accepted by /register.
type dynamicClientRegistrationRequest struct {
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
}

// handleDynamicClientRegistration implements a minimal RFC 7591 dynamic client registration endpoint.
func handleDynamicClientRegistration(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	var req dynamicClientRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	// Basic validation: require at least one redirect URI and simple scheme check.
	if len(req.RedirectURIs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_redirect_uris", "error_description": "redirect_uris must be provided"})
		return
	}
	for _, ru := range req.RedirectURIs {
		if !strings.HasPrefix(ru, "http://") && !strings.HasPrefix(ru, "https://") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_redirect_uri", "error_description": "redirect_uris must be absolute http(s) URLs"})
			return
		}
	}

	// Default token endpoint auth method
	if req.TokenEndpointAuthMethod == "" {
		req.TokenEndpointAuthMethod = "none"
	}

	clientID, err := generateRandomString(16)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server_error", "error_description": "failed to generate client ID"})
		return
	}

	var clientSecret string
	if req.TokenEndpointAuthMethod != "none" {
		clientSecret, err = generateRandomString(32)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server_error", "error_description": "failed to generate client secret"})
			return
		}
	}

	now := time.Now().Unix()
	client := ClientInfo{
		ID:                      clientID,
		Secret:                  clientSecret,
		RedirectURIs:            req.RedirectURIs,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		ClientIDIssuedAt:        now,
	}

	clients[clientID] = client

	issuer := issuerBaseURL // Use shared constant
	regURI := issuer + "/register/" + clientID

	resp := map[string]interface{}{
		"client_id":                  client.ID,
		"redirect_uris":              client.RedirectURIs,
		"client_id_issued_at":        client.ClientIDIssuedAt,
		"token_endpoint_auth_method": client.TokenEndpointAuthMethod,
		"registration_client_uri":    regURI,
	}
	if client.Secret != "" {
		resp["client_secret"] = client.Secret
	}

	w.Header().Set("Location", regURI)
	writeJSON(w, http.StatusCreated, resp)
}
