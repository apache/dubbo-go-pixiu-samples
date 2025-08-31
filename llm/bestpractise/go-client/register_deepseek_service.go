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

package main

import (
	"encoding/json"
	"fmt"
	"github.com/dubbogo/gost/log/logger"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// Nacos server configuration
const (
	NacosServerIP   = "127.0.0.1"
	NacosServerPort = 8848
	NacosNamespace  = "test_llm_registry_namespace"
	NacosGroup      = "test_llm_registry_group"

	ServiceName = "deepseek-service"
)

// Create and return a Nacos client
func createNacosClient() (naming_client.INamingClient, error) {
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(NacosServerIP, NacosServerPort),
	}
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(NacosNamespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("info"),
	)
	return clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
}

func main() {
	err := godotenv.Load("go-client/.env")
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	client, err := createNacosClient()
	if err != nil {
		logger.Fatalf("Failed to create nacos client: %v", err)
	}

	retryConfig := map[string]int{
		"times": 3,
	}
	retryConfigJSON, err := json.Marshal(retryConfig)
	if err != nil {
		logger.Fatalf("Unable to process config: %v", err)
	}

	metadata := map[string]string{
		"cluster": "chat",

		"id": "1",

		"llm-meta.provider": "deepseek",

		"llm-meta.retry_policy.name": "CountBased",

		"llm-meta.retry_policy.config": string(retryConfigJSON),

		"name": "deepseek-v2-chat-instance",

		"llm-meta.api_keys": os.Getenv("API_KEY"),

		"llm-meta.fallback": "false",
	}

	success, err := client.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: ServiceName,
		GroupName:   NacosGroup,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true, // temporary instance, will be removed on disconnect
		Metadata:    metadata,
	})

	if err != nil {
		logger.Fatalf("Register service instant failed: %v", err)
	}
	if !success {
		logger.Fatalf("Register service instant failed，please check Nacos logs")
	}

	logger.Infof("Service registered instance success on [%s] target cluster: [chat]", ServiceName)
	logger.Info("Programme running，press Ctrl+C to exit")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Received shut up signal，deregistering instance...")
	_, err = client.DeregisterInstance(vo.DeregisterInstanceParam{
		ServiceName: ServiceName,
		GroupName:   NacosGroup,
		Ephemeral:   true,
	})
	if err != nil {
		logger.Fatalf("Deregister instance failed: %v", err)
	}

	logger.Info("Service instance have been deregistered, exiting")
	time.Sleep(1 * time.Second)
}
