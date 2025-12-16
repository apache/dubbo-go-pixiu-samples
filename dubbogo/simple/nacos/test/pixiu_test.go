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

func TestPost1(t *testing.T) {
	url := "http://localhost:8881/BDTService/com.dubbogo.pixiu.UserService/GetUserByName"
	data := "[\"tc\"]"
	client := &http.Client{Timeout: 5 * time.Second}

	// Retry logic: wait for service registration in Nacos and Pixiu route setup
	var resp *http.Response
	var err error
	maxRetries := 15
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		req, reqErr := http.NewRequest("POST", url, strings.NewReader(data))
		assert.NoError(t, reqErr)
		req.Header.Set("x-dubbo-http1.1-dubbo-version", "1.0.0")
		req.Header.Set("x-dubbo-service-protocol", "dubbo")
		req.Header.Set("x-dubbo-service-version", "1.0.0")
		req.Header.Set("x-dubbo-service-group", "test")
		req.Header.Add("Content-Type", "application/json")

		resp, err = client.Do(req)
		if err == nil && resp != nil && resp.StatusCode == 200 {
			break
		}

		t.Logf("Attempt %d/%d: waiting for service registration... (status: %v)", i+1, maxRetries, resp)
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(retryInterval)
	}

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	s, _ := io.ReadAll(resp.Body)
	assert.True(t, strings.Contains(string(s), "0001"))
}
