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

package service

import (
	"context"
)

import (
	"dubbo.apache.org/dubbo-go/v3/config"

	hessian "github.com/apache/dubbo-go-hessian2"
)

func init() {
	config.SetProviderService(&GreetingService{})
	hessian.RegisterPOJO(&GreetingRequest{})
	hessian.RegisterPOJO(&GreetingResponse{})
}

type GreetingService struct {
}

func (a *GreetingService) Greeting(ctx context.Context, in *GreetingRequest) (*GreetingResponse, error) {
	return &GreetingResponse{
		Greeting: "Hello " + in.Name + ", From Dubbo-go service",
	}, nil
}

func (a *GreetingService) Reference() string {
	return "GreetingService"
}

type GreetingResponse struct {
	Greeting string `json:"greeting"`
}

func (u GreetingResponse) JavaClassName() string {
	return "com.dubbo.demo.GreetingResponse"
}

type GreetingRequest struct {
	Name string
}

func (u GreetingRequest) JavaClassName() string {
	return "com.dubbo.demo.GreetingRequest"
}
