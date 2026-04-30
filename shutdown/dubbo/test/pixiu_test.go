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
	"dubbo.apache.org/dubbo-go/v3/client"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/filter/generic"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"

	hessian "github.com/apache/dubbo-go-hessian2"

	"github.com/apache/dubbo-go-pixiu/pkg/config"
	_ "github.com/apache/dubbo-go-pixiu/pkg/pluginregistry"
	"github.com/apache/dubbo-go-pixiu/pkg/server"

	"github.com/stretchr/testify/assert"
)

func newDubboGenericService(iface, protocol string) *generic.GenericService {
	cli, err := client.NewClient(client.WithClientRequestTimeout(10 * time.Second))
	if err != nil {
		panic(err)
	}

	svc, err := cli.NewGenericService(
		iface,
		client.WithClusterFailOver(),
		client.WithProtocol(protocol),
		client.WithGroup("test"),
		client.WithVersion("1.0.0"),
		client.WithRequestTimeout(10*time.Second),
		client.WithURL("dubbo://127.0.0.1:8889/"+iface),
	)
	if err != nil {
		panic(err)
	}
	return svc
}

func newDubboContext(iface string) context.Context {
	return context.WithValue(context.Background(), constant.AttachmentKey, map[string]string{
		constant.PathKey:      iface,
		constant.InterfaceKey: iface,
		constant.GroupKey:     "test",
		constant.VersionKey:   "1.0.0",
	})
}

func TestDubboListenShutdown(t *testing.T) {
	count := int32(0)

	// start pixiu listener
	bootstrap := config.Load("../pixiu/conf.yaml")
	go server.Start(bootstrap)
	time.Sleep(3 * time.Second) // wait start already

	// start client
	tripleService := newDubboGenericService("com.dubbogo.pixiu.TripleUserService", dubbo.DUBBO)
	req_wg := &sync.WaitGroup{}
	req_wg.Add(2)
	call_func := func() {
		_, err := tripleService.Invoke(
			newDubboContext("com.dubbogo.pixiu.TripleUserService"),
			"TestByDubbo",
			[]string{"java.lang.String"},
			[]hessian.Object{"0001"},
		)
		if err == nil {
			atomic.AddInt32(&count, 1)
		}
		req_wg.Done()
	}
	go call_func()

	// shutdown
	time.Sleep(1 * time.Second) // wait request
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		// test shutdown reject
		time.Sleep(1 * time.Second)
		go call_func()
	}()
	server.GetServer().GetListenerManager().GetListenerService("0.0.0.0-8889-TCP").ShutDown(wg)
	req_wg.Wait()
	assert.Equal(t, int32(1), count)
}
