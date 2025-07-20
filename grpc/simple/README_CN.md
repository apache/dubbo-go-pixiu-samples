# gRPC 简单示例

[English](./README.md)

本示例演示了如何使用 Apache Dubbo-go-pixiu 作为标准 gRPC 服务的网关，支持一元调用、客户端流、服务端流和双向流通信。

示例包括：
- 一个实现了 RouteGuide 服务的 gRPC 服务器。
- 一个用于 RouteGuide 服务的 gRPC 客户端。
- 一个 Pixiu 配置文件，用于将请求代理到 gRPC 服务器。

## 如何运行

您将需要三个独立的终端窗口。

### 1. 启动 gRPC 服务器

在第一个终端中，进入示例目录并启动 gRPC 服务器。

```sh
# 终端 1
cd dubbo-go-pixiu-samples/grpc/simple
go run server/server.go
```

服务器将启动并监听 `localhost:50051`。

### 2. 启动 Pixiu 网关

在第二个终端中，进入示例目录并启动 Pixiu 网关。

```sh
# 终端 2
cd dubbo-go-pixiu-samples
go run pixiu/*.go gateway start -c grpc/simple/pixiu/conf.yaml
```

Pixiu 将启动并监听 `localhost:8881` 上的 gRPC 请求。

### 3. 运行 gRPC 客户端

在第三个终端中，返回到示例目录并运行客户端。

```sh
# 终端 3
cd dubbo-go-pixiu-samples/grpc/simple
go run client/client.go
```

客户端将向 `localhost:8881` (Pixiu) 发送请求，然后 Pixiu 将这些请求代理到 gRPC 服务器。您应该能看到客户端与服务器交互并打印出功能信息。

## 运行测试

在运行测试前，请确保 gRPC 服务器和 Pixiu 网关已按照“如何运行”部分的指引启动。

打开一个新的终端，在项目根目录 (`dubbo-go-pixiu-samples`) 下执行以下命令来运行集成测试：

```sh
go test -v ./grpc/simple/test/
```

您将会看到所有子测试（一元调用、服务端流、客户端流和双向流）的成功执行日志。