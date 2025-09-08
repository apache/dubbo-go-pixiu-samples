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
	"context"
	"errors"
	"github.com/dubbogo/gost/log/logger"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

type _testMetric struct {
	metricChan chan string
	buf        string
}

func (tt *_testMetric) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body, _ := io.ReadAll(request.Body)
	tt.buf = string(body)
	writer.WriteHeader(200)

	select {
	case tt.metricChan <- tt.buf:
	default:
	}
}

func waitForMetric(t *testing.T, metricServer *_testMetric, expectedSubstring string) {
	timeout := time.After(2 * time.Second)
	for {
		select {
		case receivedMetric := <-metricServer.metricChan:
			if strings.Contains(receivedMetric, expectedSubstring) {
				return
			}
		case <-timeout:
			t.Fatalf("timed out waiting for metric. last received: %q; expected to contain: %q", metricServer.buf, expectedSubstring)
		}
	}
}

func TestLocal(t *testing.T) {
	metricServer := &_testMetric{
		buf:        "",
		metricChan: make(chan string, 10),
	}

	go func() {
		server := &http.Server{Addr: ":9091", Handler: metricServer}

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Prometheus mock server exit with fail: %v", err)
		}

		server.Shutdown(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)

	verify(t, "http://localhost:8888/health", http.StatusOK)
	waitForMetric(t, metricServer, "pixiu_requests_total")

	s := verify(t, "http://localhost:8888/user", http.StatusOK)
	assert.True(t, strings.Contains(s, "user"))
	waitForMetric(t, metricServer, "pixiu_requests_total")

	s = verify(t, "http://localhost:8888/user/pixiu", http.StatusOK)
	assert.True(t, strings.Contains(s, "pixiu"))
	waitForMetric(t, metricServer, "pixiu_requests_total")

	s = verify(t, "http://localhost:8888/prefix", http.StatusOK)
	assert.True(t, strings.Contains(s, "prefix"))
	waitForMetric(t, metricServer, "pixiu_requests_total")
}

func verify(t *testing.T, url string, status int) string {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, status, resp.StatusCode)
	assert.NotNil(t, resp)
	s, _ := io.ReadAll(resp.Body)
	t.Log(string(s))
	return string(s)
}
