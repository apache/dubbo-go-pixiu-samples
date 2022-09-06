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
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

import (
	"github.com/apache/dubbo-go-pixiu/pixiu/pkg/config"
	_ "github.com/apache/dubbo-go-pixiu/pixiu/pkg/pluginregistry"
	"github.com/apache/dubbo-go-pixiu/pixiu/pkg/server"

	"github.com/stretchr/testify/assert"
)

func TestHttpListenShutdown(t *testing.T) {
	count := int32(0)
	// start pixiu listener
	bootstrap := config.Load("../pixiu/conf.yaml")
	go server.Start(bootstrap)
	time.Sleep(3 * time.Second) // wait start allready

	// start client
	url := "http://localhost:8888/user/"
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, strings.NewReader(""))
	req_wg := &sync.WaitGroup{}
	req_wg.Add(3)
	assert.NoError(t, err)
	send_fenc := func() {
		rsp, err := client.Do(req)
		if err == nil && rsp != nil && rsp.StatusCode == 200 {
			atomic.AddInt32(&count, 1)
		}
		req_wg.Done()
	}
	go send_fenc()
	go send_fenc()

	// start shutdown
	time.Sleep(1 * time.Second) // wait request
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		// test shutdown reject
		time.Sleep(1 * time.Second)
		go send_fenc()
	}()
	server.GetServer().GetListenerManager().GetListenerService("0.0.0.0-8888-HTTP").ShutDown(wg)
	req_wg.Wait()
	assert.Equal(t, count, int32(2))
}
