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
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	dconfig "dubbo.apache.org/dubbo-go/v3/config"
	"dubbo.apache.org/dubbo-go/v3/config/generic" //nolint

	hessian "github.com/apache/dubbo-go-hessian2"

	"github.com/apache/dubbo-go-pixiu/pixiu/pkg/config"
	_ "github.com/apache/dubbo-go-pixiu/pixiu/pkg/pluginregistry"
	"github.com/apache/dubbo-go-pixiu/pixiu/pkg/server"

	tpconst "github.com/dubbogo/triple/pkg/common/constant" //nolint

	"github.com/stretchr/testify/assert"
)

var count int32

func newTripleRefConf(iface, protocol string) dconfig.ReferenceConfig {

	refConf := dconfig.ReferenceConfig{
		InterfaceName:  iface,
		Cluster:        "failover",
		RegistryIDs:    []string{"zk"},
		Protocol:       protocol,
		Generic:        "true",
		Group:          "test",
		Version:        "1.0.0",
		URL:            "tri://127.0.0.1:9999/" + iface + "?" + constant.SerializationKey + "=hessian2",
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

func TestTripleListenShutdown(t *testing.T) {
	count = 0

	// start pixiu listener
	bootstrap := config.Load("../pixiu/conf.yaml")
	go server.Start(bootstrap)
	time.Sleep(3 * time.Second) // wait start already

	// start client
	tripleRefConf := newTripleRefConf("com.dubbogo.pixiu.TripleUserService", tpconst.TRIPLE)
	call_wg := &sync.WaitGroup{}
	call_wg.Add(3)
	call_func := func() {
		rsp, err := tripleRefConf.GetRPCService().(*generic.GenericService).Invoke(
			context.TODO(),
			"TestByTriple",
			[]string{"java.lang.String"},
			[]hessian.Object{"0001"},
		)
		if rsp != nil && err == nil {
			atomic.AddInt32(&count, 1)
		}
		call_wg.Done()
	}
	go call_func()
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
	server.GetServer().GetListenerManager().GetListenerService("0.0.0.0-9999-TRIPLE").ShutDown(wg)
	call_wg.Wait()
	assert.Equal(t, count, int32(2))
}
