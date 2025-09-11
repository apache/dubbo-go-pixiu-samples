# Dubbo-go-pixiu MCP 使用 Nacos 作为注册中心示例

[English](./README.md) | 中文

本示例演示了如何使用 Nacos 3.0+ 作为 Dubbo-go-pixiu 网关的 MCP（模型配置协议）服务器。

## 准备工作

- Go 编程环境
- 已安装并启动 Nacos 3.0 或更高版本

## 步骤

### 1. 配置 Nacos MCP 服务

首先，我们需要在 Nacos 控制台中配置和发布一个 MCP 服务。

1.  **登录 Nacos 控制台**：访问 `http://<nacos-server-ip>:8080/nacos`。
2.  **进入 MCP 管理**：在左侧菜单栏找到并点击 “MCP管理”。
3.  **创建 MCP Server**：
    *   点击 “MCP列表” -> “创建MCP Server”。
    *   **类型**：选择 `streamable`。
    *   **工具(Tools)**：选择 “从OpenAPI导入”，然后上传 `mcp/nacos/mcptools/mcptools.yaml` 文件。
4.  **修正后端地址（重要）**：
    *   上传成功后，Nacos 会自动解析 `mcptools.yaml` 并生成工具列表。
    *   **注意**：Nacos 3.0 版本存在一个已知问题，导入的后端地址 `http://` 可能会错误地变成 `http:/`。请手动检查并修正所有工具的后端地址，确保其为 `http://localhost:8081`。
5.  **发布服务**：
    *   确认所有配置无误后，点击 “发布”，启动 MCP Server。
    *   *注意：目前 Pixiu 只支持连接单个 MCP Server。*

### 2. 启动后端模拟服务器

此服务器提供了 OpenAPI 文件中定义的 API 接口。

```bash
cd mcp/simple/server/app
go run .
```

启动成功后，你将看到类似以下的输出：

```
🚀 Mock Backend Server starting on :8081
📚 Available endpoints:
  GET  /api/users/{id}        - Get user by ID
  GET  /api/users/search      - Search users
  POST /api/users             - Create user
  GET  /api/users/{id}/posts  - Get user posts
  GET  /api/health            - Health check
  GET  /                      - Root endpoint
```
*请在新的终端窗口执行下一步，保持此服务器运行。*

### 3. 启动 Pixiu 网关

现在，启动 Pixiu 网关，它将连接到 Nacos MCP 服务并根据获取的配置进行路由。

```shell
cd /path/to/dubbo-go-pixiu
go run cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/mcp/nacos/pixiu/conf.yaml
```

### 4. 安装并启动 MCP Inspector Client

```shell
npx @modelcontextprotocol/inspector
```

在浏览器中打开 Inspector 界面，连接到 `http://localhost:8888/mcp` 便可以进行测试。