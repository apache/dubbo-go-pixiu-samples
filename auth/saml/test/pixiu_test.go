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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleFile(rel string) string {
	return filepath.Join("..", filepath.FromSlash(rel))
}

func TestSampleAssetsExist(t *testing.T) {
	assets := []string{
		"pixiu/conf.yaml",
		"server/app/server.go",
		"docker/docker-compose.yml",
		"README.md",
		"README_CN.md",
		"certs/sp.crt",
		"certs/sp.key",
	}

	for _, asset := range assets {
		if _, err := os.Stat(sampleFile(asset)); err != nil {
			t.Fatalf("expected sample asset %s: %v", asset, err)
		}
	}
}

func TestPixiuConfigContainsSAMLFilter(t *testing.T) {
	data, err := os.ReadFile(sampleFile("pixiu/conf.yaml"))
	if err != nil {
		t.Fatalf("read pixiu config: %v", err)
	}

	content := string(data)
	required := []string{
		"dgp.filter.http.auth.saml",
		"acs_url:",
		"metadata_url:",
		"idp_metadata_url:",
		"allow_idp_initiated: true",
		"forward_attributes:",
		"X-User-Email",
		"X-User-Name",
		"/app",
	}

	for _, token := range required {
		if !strings.Contains(content, token) {
			t.Fatalf("expected pixiu config to contain %q", token)
		}
	}
}

func TestMetadataEndpoint(t *testing.T) {
	gatewayURL := requireSAMLIntegration(t)

	client := newSAMLTestHTTPClient()
	req, err := http.NewRequest(http.MethodGet, gatewayURL+"/saml/metadata", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request metadata endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from metadata endpoint, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read metadata response: %v", err)
	}

	text := string(body)
	if !strings.Contains(text, "pixiu-saml-sp") {
		t.Fatalf("expected metadata to contain entity ID, got %s", text)
	}
	if !strings.Contains(text, "AssertionConsumerService") {
		t.Fatalf("expected metadata to contain ACS endpoint, got %s", text)
	}
}

func TestProtectedRouteRedirectsToIDP(t *testing.T) {
	gatewayURL := requireSAMLIntegration(t)

	client := &http.Client{
		Timeout: samlTestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodGet, gatewayURL+"/app", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request protected route: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected redirect from protected route, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		t.Fatalf("expected redirect location header")
	}
	if !isKeycloakSAMLRedirect(location) {
		t.Fatalf("expected redirect to Keycloak, got %s", location)
	}
}
