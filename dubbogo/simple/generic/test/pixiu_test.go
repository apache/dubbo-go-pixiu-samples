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
	"io"
	"net/http"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gatewayURL = "http://localhost:8881"

func post(t *testing.T, path string, body string) (int, string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, gatewayURL+path, bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(raw)
}

// TestMapMode exercises the default map mode: the argument list carries typed
// values and the result of a string method is written as a JSON string.
func TestMapMode(t *testing.T) {
	status, body := post(t, "/api/v1/generic/map", `{"types":["java.lang.String"],"values":["pixiu"]}`)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, `"Hello pixiu"`, body)
}

// TestGsonMode exercises the JSON text mode: the whole request message travels
// as text and the provider answers with JSON text that Pixiu writes verbatim.
func TestGsonMode(t *testing.T) {
	status, body := post(t, "/api/v1/generic/gson", `{"values":{"id":"0001"}}`)
	require.Equal(t, http.StatusOK, status)

	// A quoted JSON string here would mean the response text was encoded twice.
	require.True(t, json.Valid([]byte(body)), "response is not JSON text: %s", body)
	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &got))
	assert.Equal(t, map[string]any{
		"id":   "0001",
		"code": float64(1),
		"name": "tc",
		"age":  float64(18),
	}, got)
}
