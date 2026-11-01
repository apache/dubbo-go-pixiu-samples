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
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	traffic "github.com/dubbo-go-pixiu/samples/dubbogo/simple/traffic"
)

func trafficURL(t *testing.T, path string) string {
	t.Helper()
	config, err := traffic.LoadConfig()
	if err != nil {
		t.Fatalf("invalid traffic test configuration: %v", err)
	}
	return strings.TrimRight(config.GatewayURL, "/") + "/" + strings.TrimLeft(path, "/")
}

func trafficHTTPClient(t *testing.T) *http.Client {
	t.Helper()
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err == nil && host == "localhost" {
			address = net.JoinHostPort("127.0.0.1", port)
		}
		return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, address)
	}
	t.Cleanup(transport.CloseIdleConnections)
	return &http.Client{Timeout: 5 * time.Second, Transport: transport}
}
