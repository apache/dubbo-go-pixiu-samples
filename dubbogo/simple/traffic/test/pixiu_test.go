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
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestCanaryGET(t *testing.T) {
	url := "http://localhost:8888/user"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("canary-by-header", "v1")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	s, _ := io.ReadAll(resp.Body)
	assert.True(t, strings.Contains(string(s), `"server": "v1"`))
}

func TestCanaryGET1(t *testing.T) {
	url := "http://localhost:8888/user"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	s, _ := io.ReadAll(resp.Body)
	assert.True(t, strings.Contains(string(s), `"server": "v2"`))
}

func TestCanaryGET2(t *testing.T) {
	url := "http://localhost:8888/user"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("canary-by-header", "v3")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	s, _ := io.ReadAll(resp.Body)
	assert.True(t, strings.Contains(string(s), `"server": "v3"`))
}

func TestHeaderGET1(t *testing.T) {
	url := "http://localhost:8888/user"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("X-A", "t1")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	s, _ := io.ReadAll(resp.Body)
	assert.True(t, strings.Contains(string(s), `"server": "v1"`))
}

func TestHeaderGET2(t *testing.T) {
	url := "http://localhost:8888"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("X-C", "t1")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	s, _ := io.ReadAll(resp.Body)
	assert.True(t, strings.Contains(string(s), `"server": "v2"`))
}

func TestHeaderGET3(t *testing.T) {
	url := "http://localhost:8888"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("REG", "tt")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	s, _ := io.ReadAll(resp.Body)
	assert.True(t, strings.Contains(string(s), `"server": "v3"`))
}
