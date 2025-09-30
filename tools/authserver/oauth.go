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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"
)

// tokenResponse defines the structure of the JSON response from the token endpoint.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	// Parse query parameters
	query := r.URL.Query()
	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	responseType := query.Get("response_type")
	codeChallenge := query.Get("code_challenge")
	codeChallengeMethod := query.Get("code_challenge_method")
	resource := query.Get("resource")
	state := query.Get("state") // Preserve state parameter

	// Validate client
	client, ok := clients[clientID]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_client"})
		return
	}
	// Validate matching redirect URI against registered redirect_uris
	matched := false
	for _, ru := range client.RedirectURIs {
		if ru == redirectURI {
			matched = true
			break
		}
	}
	if !matched {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_redirect_uri"})
		return
	}

	// Validate request parameters
	if responseType != "code" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_response_type"})
		return
	}
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "error_description": "code_challenge required and must be S256"})
		return
	}

	// Require resource parameter
	if resource == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "error_description": "resource parameter required"})
		return
	}

	// In a real server, this is where you would authenticate the user and ask for consent.
	// For this demo, we auto-approve.

	// Generate and store authorization code
	code, err := generateRandomString(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server_error", "error_description": "failed to generate authorization code"})
		return
	}

	authCodes[code] = AuthCodeInfo{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Resource:            resource,
		Expiry:              time.Now().Add(10 * time.Minute),
	}

	// Redirect back to the client
	redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, code)
	if state != "" {
		redirectURL += "&state=" + state
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	// Validate grant type
	grantType := r.PostForm.Get("grant_type")
	if grantType != "authorization_code" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
		return
	}

	// Validate authorization code
	code := r.PostForm.Get("code")
	authCode, ok := authCodes[code]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	if time.Now().After(authCode.Expiry) {
		delete(authCodes, code) // Clean up expired code
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant", "error_description": "authorization code expired"})
		return
	}

	// Validate client and redirect URI
	clientID := r.PostForm.Get("client_id")
	if clientID != authCode.ClientID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_client"})
		return
	}

	// Require resource parameter and verify it matches the one associated with the auth code
	resource := r.PostForm.Get("resource")
	if resource == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "error_description": "resource parameter required"})
		return
	}
	if resource != authCode.Resource {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant", "error_description": "resource mismatch"})
		return
	}

	// Perform PKCE validation
	codeVerifier := r.PostForm.Get("code_verifier")
	if !validatePKCE(authCode.CodeChallenge, codeVerifier) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant", "error_description": "PKCE verification failed"})
		return
	}

	// All checks passed, clean up the auth code
	delete(authCodes, code)

	// Issue JWT
	accessToken, err := issueJWT(authCode.Resource, "") // Scope can be added here if needed
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server_error", "error_description": "failed to issue token"})
		return
	}

	// Return the token
	resp := tokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(tokenTTL.Seconds()),
	}
	writeJSON(w, http.StatusOK, resp)
}

// validatePKCE performs the S256 PKCE challenge verification.
func validatePKCE(challenge, verifier string) bool {
	hasher := sha256.New()
	hasher.Write([]byte(verifier))
	calculatedChallenge := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
	return calculatedChallenge == challenge
}
