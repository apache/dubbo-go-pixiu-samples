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

// Package main implements a gRPC client to test the EchoService through Pixiu.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"
)

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

import (
	pb "github.com/dubbo-go-pixiu/samples/grpc/reflection/proto"
)

var (
	serverAddr = flag.String("addr", "localhost:8881", "The Pixiu gateway address")
)

func main() {
	flag.Parse()

	// Connect to Pixiu gateway (or directly to gRPC server)
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewEchoServiceClient(conn)

	// Test all RPC types
	testUnaryEcho(client)
	testStreamEcho(client)
	testClientStreamEcho(client)
	testBidirectionalEcho(client)

	log.Println("All tests completed successfully!")
}

// testUnaryEcho tests the simple unary RPC.
func testUnaryEcho(client pb.EchoServiceClient) {
	log.Println("=== Testing Unary Echo ===")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.EchoRequest{
		Message:   "Hello, gRPC Server Reflection!",
		Timestamp: time.Now().UnixNano(),
		Metadata: map[string]string{
			"client":  "demo-client",
			"version": "1.0.0",
		},
	}

	resp, err := client.Echo(ctx, req)
	if err != nil {
		log.Fatalf("Echo failed: %v", err)
	}

	log.Printf("Response message: %s", resp.Message)
	log.Printf("Server ID: %s", resp.ServerId)
	log.Printf("Reflection enabled: %v", resp.ReflectionEnabled)
	log.Printf("Server timestamp: %d", resp.ServerTimestamp)
	log.Println()
}

// testStreamEcho tests server-side streaming RPC.
func testStreamEcho(client pb.EchoServiceClient) {
	log.Println("=== Testing Stream Echo (Server Streaming) ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.EchoRequest{
		Message:   "Stream message",
		Timestamp: time.Now().UnixNano(),
		Metadata: map[string]string{
			"stream_type": "server",
		},
	}

	stream, err := client.StreamEcho(ctx, req)
	if err != nil {
		log.Fatalf("StreamEcho failed: %v", err)
	}

	count := 0
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("StreamEcho recv failed: %v", err)
		}
		count++
		log.Printf("Received [%d]: %s (server: %s)", count, resp.Message, resp.ServerId)
	}
	log.Printf("Stream completed, received %d messages", count)
	log.Println()
}

// testClientStreamEcho tests client-side streaming RPC.
func testClientStreamEcho(client pb.EchoServiceClient) {
	log.Println("=== Testing Client Stream Echo (Client Streaming) ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.ClientStreamEcho(ctx)
	if err != nil {
		log.Fatalf("ClientStreamEcho failed: %v", err)
	}

	// Send multiple messages
	messages := []string{"First", "Second", "Third", "Fourth", "Fifth"}
	for _, msg := range messages {
		req := &pb.EchoRequest{
			Message:   msg,
			Timestamp: time.Now().UnixNano(),
			Metadata: map[string]string{
				"stream_type": "client",
			},
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("ClientStreamEcho send failed: %v", err)
		}
		log.Printf("Sent: %s", msg)
	}

	// Close and receive response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("ClientStreamEcho close failed: %v", err)
	}
	log.Printf("Response: %s (server: %s)", resp.Message, resp.ServerId)
	log.Println()
}

// testBidirectionalEcho tests bidirectional streaming RPC.
func testBidirectionalEcho(client pb.EchoServiceClient) {
	log.Println("=== Testing Bidirectional Echo ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.BidirectionalEcho(ctx)
	if err != nil {
		log.Fatalf("BidirectionalEcho failed: %v", err)
	}

	// Send and receive in goroutines
	messages := []string{
		"First bidirectional message",
		"Second bidirectional message",
		"Third bidirectional message",
	}

	// Send messages
	go func() {
		for i, msg := range messages {
			req := &pb.EchoRequest{
				Message:   msg,
				Timestamp: time.Now().UnixNano(),
				Metadata: map[string]string{
					"index": fmt.Sprintf("%d", i),
				},
			}
			if err := stream.Send(req); err != nil {
				log.Fatalf("BidirectionalEcho send failed: %v", err)
			}
			log.Printf("Sent: %s", msg)
			time.Sleep(100 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	// Receive responses
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("BidirectionalEcho recv failed: %v", err)
		}
		log.Printf("Received: %s (server: %s)", resp.Message, resp.ServerId)
	}
	log.Println("Bidirectional stream completed")
	log.Println()
}
