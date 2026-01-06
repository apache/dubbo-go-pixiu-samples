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
	"io"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

import (
	pb "github.com/dubbo-go-pixiu/samples/grpc/reflection/proto"
)

const (
	// pixiuAddr is the default Pixiu gateway address for testing
	pixiuAddr = "localhost:8881"
)

func getClient(t *testing.T) (pb.EchoServiceClient, func()) {
	conn, err := grpc.NewClient(pixiuAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "Failed to connect to Pixiu gateway")

	cleanup := func() {
		conn.Close()
	}

	return pb.NewEchoServiceClient(conn), cleanup
}

func TestUnaryEcho(t *testing.T) {
	client, cleanup := getClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.EchoRequest{
		Message:   "Test message for reflection",
		Timestamp: time.Now().UnixNano(),
		Metadata: map[string]string{
			"test": "unary",
		},
	}

	resp, err := client.Echo(ctx, req)
	require.NoError(t, err, "Echo RPC should succeed")

	assert.Equal(t, req.Message, resp.Message, "Message should be echoed back")
	assert.NotEmpty(t, resp.ServerId, "Server ID should be set")
	assert.True(t, resp.ReflectionEnabled, "Reflection should be enabled on server")
	assert.Greater(t, resp.ServerTimestamp, int64(0), "Server timestamp should be positive")
}

func TestStreamEcho(t *testing.T) {
	client, cleanup := getClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.EchoRequest{
		Message:   "Stream test message",
		Timestamp: time.Now().UnixNano(),
		Metadata: map[string]string{
			"test": "stream",
		},
	}

	stream, err := client.StreamEcho(ctx, req)
	require.NoError(t, err, "StreamEcho should start successfully")

	messageCount := 0
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err, "StreamEcho receive should not fail")

		messageCount++
		assert.Contains(t, resp.Message, req.Message, "Response should contain original message")
		assert.NotEmpty(t, resp.ServerId, "Server ID should be set")
		assert.True(t, resp.ReflectionEnabled, "Reflection should be enabled")
	}

	assert.Equal(t, 5, messageCount, "Should receive exactly 5 streamed messages")
}

func TestClientStreamEcho(t *testing.T) {
	client, cleanup := getClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.ClientStreamEcho(ctx)
	require.NoError(t, err, "ClientStreamEcho should start successfully")

	messages := []string{"First", "Second", "Third"}

	// Send all messages
	for _, msg := range messages {
		req := &pb.EchoRequest{
			Message:   msg,
			Timestamp: time.Now().UnixNano(),
		}
		err := stream.Send(req)
		require.NoError(t, err, "Send should succeed")
	}

	// Close and receive response
	resp, err := stream.CloseAndRecv()
	require.NoError(t, err, "CloseAndRecv should succeed")

	assert.Contains(t, resp.Message, "3 messages", "Response should mention message count")
	assert.True(t, resp.ReflectionEnabled, "Reflection should be enabled")
	assert.NotEmpty(t, resp.ServerId, "Server ID should be set")
}

func TestBidirectionalEcho(t *testing.T) {
	client, cleanup := getClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.BidirectionalEcho(ctx)
	require.NoError(t, err, "BidirectionalEcho should start successfully")

	messages := []string{"Msg1", "Msg2", "Msg3"}

	// Send all messages
	for _, msg := range messages {
		req := &pb.EchoRequest{
			Message:   msg,
			Timestamp: time.Now().UnixNano(),
		}
		err := stream.Send(req)
		require.NoError(t, err, "Send should succeed")
	}
	stream.CloseSend()

	// Receive all responses
	receivedCount := 0
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err, "Receive should not fail")

		assert.Equal(t, messages[receivedCount], resp.Message, "Message should match")
		assert.True(t, resp.ReflectionEnabled, "Reflection should be enabled")
		receivedCount++
	}

	assert.Equal(t, len(messages), receivedCount, "Should receive same number of messages as sent")
}

func TestEchoWithMetadata(t *testing.T) {
	client, cleanup := getClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	metadata := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	req := &pb.EchoRequest{
		Message:   "Metadata test",
		Timestamp: time.Now().UnixNano(),
		Metadata:  metadata,
	}

	resp, err := client.Echo(ctx, req)
	require.NoError(t, err, "Echo should succeed")

	// Verify metadata is echoed back
	for k, v := range metadata {
		assert.Equal(t, v, resp.Metadata[k], "Metadata key %s should be echoed", k)
	}
}
