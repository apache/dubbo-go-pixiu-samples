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

package saml

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	defaultSAMLGatewayURL   = "http://127.0.0.1:8888"
	samlGatewayURLEnv       = "SAML_GATEWAY_URL"
	samlIntegrationTestsEnv = "RUN_SAML_INTEGRATION_TESTS"
)

var samlTestHTTPClient = &http.Client{Timeout: 5 * time.Second}

func envOrDefault(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func shouldRunSAMLIntegrationTests(shortMode bool, optIn string) bool {
	return !shortMode && optIn != ""
}

func requireSAMLIntegration(t *testing.T) string {
	t.Helper()

	if !shouldRunSAMLIntegrationTests(testing.Short(), os.Getenv(samlIntegrationTestsEnv)) {
		if testing.Short() {
			t.Skip("skipping SAML integration tests in short mode")
		}
		t.Skip("set RUN_SAML_INTEGRATION_TESTS=1 after starting Pixiu and Keycloak to run SAML integration tests")
	}

	gatewayURL := envOrDefault(samlGatewayURLEnv, defaultSAMLGatewayURL)
	if !checkSAMLServiceAvailable(gatewayURL + "/saml/metadata") {
		t.Skipf("saml gateway is unavailable at %s; start Pixiu and Keycloak first", gatewayURL)
	}

	return strings.TrimRight(gatewayURL, "/")
}

func checkSAMLServiceAvailable(url string) bool {
	resp, err := samlTestHTTPClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func isKeycloakSAMLRedirect(location string) bool {
	return strings.Contains(location, "18080") && strings.Contains(location, "/protocol/saml")
}

func TestShouldRunSAMLIntegrationTests(t *testing.T) {
	tests := []struct {
		name     string
		short    bool
		optIn    string
		expected bool
	}{
		{
			name:     "disabled in short mode",
			short:    true,
			optIn:    "1",
			expected: false,
		},
		{
			name:     "disabled without opt-in",
			short:    false,
			optIn:    "",
			expected: false,
		},
		{
			name:     "enabled with opt-in",
			short:    false,
			optIn:    "1",
			expected: true,
		},
	}

	for _, tt := range tests {
		if got := shouldRunSAMLIntegrationTests(tt.short, tt.optIn); got != tt.expected {
			t.Fatalf("%s: expected %v, got %v", tt.name, tt.expected, got)
		}
	}
}

func TestIsKeycloakSAMLRedirect(t *testing.T) {
	tests := []struct {
		name     string
		location string
		expected bool
	}{
		{
			name:     "matches keycloak saml endpoint",
			location: "http://localhost:18080/realms/pixiu/protocol/saml",
			expected: true,
		},
		{
			name:     "rejects non saml path on expected port",
			location: "http://localhost:18080/not-saml",
			expected: false,
		},
		{
			name:     "rejects saml path on unexpected host",
			location: "http://example.com/protocol/saml/login",
			expected: false,
		},
	}

	for _, tt := range tests {
		if got := isKeycloakSAMLRedirect(tt.location); got != tt.expected {
			t.Fatalf("%s: expected %v, got %v", tt.name, tt.expected, got)
		}
	}
}
