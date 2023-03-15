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

package prometheus

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

type _testMetric struct {
	buf []byte
}

func (tt *_testMetric) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	tt.buf, _ = ioutil.ReadAll(request.Body)
	//fmt.Printf("read (%d, content-length: %d) => %s\n", len(tt.buf), request.ContentLength, tt.buf)
	writer.WriteHeader(200)
}

func TestLocal(t *testing.T) {
	metricServer := &_testMetric{
		buf: make([]byte, 204800),
	}
	go http.ListenAndServe(":9091", metricServer)

	verify(t, "http://localhost:8888/health", http.StatusOK)
	assert.True(t, strings.Contains(string(metricServer.buf), "pixiu_requests_total"))
	metricServer.buf = metricServer.buf[0:0]

	s := verify(t, "http://localhost:8888/user", http.StatusOK)
	assert.True(t, strings.Contains(s, "user"))
	assert.True(t, strings.Contains(string(metricServer.buf), "pixiu_requests_total"))
	metricServer.buf = metricServer.buf[0:0]

	s = verify(t, "http://localhost:8888/user/pixiu", http.StatusOK)
	assert.True(t, strings.Contains(s, "pixiu"))
	assert.True(t, strings.Contains(string(metricServer.buf), "pixiu_requests_total"))
	metricServer.buf = metricServer.buf[0:0]

	s = verify(t, "http://localhost:8888/prefix", http.StatusOK)
	assert.True(t, strings.Contains(s, "prefix"))
	assert.True(t, strings.Contains(string(metricServer.buf), "pixiu_requests_total"))
	metricServer.buf = metricServer.buf[0:0]
}

func verify(t *testing.T, url string, status int) string {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, status, resp.StatusCode)
	assert.NotNil(t, resp)
	s, _ := ioutil.ReadAll(resp.Body)
	t.Log(string(s))
	return string(s)
}
