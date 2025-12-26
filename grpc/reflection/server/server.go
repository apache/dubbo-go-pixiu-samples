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

// Package main implements a gRPC server with Server Reflection enabled.
// This demonstrates the gRPC Server Reflection feature for dynamic message parsing.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

import (
	pb "github.com/dubbo-go-pixiu/samples/grpc/reflection/proto"
)

var (
	port     = flag.Int("port", 50051, "The server port")
	serverID = flag.String("server_id", "", "Server identifier for load balancing verification")
)

// echoServer implements the EchoService.
type echoServer struct {
	pb.UnimplementedEchoServiceServer
	serverID string
}

// Echo returns the message back with additional metadata.
func (s *echoServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	log.Printf("[Echo] Received message: %s", req.Message)

	return &pb.EchoResponse{
		Message:           req.Message,
		ServerTimestamp:   time.Now().UnixNano(),
		ReflectionEnabled: true, // This server has reflection enabled
		ServerId:          s.serverID,
		Metadata:          req.Metadata,
	}, nil
}

// StreamEcho demonstrates server streaming with reflection support.
func (s *echoServer) StreamEcho(req *pb.EchoRequest, stream pb.EchoService_StreamEchoServer) error {
	log.Printf("[StreamEcho] Received message: %s", req.Message)

	// Send 5 responses for demonstration
	for i := 0; i < 5; i++ {
		resp := &pb.EchoResponse{
			Message:           fmt.Sprintf("[%d] %s", i+1, req.Message),
			ServerTimestamp:   time.Now().UnixNano(),
			ReflectionEnabled: true,
			ServerId:          s.serverID,
			Metadata:          req.Metadata,
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// ClientStreamEcho demonstrates client streaming with reflection support.
func (s *echoServer) ClientStreamEcho(stream pb.EchoService_ClientStreamEchoServer) error {
	log.Printf("[ClientStreamEcho] Stream started")

	var messages []string
	var lastMetadata map[string]string

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending, return aggregated response
			return stream.SendAndClose(&pb.EchoResponse{
				Message:           fmt.Sprintf("Received %d messages: %v", len(messages), messages),
				ServerTimestamp:   time.Now().UnixNano(),
				ReflectionEnabled: true,
				ServerId:          s.serverID,
				Metadata:          lastMetadata,
			})
		}
		if err != nil {
			return err
		}

		log.Printf("[ClientStreamEcho] Received: %s", req.Message)
		messages = append(messages, req.Message)
		lastMetadata = req.Metadata
	}
}

// BidirectionalEcho demonstrates bidirectional streaming.
func (s *echoServer) BidirectionalEcho(stream pb.EchoService_BidirectionalEchoServer) error {
	log.Printf("[BidirectionalEcho] Stream started")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		log.Printf("[BidirectionalEcho] Received: %s", req.Message)

		resp := &pb.EchoResponse{
			Message:           req.Message,
			ServerTimestamp:   time.Now().UnixNano(),
			ReflectionEnabled: true,
			ServerId:          s.serverID,
			Metadata:          req.Metadata,
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func getServerID() string {
	if *serverID != "" {
		return *serverID
	}
	// Use hostname as default server ID
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create gRPC server with default options
	grpcServer := grpc.NewServer()

	// Register the EchoService
	pb.RegisterEchoServiceServer(grpcServer, &echoServer{
		serverID: getServerID(),
	})

	// IMPORTANT: Enable gRPC Server Reflection
	// This allows Pixiu to dynamically discover and parse service methods
	reflection.Register(grpcServer)

	log.Printf("gRPC server with reflection enabled listening on port %d", *port)
	log.Printf("Server ID: %s", getServerID())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
