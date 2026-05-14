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
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

const (
	httpsDomain = "sample.domain.com"
	httpsPort   = "8443"
	httpsTarget = "127.0.0.1:" + httpsPort
	httpsAPI    = "https://" + httpsDomain + ":" + httpsPort + "/api/v1/test-dubbo/com.dubbogo.pixiu.UserService"
	waitTimeout = 180 * time.Second
)

func newHTTPSClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if addr == httpsDomain+":"+httpsPort {
					addr = httpsTarget
				}
				return dialer.DialContext(ctx, network, addr)
			},
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				ServerName:         httpsDomain,
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
	}
}

func post(t *testing.T, url string, data string) string {
	t.Helper()

	client := newHTTPSClient()
	deadline := time.Now().Add(waitTimeout)
	var lastBody string
	var lastErr error
	var lastStatus int

	for {
		req, err := http.NewRequest("POST", url, strings.NewReader(data))
		assert.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err == nil && resp != nil {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			lastBody = string(body)
			lastStatus = resp.StatusCode
			if resp.StatusCode == http.StatusOK {
				return lastBody
			}
		} else {
			lastErr = err
		}

		if time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if lastErr != nil {
		t.Fatalf("request failed: %v", lastErr)
	}
	t.Fatalf("unexpected status %d: %s", lastStatus, lastBody)
	return ""
}

func TestPost1(t *testing.T) {
	url := httpsAPI + "?group=test&version=1.0.0&method=GetUserByName"
	data := "{\"types\":\"string\",\"values\":\"tc\"}"
	s := post(t, url, data)
	assert.True(t, strings.Contains(s, "0001"))
}

func TestPost2(t *testing.T) {
	url := httpsAPI + "?group=test&version=1.0.0&method=UpdateUserByName"
	data := "{\"types\":\"string,object\",\"values\":[\"tc\",{\"id\":\"0001\",\"code\":1,\"name\":\"tc\",\"age\":15}]}"
	s := post(t, url, data)
	assert.Equal(t, "true", s)
}

func TestPost3(t *testing.T) {
	url := httpsAPI + "?group=test&version=1.0.0&method=GetUserByCode"
	data := "{\"types\":\"int\",\"values\":1}"
	s := post(t, url, data)
	assert.True(t, strings.Contains(s, "0001"))
}
