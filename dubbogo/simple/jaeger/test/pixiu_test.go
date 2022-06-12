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
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/protobuf/types/known/timestamppb"
)

import (
	"github.com/dubbo-go-pixiu/samples/dubbogo/simple/jaeger/grpc/api_v2"
)

func GetTracesFromJaeger(t *testing.T) []*api_v2.Span {
	conn, err := grpc.Dial("localhost:16685", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()

	queryParams := &api_v2.TraceQueryParameters{
		ServiceName:  "dubbo-go-pixiu",
		StartTimeMin: timestamppb.New(time.Now().Add(time.Duration(-5) * time.Minute)),
		StartTimeMax: timestamppb.Now(),
	}
	client := api_v2.NewQueryServiceClient(conn)

	serverStream, err := client.FindTraces(context.Background(), &api_v2.FindTracesRequest{
		Query: queryParams,
	})
	require.NoError(t, err)

	spans := make([]*api_v2.Span, 0)
	for {
		spansResponseChunk, err := serverStream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		spans = append(spans, spansResponseChunk.Spans...)
	}
	return spans
}

func TestPost(t *testing.T) {
	url := "http://localhost:8881/api/v1/test-dubbo/user"
	data := "{\"id\":\"0003\",\"code\":3,\"name\":\"dubbogo\",\"age\":99}"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	s, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "{\"age\":99,\"code\":3,\"iD\":\"0003\",\"name\":\"dubbogo\"}", string(s))
}

func TestFindTraces(t *testing.T) {
	time.Sleep(5 * time.Second)

	operations := []string{"DUBBOGO CLIENT", "HTTP_POST"}
	spans := GetTracesFromJaeger(t)
	assert.Len(t, spans, len(operations))
	for i := 0; i < len(spans); i++ {
		assert.Equal(t, spans[i].OperationName, operations[i])
	}
}
