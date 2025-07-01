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
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

import (
	pb "github.com/dubbo-go-pixiu/samples/grpc/simple/routeguide"
	"github.com/stretchr/testify/assert"
)

const (
	serverAddr = "localhost:8881"
)

func TestRouteGuide(t *testing.T) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(serverAddr, opts...)
	assert.NoError(t, err, "fail to dial")
	defer conn.Close()
	client := pb.NewRouteGuideClient(conn)

	t.Run("unary", func(t *testing.T) {
		// Looking for a valid feature
		printFeature(t, client, &pb.Point{Latitude: 409146138, Longitude: -746188906})
		// Feature missing.
		printMissingFeature(t, client, &pb.Point{Latitude: 0, Longitude: 0})
	})

	t.Run("server-streaming", func(t *testing.T) {
		// Looking for features between 40, -75 and 42, -73.
		printFeatures(t, client, &pb.Rectangle{
			Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
			Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
		})
	})

	t.Run("client-streaming", func(t *testing.T) {
		runRecordRoute(t, client)
	})

	t.Run("bidirectional-streaming", func(t *testing.T) {
		runRouteChat(t, client)
	})
}

// printFeature gets the feature for the given point.
func printFeature(t *testing.T, client pb.RouteGuideClient, point *pb.Point) {
	t.Logf("Getting feature for point (%d, %d)", point.Latitude, point.Longitude)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	feature, err := client.GetFeature(ctx, point)
	assert.NoError(t, err, "client.GetFeature should not fail for a valid feature")
	assert.Equal(t, "Berkshire Valley Management Area Trail, Jefferson, NJ, USA", feature.Name)
	t.Log(feature)
}

// printMissingFeature gets a feature that doesn't exist.
func printMissingFeature(t *testing.T, client pb.RouteGuideClient, point *pb.Point) {
	t.Logf("Getting feature for point (%d, %d)", point.Latitude, point.Longitude)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	feature, err := client.GetFeature(ctx, point)
	assert.NoError(t, err, "client.GetFeature should not fail for a missing feature")
	assert.Empty(t, feature.Name, "Feature name should be empty for a missing feature")
	t.Log(feature)
}

// printFeatures lists all the features within the given bounding Rectangle.
func printFeatures(t *testing.T, client pb.RouteGuideClient, rect *pb.Rectangle) {
	t.Logf("Looking for features within %v", rect)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.ListFeatures(ctx, rect)
	assert.NoError(t, err, "client.ListFeatures should not fail")

	featureCount := 0
	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err, "stream.Recv should not fail")
		featureCount++
		t.Logf("Feature: name: %q, point:(%v, %v)", feature.GetName(),
			feature.GetLocation().GetLatitude(), feature.GetLocation().GetLongitude())
	}
	assert.Greater(t, featureCount, 0, "Should receive at least one feature")
}

// runRecordRoute sends a sequence of points to server and expects to get a RouteSummary from server.
func runRecordRoute(t *testing.T, client pb.RouteGuideClient) {
	// Create a random number of random points
	pointCount := int(rand.Int32N(100)) + 2 // Traverse at least two points
	var points []*pb.Point
	for i := 0; i < pointCount; i++ {
		points = append(points, randomPoint())
	}
	t.Logf("Traversing %d points.", len(points))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.RecordRoute(ctx)
	assert.NoError(t, err, "client.RecordRoute should not fail")

	for _, point := range points {
		err := stream.Send(point)
		assert.NoError(t, err, "stream.Send should not fail")
	}
	reply, err := stream.CloseAndRecv()
	assert.NoError(t, err, "stream.CloseAndRecv should not fail")
	assert.Equal(t, int32(len(points)), reply.PointCount, "PointCount should match the number of sent points")
	t.Logf("Route summary: %v", reply)
}

// runRouteChat receives a sequence of route notes, while sending notes for various locations.
func runRouteChat(t *testing.T, client pb.RouteGuideClient) {
	baseLatitude := rand.Int32()
	notes := []*pb.RouteNote{
		{Location: &pb.Point{Latitude: baseLatitude, Longitude: 1}, Message: "First message"},
		{Location: &pb.Point{Latitude: baseLatitude, Longitude: 2}, Message: "Second message"},
		{Location: &pb.Point{Latitude: baseLatitude, Longitude: 3}, Message: "Third message"},
		{Location: &pb.Point{Latitude: baseLatitude, Longitude: 1}, Message: "Fourth message"},
		{Location: &pb.Point{Latitude: baseLatitude, Longitude: 2}, Message: "Fifth message"},
		{Location: &pb.Point{Latitude: baseLatitude, Longitude: 3}, Message: "Sixth message"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.RouteChat(ctx)
	assert.NoError(t, err, "client.RouteChat should not fail")

	waitc := make(chan struct{})
	var receivedNotes []*pb.RouteNote
	var mu sync.Mutex
	go func() {
		defer close(waitc)
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if !assert.NoError(t, err, "stream.Recv in goroutine should not fail") {
				return
			}
			t.Logf("Got message %s at point(%d, %d)", in.Message, in.Location.Latitude, in.Location.Longitude)
			mu.Lock()
			receivedNotes = append(receivedNotes, in)
			mu.Unlock()
		}
	}()
	for _, note := range notes {
		err := stream.Send(note)
		assert.NoError(t, err, "stream.Send should not fail")
	}
	stream.CloseSend()
	<-waitc

	// Server logic sends back all previous notes at a location, so for 6 sent notes, we expect 9 received notes (1+1+1+2+2+2).
	assert.Equal(t, 9, len(receivedNotes), "Should receive the correct number of notes")
}

func randomPoint() *pb.Point {
	lat := (rand.Int32N(180) - 90) * 1e7
	long := (rand.Int32N(360) - 180) * 1e7
	return &pb.Point{Latitude: lat, Longitude: long}
}
