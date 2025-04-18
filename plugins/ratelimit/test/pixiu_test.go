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
	"github.com/apache/dubbo-go-pixiu/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestRatelimit(t *testing.T) {
	url := "http://localhost:8888/v1/"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	var resp *http.Response

	var cnt200 int
	var cnt429 int
	for i := 0; i < 5; i++ {
		resp, err = client.Do(req)
		if resp.StatusCode == 200 {
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, 200, resp.StatusCode)
			s, _ := io.ReadAll(resp.Body)
			assert.True(t, strings.Contains(string(s), "resp"))
			cnt200++
			logger.Info("status: ", 200)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, 429, resp.StatusCode)
			cnt429++
			logger.Info("status: ", 429)
		}
	}

	assert.Equal(t, cnt200, 1)
	assert.Equal(t, cnt429, 4)

}
