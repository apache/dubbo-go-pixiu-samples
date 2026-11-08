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
	"context"
	"errors"
)

import (
	"dubbo.apache.org/dubbo-go/v3"

	hessian "github.com/apache/dubbo-go-hessian2"
)

// User is the request and response message of GetUser.
type User struct {
	ID   string `hessian:"id" json:"id"`
	Code int64  `hessian:"code" json:"code"`
	Name string `hessian:"name" json:"name"`
	Age  int32  `hessian:"age" json:"age"`
}

func (User) JavaClassName() string {
	return "com.dubbogo.pixiu.User"
}

// GreeterProvider serves the generic calls that Pixiu forwards.
type GreeterProvider struct{}

func (*GreeterProvider) SayHello(_ context.Context, name string) (string, error) {
	return "Hello " + name, nil
}

// GetUser takes and returns the message by pointer so that the gson generalizer
// can create it with reflect.New and hand it over without a conversion.
func (*GreeterProvider) GetUser(_ context.Context, request *User) (*User, error) {
	if request.ID != "0001" {
		return nil, errors.New("user not found")
	}
	return &User{ID: "0001", Code: 1, Name: "tc", Age: 18}, nil
}

func (*GreeterProvider) Reference() string {
	return "com.dubbogo.pixiu.GenericGreeter"
}

func init() {
	hessian.RegisterPOJO(&User{})
	dubbo.SetProviderService(new(GreeterProvider))
}
