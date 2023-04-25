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
	dconfig "dubbo.apache.org/dubbo-go/v3/config"
	"dubbo.apache.org/dubbo-go/v3/config/generic" //nolint
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"

	hessian "github.com/apache/dubbo-go-hessian2"

	"github.com/apache/dubbo-go-pixiu/pixiu/pkg/config"
	_ "github.com/apache/dubbo-go-pixiu/pixiu/pkg/pluginregistry"
	"github.com/apache/dubbo-go-pixiu/pixiu/pkg/server"

	"github.com/stretchr/testify/assert"
)

func newDubboRefConf(iface, protocol string) dconfig.ReferenceConfig {
	refConf := dconfig.ReferenceConfig{
		InterfaceName:  iface,
		Cluster:        "failover",
		RegistryIDs:    []string{"zk"},
		Protocol:       protocol,
		Generic:        "true",
		URL:            "dubbo://127.0.0.1:8889/" + iface,
		Group:          "test",
		Version:        "1.0.0",
		RequestTimeout: "10s",
	}
	rootConfig := dconfig.NewRootConfigBuilder().
		Build()
	if err := dconfig.Load(dconfig.WithRootConfig(rootConfig)); err != nil {
		panic(err)
	}
	_ = refConf.Init(rootConfig)
	refConf.GenericLoad("dubbo.io")
	return refConf
}

func TestDubboListenShutdown(t *testing.T) {
	count := int32(0)

	// start pixiu listener
	bootstrap := config.Load("../pixiu/conf.yaml")
	go server.Start(bootstrap)
	time.Sleep(3 * time.Second) // wait start already

	// start client
	tripleRefConf := newDubboRefConf("com.dubbogo.pixiu.TripleUserService", dubbo.DUBBO)
	req_wg := &sync.WaitGroup{}
	req_wg.Add(2)
	call_func := func() {
		_, err := tripleRefConf.GetRPCService().(*generic.GenericService).Invoke(
			context.TODO(),
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
	assert.Equal(t, count, int32(1))
}
