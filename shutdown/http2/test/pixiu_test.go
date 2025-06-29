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
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

import (
	"github.com/apache/dubbo-go-pixiu/pkg/config"
	_ "github.com/apache/dubbo-go-pixiu/pkg/pluginregistry"
	"github.com/apache/dubbo-go-pixiu/pkg/server"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

import (
	gproto "github.com/dubbo-go-pixiu/samples/grpc/deprecated/proto"
)

func TestHttpListenShutdown(t *testing.T) {
	count := int32(0)
	// start pixiu listener
	bootstrap := config.Load("../pixiu/conf.yaml")
	go server.Start(bootstrap)
	time.Sleep(3 * time.Second) // wait start already

	// start client
	conn, err := grpc.Dial("localhost:8881", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	client := gproto.NewUserProviderClient(conn)
	call_wg := &sync.WaitGroup{}
	call_wg.Add(3)
	send_fenc := func() {
		_, err := client.GetUser(context.Background(), &gproto.GetUserRequest{UserId: 1})
		if err == nil {
			atomic.AddInt32(&count, 1)
		}
		call_wg.Done()
	}
	go send_fenc()
	go send_fenc()

	// shutdown
	time.Sleep(1 * time.Second) // wait request
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		// test shutdown reject
		time.Sleep(1 * time.Second)
		go send_fenc()
	}()
	server.GetServer().GetListenerManager().GetListenerService("0.0.0.0-8881-HTTP2").ShutDown(wg)
	call_wg.Wait()
	assert.Equal(t, count, int32(2))
}
