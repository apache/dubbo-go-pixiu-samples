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
	"time"
)

import (
	"github.com/dubbogo/gost/log/logger"
)

func init() {
	//config.SetProviderService(&UserProvider{})
}

type User struct {
	ID   string    `json:"iD,omitempty"`
	Name string    `json:"name,omitempty"`
	Age  int32     `json:"age,omitempty"`
	Time time.Time `json:"time,omitempty"`
}

type UserProvider struct{}

func (u *UserProvider) GetUserByNameAndAge(ctx context.Context, name string, age int32) (*User, error) {
	logger.Infof("GetUserByNameAndAge called with name: %s, age: %d", name, age)
	return &User{
		ID:   "001",
		Name: name,
		Age:  age,
		Time: time.Now(),
	}, nil
}

func (UserProvider) Reference() string {
	return "UserProvider"
}


