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

// OAuth tests for the MCP authorization sample.
// They verify 401 without token, successful tool listing and read-only calls
// with a 'read' token, and basic service availability checks.
package test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
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
)

// JSON-RPC types
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

// Helpers
// httpGetOK returns true when a GET call returns status 200.
func httpGetOK(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// getAccessToken obtains a bearer token from the local auth server using
// client_credentials with a specific scope (read/write).
func getAccessToken(t *testing.T, scope string) string {
	t.Helper()
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", "sample-client")
	form.Set("client_secret", "secret")
	form.Set("resource", pixiuBaseURL+mcpPath)
	form.Set("scope", scope)

	req, err := http.NewRequest(http.MethodPost, authBaseURL+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("build token request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("token request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 from token endpoint, got %d: %s", resp.StatusCode, string(b))
	}

	var obj struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	if obj.AccessToken == "" {
		t.Fatalf("empty access token")
	}
	return obj.AccessToken
}

// sendJSONRPC posts a JSON-RPC request to /mcp with an optional bearer token.
func sendJSONRPC(t *testing.T, method string, params any, token string) (int, []byte) {
	t.Helper()
	reqObj := JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params}
	body, err := json.Marshal(reqObj)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, pixiuBaseURL+mcpPath, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

// Tests
// TestServiceAvailability_OAuth ensures all services are up before running auth tests.
func TestServiceAvailability_OAuth(t *testing.T) {
	if !httpGetOK(backendBaseURL) {
		t.Skip("Backend service not available. Start: cd mcp/simple/server && go run server.go")
	}
	if !httpGetOK(authBaseURL + "/.well-known/oauth-authorization-server") {
		t.Skip("Auth server not available. Start: cd mcp/oauth/authserver && go run server.go")
	}
	if !httpGetOK(pixiuBaseURL) {
		t.Skip("Pixiu not available. Start gateway with mcp/oauth/pixiu/conf.yaml")
	}
}

// TestUnauthorized_NoToken verifies that calling /mcp without a bearer token
// returns 401 (or a JSON-RPC error envelope depending on proxy behavior).
func TestUnauthorized_NoToken(t *testing.T) {
	if !httpGetOK(pixiuBaseURL) {
		t.Skip("Pixiu not available")
	}
	status, body := sendJSONRPC(t, "tools/list", nil, "")
	if status != http.StatusUnauthorized {
		// Some proxies may return JSON-RPC error with 200; accept either 401 or JSON-RPC error
		// Try to parse content type and error
		ct, _, _ := mime.ParseMediaType(http.DetectContentType(body))
		if status == http.StatusOK && strings.Contains(string(body), "error") && strings.Contains(ct, "application/json") {
			t.Logf("Received 200 with error body (acceptable for some setups): %s", string(body))
			return
		}
		t.Fatalf("expected 401 without token, got %d body=%s", status, string(body))
	}
}

// TestAuthorized_ReadToken_ToolsList validates that a read token can list tools.
func TestAuthorized_ReadToken_ToolsList(t *testing.T) {
	if !httpGetOK(pixiuBaseURL) || !httpGetOK(authBaseURL+"/.well-known/oauth-authorization-server") {
		t.Skip("Services not available")
	}
	token := getAccessToken(t, "read")
	status, data := sendJSONRPC(t, "tools/list", nil, token)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", status, string(data))
	}
	var resp JSONRPCResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("json-rpc error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("result not object")
	}
	if _, ok := result["tools"].([]any); !ok {
		t.Fatalf("tools not found in result")
	}
}

// TestAuthorized_ReadToken_GetUser validates that a read token can call a read-only tool.
func TestAuthorized_ReadToken_GetUser(t *testing.T) {
	if !httpGetOK(pixiuBaseURL) || !httpGetOK(backendBaseURL) || !httpGetOK(authBaseURL+"/.well-known/oauth-authorization-server") {
		t.Skip("Services not available")
	}
	token := getAccessToken(t, "read")
	params := toolCallParams{
		Name: "get_user",
		Arguments: map[string]any{
			"id":              1,
			"include_profile": true,
		},
	}
	status, data := sendJSONRPC(t, "tools/call", params, token)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", status, string(data))
	}
	var resp JSONRPCResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("json-rpc error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("result not object")
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("content missing")
	}
}
