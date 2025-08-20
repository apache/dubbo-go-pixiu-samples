# Dubbo-go-pixiu MCP 示例

## 简介

本示例演示了如何使用 Dubbo-go-pixiu 的 MCP (Model Context Protocol) 过滤器，将后端 HTTP API 服务暴露为 MCP 工具，供大型语言模型 (LLM) 调用。

MCP 是一种专为 LLM 与外部工具交互而设计的协议，基于 JSON-RPC 2.0 实现。通过 MCP 过滤器，LLM 可以发现并调用后端服务的 API，无需修改现有的后端代码。

## 快速开始

### 1. 启动后端服务

```shell
cd mcp/simple/server
go run server.go
```

后端服务将在 8081 端口启动，提供用户管理相关的 API。

### 2. 启动 Pixiu 网关

```shell
cd /path/to/dubbo-go-pixiu
go run cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/mcp/simple/pixiu/conf.yaml
```

Pixiu 网关将在 8888 端口启动，MCP 端点为 `/mcp`。

### 3. 安装并启动 MCP Inspector Client

```shell
npx @modelcontextprotocol/inspector
```

在浏览器中打开 Inspector 界面，连接到 `http://localhost:8888/mcp`。

连接成功后，你可以测试以下工具：

## 配置说明

本示例使用的配置文件位于 `mcp/simple/pixiu/conf.yaml`，主要配置包括：

### 服务器信息

```yaml
server_info:
  name: "MCP Tools Test Server"
  version: "1.0.0"
  description: "MCP Server for testing all tools functionality"
  instructions: "Use these tools to interact with the mock server API"
```

### 工具配置

每个工具对应一个后端 API 端点：

```yaml
tools:
  - name: "get_user"                    # 工具名称
    description: "Get user by ID"       # 工具描述
    cluster: "mock-backend"             # 目标集群
    request:
      method: "GET"                     # HTTP 方法
      path: "/api/users/{id}"           # API 路径
      timeout: "10s"                    # 超时时间
    args:                               # 参数定义
      - name: "id"
        type: "integer"
        in: "path"
        description: "User ID"
        required: true
```

### 参数位置说明

- `path`：路径参数，如 `/api/users/{id}` 中的 `id`
- `query`：查询参数，如 `?page=1&limit=10`
- `body`：请求体参数，用于 POST/PUT 请求
- `header`：请求头参数

## 后端服务

示例后端服务 (`mcp/simple/server/server.go`) 提供了完整的用户管理 API，包括：

- 用户 CRUD 操作
- 用户搜索功能
- 用户帖子管理
- 健康检查接口

服务启动后会在控制台显示所有可用的 API 端点。

## 测试

本示例提供了测试用例来验证 MCP 功能的正确性。

### 运行测试

1. 确保服务已启动（按照快速开始部分的步骤）
2. 运行测试：

```shell
cd mcp/simple/test
go test -v
```

### 测试用例

测试套件包含以下测试用例：

- **TestServiceAvailability** - 检查服务可用性
- **TestMCPInitialize** - 测试 MCP 初始化
- **TestToolsList** - 测试获取工具列表
- **TestGetUser** - 测试获取用户工具
- **TestSearchUsers** - 测试搜索用户工具
- **TestCreateUser** - 测试创建用户工具
- **TestHealthCheck** - 测试健康检查工具

## 故障排除

1. **连接失败**：确保 Pixiu 网关和后端服务都已启动
2. **工具调用失败**：检查配置文件中的集群名称和路径是否正确
3. **参数错误**：确保传递的参数类型和格式符合工具定义
4. **测试失败**：检查服务是否正常运行，端口是否被占用

更多信息请参考 [MCP 官方文档](https://github.com/modelcontextprotocol/specification)。

