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

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const (
	// Test service addresses
	pixiuURL    = "http://localhost:8888"
	mcpEndpoint = "/mcp"
	backendURL  = "http://localhost:8081"
)

// JSONRPCRequest represents JSON-RPC request structure
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// JSONRPCResponse represents JSON-RPC response structure
type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// ToolCallParams represents tool call parameters
type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Send JSON-RPC request
func sendJSONRPCRequest(t *testing.T, method string, params any) *JSONRPCResponse {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(pixiuURL+mcpEndpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("JSON-RPC error: %v", jsonResp.Error)
	}

	return &jsonResp
}

// Check if service is available
func checkServiceAvailable(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// TestServiceAvailability tests service availability
func TestServiceAvailability(t *testing.T) {
	t.Log("Checking backend service availability...")
	if !checkServiceAvailable(backendURL) {
		t.Skip("Backend service is not available, skipping test. Please start backend service first: go run mcp/server/server.go")
	}

	t.Log("Checking Pixiu gateway availability...")
	if !checkServiceAvailable(pixiuURL) {
		t.Skip("Pixiu gateway is not available, skipping test. Please start Pixiu gateway first")
	}
}

// TestMCPInitialize tests MCP initialization
func TestMCPInitialize(t *testing.T) {
	if !checkServiceAvailable(pixiuURL) {
		t.Skip("Pixiu gateway is not available")
	}

	t.Log("Testing MCP initialization...")

	params := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"roots": map[string]any{
				"listChanged": true,
			},
		},
		"clientInfo": map[string]any{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	resp := sendJSONRPCRequest(t, "initialize", params)

	// Validate response structure
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be an object")
	}

	// Check server info
	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatalf("Expected serverInfo in result")
	}

	if name, ok := serverInfo["name"].(string); !ok || name == "" {
		t.Fatalf("Expected server name")
	}

	t.Logf("MCP server initialized successfully: %s", serverInfo["name"])
}

// TestToolsList tests getting tools list
func TestToolsList(t *testing.T) {
	if !checkServiceAvailable(pixiuURL) {
		t.Skip("Pixiu gateway is not available")
	}

	t.Log("Testing tools list...")

	resp := sendJSONRPCRequest(t, "tools/list", nil)

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be an object")
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("Expected tools array in result")
	}

	if len(tools) == 0 {
		t.Fatalf("Expected at least one tool")
	}

	t.Logf("Found %d tools", len(tools))

	// Validate tool structure
	for i, tool := range tools {
		toolObj, ok := tool.(map[string]any)
		if !ok {
			t.Fatalf("Tool %d is not an object", i)
		}

		if name, ok := toolObj["name"].(string); !ok || name == "" {
			t.Fatalf("Tool %d missing name", i)
		}

		if desc, ok := toolObj["description"].(string); !ok || desc == "" {
			t.Fatalf("Tool %d missing description", i)
		}

		t.Logf("Tool %d: %s - %s", i+1, toolObj["name"], toolObj["description"])
	}
}

// TestGetUser tests get user tool
func TestGetUser(t *testing.T) {
	if !checkServiceAvailable(pixiuURL) || !checkServiceAvailable(backendURL) {
		t.Skip("Services are not available")
	}

	t.Log("Testing get user tool...")

	params := ToolCallParams{
		Name: "get_user",
		Arguments: map[string]any{
			"id":              1,
			"include_profile": true,
		},
	}

	resp := sendJSONRPCRequest(t, "tools/call", params)

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be an object")
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("Expected content array in result")
	}

	contentItem, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("Expected content item to be an object")
	}

	text, ok := contentItem["text"].(string)
	if !ok {
		t.Fatalf("Expected text in content item")
	}

	// Parse returned user data
	var userData map[string]any
	if err := json.Unmarshal([]byte(text), &userData); err != nil {
		t.Fatalf("Failed to parse user data: %v", err)
	}

	// Validate user data
	if id, ok := userData["id"].(float64); !ok || id != 1 {
		t.Fatalf("Expected user ID to be 1")
	}

	if name, ok := userData["name"].(string); !ok || name == "" {
		t.Fatalf("Expected user name")
	}

	t.Logf("Successfully got user: %s", userData["name"])
}

