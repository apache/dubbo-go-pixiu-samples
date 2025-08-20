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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePKCE(t *testing.T) {
	testCases := []struct {
		name      string
		verifier  string
		challenge string
		expected  bool
	}{
		{
			name:      "Valid PKCE",
			verifier:  "test_verifier",
			challenge: calculateS256Challenge("test_verifier"),
			expected:  true,
		},
		{
			name:      "Invalid PKCE",
			verifier:  "wrong_verifier",
			challenge: calculateS256Challenge("test_verifier"),
			expected:  false,
		},
		{
			name:      "Empty Verifier",
			verifier:  "",
			challenge: calculateS256Challenge("test_verifier"),
			expected:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, validatePKCE(tc.challenge, tc.verifier))
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	// Test length
	s32 := generateRandomString(32)
	assert.Len(t, s32, 64) // 32 bytes = 64 hex characters

	s16 := generateRandomString(16)
	assert.Len(t, s16, 32) // 16 bytes = 32 hex characters

	// Test for randomness (not a perfect test, but checks for non-empty and different results)
	s32_another := generateRandomString(32)
	assert.NotEmpty(t, s32)
	assert.NotEqual(t, s32, s32_another, "Two generated strings should not be the same")
}

func TestHandleMetadata(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
	w := httptest.NewRecorder()

	handleMetadata(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var meta map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&meta)
	require.NoError(t, err)

	issuer := "http://localhost" + listenAddr
	assert.Equal(t, issuer, meta["issuer"])
	assert.Equal(t, issuer+"/oauth/authorize", meta["authorization_endpoint"])
	assert.Equal(t, issuer+"/oauth/token", meta["token_endpoint"])
	assert.Equal(t, issuer+"/.well-known/jwks.json", meta["jwks_uri"])
	assert.Equal(t, issuer+"/register", meta["registration_endpoint"])
}

func TestHandleJwks(t *testing.T) {
	initJWT()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	handleJwks(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var jwksResponse jwks
	err := json.NewDecoder(resp.Body).Decode(&jwksResponse)
	require.NoError(t, err)

	require.Len(t, jwksResponse.Keys, 1)
	key := jwksResponse.Keys[0]
	assert.Equal(t, "RSA", key.Kty)
	assert.Equal(t, keyID, key.Kid)
	assert.Equal(t, "sig", key.Use)
	assert.Equal(t, "RS256", key.Alg)
}

func TestHandleAuthorize(t *testing.T) {
	initStore()

	t.Run("Successful authorization", func(t *testing.T) {
		q := url.Values{}
		q.Set("client_id", "sample-client")
		q.Set("redirect_uri", "http://localhost:8081/callback")
		q.Set("response_type", "code")
		q.Set("code_challenge", "challenge")
		q.Set("code_challenge_method", "S256")
		q.Set("state", "12345")
		q.Set("resource", "test-resource")

		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handleAuthorize(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusFound, resp.StatusCode)

		loc, err := resp.Location()
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:8081/callback", loc.Scheme+"://"+loc.Host+loc.Path)

		code := loc.Query().Get("code")
		assert.NotEmpty(t, code)
		assert.Equal(t, "12345", loc.Query().Get("state"))

		// Check that the code was stored
		_, ok := authCodes[code]
		assert.True(t, ok, "Auth code should be stored")
	})

	t.Run("Invalid client ID", func(t *testing.T) {
		q := url.Values{}
		q.Set("client_id", "invalid-client")
		q.Set("redirect_uri", "http://localhost:8081/callback")
		q.Set("response_type", "code")
		q.Set("code_challenge", "challenge")
		q.Set("code_challenge_method", "S256")

		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handleAuthorize(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandleToken(t *testing.T) {
	initStore()
	initJWT()

	verifier := "test_verifier"
	challenge := calculateS256Challenge(verifier)

	t.Run("Successful token exchange", func(t *testing.T) {
		// 1. Setup: Store a valid auth code
		code := "test_code_success"
		authCodes[code] = AuthCodeInfo{
			ClientID:      "sample-client",
			RedirectURI:   "http://localhost:8081/callback",
			CodeChallenge: challenge,
			Resource:      "test-resource",
			Expiry:        time.Now().Add(10 * time.Minute),
		}

		// 2. Execute: Make the token request
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("client_id", "sample-client")
		data.Set("code_verifier", verifier)
		data.Set("resource", "test-resource")

		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handleToken(w, req)

		// 3. Assert: Check the response
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tokenResp tokenResponse
		err := json.NewDecoder(resp.Body).Decode(&tokenResp)
		require.NoError(t, err)

		assert.NotEmpty(t, tokenResp.AccessToken)
		assert.Equal(t, "Bearer", tokenResp.TokenType)
		assert.Equal(t, int64(tokenTTL.Seconds()), tokenResp.ExpiresIn)

		// Check that the auth code was deleted
		_, ok := authCodes[code]
		assert.False(t, ok, "Auth code should be deleted after use")
	})

	t.Run("Invalid code", func(t *testing.T) {
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", "invalid_code")
		data.Set("client_id", "sample-client")
		data.Set("code_verifier", verifier)

		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handleToken(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("PKCE verification failed", func(t *testing.T) {
		// 1. Setup: Store a valid auth code
		code := "test_code_pkce_fail"
		authCodes[code] = AuthCodeInfo{
			ClientID:      "sample-client",
			CodeChallenge: challenge,
			Resource:      "test-resource",
			Expiry:        time.Now().Add(10 * time.Minute),
		}

		// 2. Execute: Make the token request with a wrong verifier
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("client_id", "sample-client")
		data.Set("code_verifier", "wrong_verifier")

		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handleToken(w, req)

		// 3. Assert: Check for bad request
		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// Helper function for generating challenges in tests
func calculateS256Challenge(verifier string) string {
	hasher := sha256.New()
	hasher.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
}
