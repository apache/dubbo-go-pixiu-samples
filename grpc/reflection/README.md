# gRPC Server Reflection Example

[中文](./README_CN.md)

This example demonstrates how to use Apache Dubbo-go-pixiu as a gateway with **gRPC Server Reflection** support. This feature enables dynamic message parsing and inspection at the gateway level without requiring pre-compiled proto files.

> **Related Issue**: [#821 - Add gRPC Server Reflection support](https://github.com/apache/dubbo-go-pixiu/issues/821)

## Overview

The gRPC Server Reflection feature (implemented in [PR #849](https://github.com/apache/dubbo-go-pixiu/pull/849) to resolve [Issue #821](https://github.com/apache/dubbo-go-pixiu/issues/821)) provides three reflection modes:

| Mode | Description | Use Case |
|------|-------------|----------|
| **passthrough** | Binary pass-through without decoding (default) | High-performance scenarios where message inspection is not needed |
| **reflection** | Full dynamic message decoding via gRPC reflection | Content-based routing, field filtering, message transformation |
| **hybrid** | Reflection with passthrough fallback | Mixed environments with varying reflection support |

## Project Structure

```
grpc/reflection/
├── proto/
│   ├── echo.proto           # Service definition
│   ├── echo.pb.go           # Generated protobuf code
│   └── echo_grpc.pb.go      # Generated gRPC code
├── server/
│   └── server.go            # gRPC server with reflection enabled
├── client/
│   └── client.go            # Test client
├── pixiu/
│   ├── conf.yaml            # Default configuration (reflection mode)
│   ├── conf-passthrough.yaml # Passthrough mode configuration
│   ├── conf-reflection.yaml  # Reflection mode configuration
│   └── conf-hybrid.yaml      # Hybrid mode configuration
├── test/
│   └── reflection_test.go   # Integration tests
├── README.md
└── README_CN.md
```

## Prerequisites

- Go 1.21+
- Protocol Buffers compiler (protoc)
- Apache Dubbo-go-pixiu with gRPC Server Reflection support ([PR #849](https://github.com/apache/dubbo-go-pixiu/pull/849))

## How to Run

You will need three separate terminal windows.

### 1. Start the gRPC Server

In the first terminal, navigate to the project root and start the gRPC server with reflection enabled:

```sh
# Terminal 1
cd dubbo-go-pixiu-samples
go run grpc/reflection/server/server.go
```

The server will start on `localhost:50051` with gRPC Server Reflection enabled.

### 2. Start the Pixiu Gateway

In the second terminal, start the Pixiu gateway with your preferred reflection mode:

```sh
# Terminal 2: Using reflection mode (default)
cd dubbo-go-pixiu-samples
go run pixiu/*.go gateway start -c grpc/reflection/pixiu/conf.yaml

# Alternative: Use passthrough mode
# go run pixiu/*.go gateway start -c grpc/reflection/pixiu/conf-passthrough.yaml

# Alternative: Use hybrid mode
# go run pixiu/*.go gateway start -c grpc/reflection/pixiu/conf-hybrid.yaml
```

Pixiu will listen for gRPC requests on `localhost:8881`.

### 3. Run the Client

In the third terminal, run the test client:

```sh
# Terminal 3
cd dubbo-go-pixiu-samples
go run grpc/reflection/client/client.go
```

The client will send requests to Pixiu, which proxies them to the gRPC server. You should see output demonstrating all three RPC types (unary, server streaming, and bidirectional streaming).

## Running Tests

Ensure the gRPC server and Pixiu gateway are running, then execute:

```sh
go test -v ./grpc/reflection/test/
```

## Configuration Details

### Reflection Mode Configuration

```yaml
grpc_filters:
  - name: dgp.filter.grpc.proxy
    config:
      # Options: "passthrough" | "reflection" | "hybrid"
      reflection_mode: "reflection"

      # Cache TTL for method descriptors (seconds)
      descriptor_cache_ttl: 300

      # Enable Triple protocol detection
      enable_protocol_detection: true

      # Timeout for reflection operations (hybrid mode)
      reflection_timeout: 5s
```

### Key Features

1. **Dynamic Message Decoding**: Parse and inspect gRPC messages at runtime without proto files
2. **TTL-based Caching**: Efficient caching of method descriptors with automatic cleanup
3. **Protocol Detection**: Support for both gRPC and Triple protocol compatibility
4. **Fallback Mechanism**: Hybrid mode provides graceful degradation

## Enabling Server Reflection

To enable gRPC Server Reflection on your server, add this line after registering your service:

```go
import "google.golang.org/grpc/reflection"

// After registering your service
reflection.Register(grpcServer)
```

## Related Resources

- [Issue #821: Add gRPC Server Reflection support](https://github.com/apache/dubbo-go-pixiu/issues/821) - Original feature request
- [PR #849: Implementation](https://github.com/apache/dubbo-go-pixiu/pull/849) - Pull request implementing this feature
- [gRPC Server Reflection Protocol](https://github.com/grpc/grpc/blob/master/doc/server-reflection.md) - Official gRPC reflection specification
