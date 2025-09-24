# Dubbo-go-pixiu MCP OAuth 授权示例

[English](./README.md) | 中文

## 简介

本示例演示了如何在 Pixiu Gateway 中集成 OAuth2 授权来保护 MCP (Model Context Protocol) 端点。该示例实现了完整的 OAuth2 授权码流程（带 PKCE），展示了三层架构的集成：

- **授权服务器** (localhost:9000) - 提供 OAuth2 授权服务，支持 PKCE
- **资源服务器** (Pixiu Gateway, localhost:8888) - 保护 MCP 端点 `/mcp`
- **后端 API 服务器** (localhost:8081) - 提供实际的业务 API

## 架构图

```
Client → Authorization Server (OAuth2) → Pixiu Gateway (MCP) → Backend API
        ↑                                  ↑                    ↑
    获取访问令牌                        JWT验证/MCP处理           实际API调用
```

## 快速开始

### 1. 启动后端 API 服务器

```bash
cd mcp/simple/server
go run server.go
# 服务将在 http://localhost:8081 启动
```

### 2. 启动 OAuth2 授权服务器

```bash
cd tools/authserver
go run *.go
# 授权服务器将在 http://localhost:9000 启动
```

### 3. 启动 Pixiu Gateway

```bash
# 在 dubbo-go-pixiu 项目根目录下执行
go run cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/mcp/oauth/pixiu/conf.yaml
# 网关将在 http://localhost:8888 启动，保护 /mcp 端点
```

### 4. 验证服务状态

```bash
# 检查授权服务器
curl http://localhost:9000/.well-known/oauth-authorization-server

# 检查后端 API（无需授权）
curl http://localhost:8081/api/health

# 检查 Pixiu MCP 端点（应返回 401 未授权）
curl http://localhost:8888/mcp
```

## OAuth2 授权流程测试

### 方式一：运行自动化测试

```bash
cd test
go test -v
```

测试将自动执行：
- 生成 PKCE 参数
- 获取授权码
- 交换访问令牌  
- 调用受保护的 MCP 端点

### 方式二：手动测试（模拟授权码流程）

由于授权服务器在演示模式下会自动批准请求，你可以手动模拟完整流程：

```bash
# 1. 生成 PKCE 参数（在实际应用中由客户端生成）
CODE_VERIFIER=$(head -c 32 /dev/urandom | base64 | tr -d "=+/" | cut -c1-43)
CODE_CHALLENGE=$(echo -n $CODE_VERIFIER | shasum -a 256 | cut -d' ' -f1 | xxd -r -p | base64 | tr -d "=+/")

# 2. 获取授权码（模拟浏览器重定向，授权服务器会自动批准）
AUTH_URL="http://localhost:9000/oauth/authorize?client_id=sample-client&redirect_uri=http://localhost:8081/callback&response_type=code&code_challenge=$CODE_CHALLENGE&code_challenge_method=S256&resource=http://localhost:8888/mcp"

# 访问授权URL并从重定向中提取code（需要手动操作）
echo "请访问: $AUTH_URL"
echo "从重定向URL中提取 code 参数，然后执行下一步"

# 3. 使用授权码交换访问令牌
read -p "请输入授权码: " AUTH_CODE
TOKEN=$(curl -s -X POST http://localhost:9000/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=$AUTH_CODE" \
  -d "client_id=sample-client" \
  -d "redirect_uri=http://localhost:8081/callback" \
  -d "code_verifier=$CODE_VERIFIER" \
  -d "resource=http://localhost:8888/mcp" | jq -r .access_token)

echo "获取到访问令牌: $TOKEN"
```

### 方式三：使用访问令牌调用 MCP API

```bash
# 获取工具列表
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' \
  http://localhost:8888/mcp | jq

# 调用健康检查工具
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"health_check","arguments":{}}}' \
  http://localhost:8888/mcp | jq

# 调用用户查询工具
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_user","arguments":{"id":1,"include_profile":true}}}' \
  http://localhost:8888/mcp | jq
```

## 配置说明

### OAuth2 配置 (pixiu/conf.yaml)

```yaml
# MCP 授权过滤器配置
- name: "dgp.filter.http.auth.mcp"
  config:
    resource_metadata:
      path: "/.well-known/oauth-protected-resource/mcp"
      resource: "http://localhost:8888/mcp"
      authorization_servers:
        - "http://localhost:9000"
    providers:
      - name: "local"
        issuer: "http://localhost:9000"
        jwks: "http://localhost:9000/.well-known/jwks.json"
    rules:
      - cluster: "mcp-protected"
```

### 关键特性

- **PKCE 支持**: 提高了授权码流程的安全性
- **JWT 验证**: 使用远程 JWKS 验证令牌签名
- **细粒度保护**: 仅保护 `/mcp` 端点，其他端点直通
- **MCP 集成**: 完整支持 MCP JSON-RPC 协议

## 故障排除

### 常见问题

1. **401 Unauthorized Error**
   - 检查访问令牌是否正确
   - 确认令牌未过期
   - 验证 `resource` 参数是否匹配

2. **Service Not Available**
   - 确保所有三个服务都已启动
   - 检查端口是否被占用
   - 查看服务日志了解详细错误

3. **JWT Verification Failed**
   - 确认授权服务器的 JWKS 端点可访问
   - 检查系统时间是否同步

### 日志查看

```bash
# 查看 Pixiu 日志（JWT 验证相关）
tail -f pixiu.log

# 查看授权服务器日志
# 在 authserver 目录下查看控制台输出
```

## 扩展用法

### 添加新的 MCP 工具

在 `pixiu/conf.yaml` 的 `tools` 部分添加新工具定义：

```yaml
- name: "custom_tool"
  description: "Custom tool description"
  cluster: "mock-server"
  request:
    method: "GET"
    path: "/api/custom"
    timeout: "10s"
```

## 安全注意事项

⚠️ **重要**: 此示例仅用于开发和演示目的，生产环境需要额外考虑：

- 使用 HTTPS 保护所有通信
- 实现真实的用户认证和授权
- 使用安全的密钥管理
- 添加请求限流和监控
- 实现令牌刷新机制
