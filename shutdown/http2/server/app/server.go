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
	"net"
	"time"
)

import (
	"google.golang.org/grpc"
)

import (
	gproto "github.com/dubbo-go-pixiu/samples/grpc/deprecated/proto"
)

type grpcServer struct {
	users map[int32]*gproto.User
	gproto.UnimplementedUserProviderServer
}

func (s *grpcServer) GetUser(ctx context.Context, request *gproto.GetUserRequest) (*gproto.GetUserResponse, error) {
	// need 3s
	time.Sleep(3 * time.Second)
	return &gproto.GetUserResponse{Message: "receive"}, nil
}

func main() {
	listener, _ := net.Listen("tcp", ":50001")
	ser := &grpcServer{users: make(map[int32]*gproto.User)}
	gs := grpc.NewServer()
	gproto.RegisterUserProviderServer(gs, ser)
	gs.Serve(listener)
}
