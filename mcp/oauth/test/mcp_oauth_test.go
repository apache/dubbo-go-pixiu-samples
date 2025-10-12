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

// OAuth tests for MCP authorization integration.
// Prerequisites: Authorization Server (port 9000), Backend API (port 8081), Pixiu Gateway (port 8888)
package test

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	pixiuBaseURL   = "http://localhost:8888"
	mcpPath        = "/mcp"
	backendBaseURL = "http://localhost:8081"
	authBaseURL    = "http://localhost:9000"
	clientID       = "sample-client"
	redirectURI    = "http://localhost:8081/callback"
)

// JSON-RPC request/response types
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func TestMain(m *testing.M) {
	if !waitForAllServices() {
		panic("Services not available. Start: Authorization Server (9000), Backend API (8081), Pixiu Gateway (8888)")
	}
	m.Run()
}

func waitForAllServices() bool {
	services := map[string]string{
		"Backend": backendBaseURL + "/api/health",
		"Auth":    authBaseURL + "/.well-known/oauth-authorization-server",
		"Pixiu":   pixiuBaseURL + mcpPath,
	}

	for name, url := range services {
		if !waitForService(name, url) {
			return false
		}
	}
	return true
}

func waitForService(name, url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 30; i++ {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK || (name == "Pixiu" && resp.StatusCode == http.StatusUnauthorized) {
				return true
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

// getAccessToken implements OAuth2 Authorization Code with PKCE flow
func getAccessToken(t *testing.T) string {
	t.Helper()
	codeVerifier := generateCodeVerifier(32)
	codeChallenge := generateCodeChallenge(codeVerifier)
	code := getAuthorizationCode(t, codeChallenge)
	return exchangeCodeForToken(t, code, codeVerifier)
}

func generateCodeVerifier(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateCodeChallenge(codeVerifier string) string {
	hasher := sha256.New()
	hasher.Write([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
}

func getAuthorizationCode(t *testing.T, codeChallenge string) string {
	t.Helper()
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	params.Set("resource", pixiuBaseURL+mcpPath)

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(authBaseURL + "/oauth/authorize?" + params.Encode())
	if err != nil {
		t.Fatalf("authorization request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected redirect, got %d: %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	parsedURL, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse redirect location: %v", err)
	}

	code := parsedURL.Query().Get("code")
	if code == "" {
		t.Fatalf("authorization code missing")
	}
	return code
}

func exchangeCodeForToken(t *testing.T, code, codeVerifier string) string {
	t.Helper()
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", clientID)
	form.Set("redirect_uri", redirectURI)
	form.Set("code_verifier", codeVerifier)
	form.Set("resource", pixiuBaseURL+mcpPath)

	req, err := http.NewRequest(http.MethodPost, authBaseURL+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("failed to build token request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("token request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 from token endpoint, got %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}

	if tokenResp.AccessToken == "" {
		t.Fatalf("access token is empty")
	}
	return tokenResp.AccessToken
}

func sendJSONRPC(t *testing.T, method string, params any, token string) (int, []byte) {
	t.Helper()
	reqObj := JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params}
	reqBody, err := json.Marshal(reqObj)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, pixiuBaseURL+mcpPath, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	return resp.StatusCode, body
}

// TestUnauthorizedAccess verifies requests without tokens are rejected
func TestUnauthorizedAccess(t *testing.T) {
	status, body := sendJSONRPC(t, "tools/list", nil, "")
	if status == http.StatusOK && strings.Contains(string(body), "error") {
		t.Logf("Received JSON-RPC error (acceptable): %s", string(body))
		return
	}
	if status != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", status, string(body))
	}
}

// TestAuthorizedAccess verifies OAuth2 flow and MCP operations
func TestAuthorizedAccess(t *testing.T) {
	token := getAccessToken(t)

	// Test tools/list
	status, body := sendJSONRPC(t, "tools/list", nil, token)
	if status != http.StatusOK {
		t.Fatalf("tools/list failed: %d, %s", status, string(body))
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("JSON-RPC error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("result is not an object")
	}
	tools, ok := result["tools"].([]any)
	if !ok || len(tools) == 0 {
		t.Fatalf("no tools found")
	}
	t.Logf("Retrieved %d tools", len(tools))

	// Test tool call
	params := toolCallParams{
		Name:      "health_check",
		Arguments: map[string]any{},
	}
	status, body = sendJSONRPC(t, "tools/call", params, token)
	if status != http.StatusOK {
		t.Fatalf("health_check failed: %d, %s", status, string(body))
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal tool call response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("tool call JSON-RPC error: %v", resp.Error)
	}
	t.Logf("Tool call successful")
}
