# gRPC Simple Example

[中文](./README_CN.md)

This example demonstrates how to use Apache Dubbo-go-pixiu as a gateway for a standard gRPC service, supporting unary, client-side streaming, server-side streaming, and bidirectional streaming RPCs.

The example includes:
- A gRPC server implementing the RouteGuide service.
- A gRPC client for the RouteGuide service.
- A Pixiu configuration file to proxy requests to the gRPC server.

## How to Run

You will need three separate terminal windows.

### 1. Start the gRPC Server

In the first terminal, navigate to the example directory and start the gRPC server.

```sh
# Terminal 1
cd dubbo-go-pixiu-samples/grpc/simple
go run server/server.go
```

The server will start and listen on `localhost:50051`.

### 2. Start the Pixiu Gateway

In the second terminal, navigate to the example directory and start the Pixiu gateway.

```sh
# Terminal 2
cd dubbo-go-pixiu-samples
go run pixiu/*.go gateway start -c grpc/simple/pixiu/conf.yaml
```

Pixiu will start and listen for gRPC requests on `localhost:8881`.

### 3. Run the gRPC Client

In the third terminal, navigate back to the example directory and run the client.

```sh
# Terminal 3
cd dubbo-go-pixiu-samples/grpc/simple
go run client/client.go
```

The client will send requests to Pixiu at `localhost:8881`, which then proxies them to the gRPC server. You should see the client interacting with the server and printing feature information.

## Running the Tests

Before running the tests, ensure that the gRPC server and the Pixiu gateway are running as described in the "How to Run" section.

Open a new terminal and run the following command from the project root (`dubbo-go-pixiu-samples`) to execute the integration tests:

```sh
go test -v ./grpc/simple/test/
```

You will see the successful execution logs for all subtests (unary, server-streaming, client-streaming, and bidirectional-streaming).
