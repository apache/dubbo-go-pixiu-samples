# gRPC Server Reflection 示例

[English](./README.md)

本示例演示如何使用 Apache Dubbo-go-pixiu 作为支持 **gRPC Server Reflection** 的网关。该功能支持在网关层进行动态消息解析和检查，无需预编译的 proto 文件。

> **相关 Issue**: [#821 - 添加 gRPC Server Reflection 支持](https://github.com/apache/dubbo-go-pixiu/issues/821)

## 概述

gRPC Server Reflection 功能（通过 [PR #849](https://github.com/apache/dubbo-go-pixiu/pull/849) 实现，解决 [Issue #821](https://github.com/apache/dubbo-go-pixiu/issues/821)）提供三种反射模式：

| 模式 | 描述 | 使用场景 |
|------|------|----------|
| **passthrough** | 二进制直通，不进行解码（默认） | 不需要消息检查的高性能场景 |
| **reflection** | 通过 gRPC 反射进行完整的动态消息解码 | 基于内容的路由、字段过滤、消息转换 |
| **hybrid** | 反射模式，失败时回退到直通模式 | 反射支持程度不一的混合环境 |

## 项目结构

```
grpc/reflection/
├── proto/
│   ├── echo.proto           # 服务定义
│   ├── echo.pb.go           # 生成的 protobuf 代码
│   └── echo_grpc.pb.go      # 生成的 gRPC 代码
├── server/
│   └── server.go            # 启用反射的 gRPC 服务器
├── client/
│   └── client.go            # 测试客户端
├── pixiu/
│   ├── conf.yaml            # 默认配置（reflection 模式）
│   ├── conf-passthrough.yaml # passthrough 模式配置
│   ├── conf-reflection.yaml  # reflection 模式配置
│   └── conf-hybrid.yaml      # hybrid 模式配置
├── test/
│   └── reflection_test.go   # 集成测试
├── README.md
└── README_CN.md
```

## 前置条件

- Go 1.21+
- Protocol Buffers 编译器 (protoc)
- 支持 gRPC Server Reflection 的 Apache Dubbo-go-pixiu（[PR #849](https://github.com/apache/dubbo-go-pixiu/pull/849)）

## 运行方法

需要三个独立的终端窗口。

### 1. 启动 gRPC 服务器

在第一个终端中，导航到项目根目录并启动启用了反射的 gRPC 服务器：

```sh
# 终端 1
cd dubbo-go-pixiu-samples
go run grpc/reflection/server/server.go
```

服务器将在 `localhost:50051` 上启动，并启用 gRPC Server Reflection。

### 2. 启动 Pixiu 网关

在第二个终端中，使用您选择的反射模式启动 Pixiu 网关：

```sh
# 终端 2：使用 reflection 模式（默认）
cd dubbo-go-pixiu-samples
go run pixiu/*.go gateway start -c grpc/reflection/pixiu/conf.yaml

# 可选：使用 passthrough 模式
# go run pixiu/*.go gateway start -c grpc/reflection/pixiu/conf-passthrough.yaml

# 可选：使用 hybrid 模式
# go run pixiu/*.go gateway start -c grpc/reflection/pixiu/conf-hybrid.yaml
```

Pixiu 将在 `localhost:8881` 上监听 gRPC 请求。

### 3. 运行客户端

在第三个终端中，运行测试客户端：

```sh
# 终端 3
cd dubbo-go-pixiu-samples
go run grpc/reflection/client/client.go
```

客户端将向 Pixiu 发送请求，Pixiu 将请求代理到 gRPC 服务器。您将看到演示所有三种 RPC 类型（一元调用、服务端流式、双向流式）的输出。

## 运行测试

确保 gRPC 服务器和 Pixiu 网关正在运行，然后执行：

```sh
go test -v ./grpc/reflection/test/
```

## 配置详情

### 反射模式配置

```yaml
grpc_filters:
  - name: dgp.filter.grpc.proxy
    config:
      # 选项: "passthrough" | "reflection" | "hybrid"
      reflection_mode: "reflection"

      # 方法描述符缓存 TTL（秒）
      descriptor_cache_ttl: 300

      # 启用 Triple 协议检测
      enable_protocol_detection: true

      # 反射操作超时时间（hybrid 模式）
      reflection_timeout: 5s
```

### 核心特性

1. **动态消息解码**：运行时解析和检查 gRPC 消息，无需 proto 文件
2. **基于 TTL 的缓存**：高效缓存方法描述符，支持自动清理
3. **协议检测**：支持 gRPC 和 Triple 协议兼容性
4. **回退机制**：hybrid 模式提供优雅降级

## 启用服务端反射

要在您的服务器上启用 gRPC Server Reflection，在注册服务后添加以下代码：

```go
import "google.golang.org/grpc/reflection"

// 注册服务后
reflection.Register(grpcServer)
```

## 相关资源

- [Issue #821: 添加 gRPC Server Reflection 支持](https://github.com/apache/dubbo-go-pixiu/issues/821) - 原始功能请求
- [PR #849: 功能实现](https://github.com/apache/dubbo-go-pixiu/pull/849) - 实现此功能的 Pull Request
- [gRPC Server Reflection 协议](https://github.com/grpc/grpc/blob/master/doc/server-reflection.md) - 官方 gRPC 反射规范