// TestSearchUsers tests search users tool
func TestSearchUsers(t *testing.T) {
	if !checkServiceAvailable(pixiuURL) || !checkServiceAvailable(backendURL) {
		t.Skip("Services are not available")
	}

	t.Log("Testing search users tool...")

	params := ToolCallParams{
		Name: "search_users",
		Arguments: map[string]any{
			"q":     "alice",
			"page":  1,
			"limit": 10,
		},
	}

	resp := sendJSONRPCRequest(t, "tools/call", params)

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be an object")
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("Expected content array in result")
	}

	contentItem, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("Expected content item to be an object")
	}

	text, ok := contentItem["text"].(string)
	if !ok {
		t.Fatalf("Expected text in content item")
	}

	// Parse search result
	var searchResult map[string]any
	if err := json.Unmarshal([]byte(text), &searchResult); err != nil {
		t.Fatalf("Failed to parse search result: %v", err)
	}

	// Validate search result
	users, ok := searchResult["users"].([]any)
	if !ok {
		t.Fatalf("Expected users array in search result")
	}

	total, ok := searchResult["total"].(float64)
	if !ok {
		t.Fatalf("Expected total in search result")
	}

	t.Logf("Found %d users, total %.0f", len(users), total)
}

// TestCreateUser tests create user tool
func TestCreateUser(t *testing.T) {
	if !checkServiceAvailable(pixiuURL) || !checkServiceAvailable(backendURL) {
		t.Skip("Services are not available")
	}

	t.Log("Testing create user tool...")

	// Use timestamp to ensure unique email
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("test%d@example.com", timestamp)

	params := ToolCallParams{
		Name: "create_user",
		Arguments: map[string]any{
			"name":  "Test User",
			"email": email,
			"age":   25,
		},
	}

	resp := sendJSONRPCRequest(t, "tools/call", params)

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be an object")
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("Expected content array in result")
	}

	contentItem, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("Expected content item to be an object")
	}

	text, ok := contentItem["text"].(string)
	if !ok {
		t.Fatalf("Expected text in content item")
	}

	// Parse created user data
	var userData map[string]any
	if err := json.Unmarshal([]byte(text), &userData); err != nil {
		t.Fatalf("Failed to parse user data: %v", err)
	}

	// Validate created user data
	if name, ok := userData["name"].(string); !ok || name != "Test User" {
		t.Fatalf("Expected user name to be 'Test User'")
	}

	if userEmail, ok := userData["email"].(string); !ok || userEmail != email {
		t.Fatalf("Expected user email to be '%s'", email)
	}

	t.Logf("Successfully created user: %s (%s)", userData["name"], userData["email"])
}

// TestHealthCheck tests health check tool
func TestHealthCheck(t *testing.T) {
	if !checkServiceAvailable(pixiuURL) || !checkServiceAvailable(backendURL) {
		t.Skip("Services are not available")
	}

	t.Log("Testing health check tool...")

	params := ToolCallParams{
		Name:      "health_check",
		Arguments: map[string]any{},
	}

	resp := sendJSONRPCRequest(t, "tools/call", params)

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be an object")
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("Expected content array in result")
	}

	contentItem, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("Expected content item to be an object")
	}

	text, ok := contentItem["text"].(string)
	if !ok {
		t.Fatalf("Expected text in content item")
	}

	// Parse health check result
	var healthData map[string]any
	if err := json.Unmarshal([]byte(text), &healthData); err != nil {
		t.Fatalf("Failed to parse health data: %v", err)
	}

	// Validate health check result
	if status, ok := healthData["status"].(string); !ok || status != "healthy" {
		t.Fatalf("Expected status to be 'healthy'")
	}

	t.Logf("Health check passed: %s", healthData["status"])
}
