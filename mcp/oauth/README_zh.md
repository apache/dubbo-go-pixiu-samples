# Dubbo-go-pixiu MCP 授权（OAuth）示例

## 简介

本示例演示如何在 Pixiu 中启用 MCP 授权过滤器，仅保护 `/mcp` 端点，并区分 `read` 与 `write` 的访问范围（scope）。样例提供一个本地授权服务器（`client_credentials` 流程），并通过远程 JWKS（`http://localhost:9000/.well-known/jwks.json`）供 Pixiu 验证令牌签名。

## 目录结构

```
mcp/oauth/
  ├── authserver/        # 本地授权服务器
  │   └── server.go
  ├── pixiu/
  │   └── conf.yaml      # 仅保护 /mcp，使用远程 JWKS
  └── test/              # （后续可加入）授权相关测试
```

## 启动步骤

1) 启动后端服务（与 simple 共用）
```bash
cd mcp/simple/server
go run server.go
```

2) 启动本地授权服务器（端口 9000）
```bash
cd mcp/oauth/authserver
go run server.go
```

3) 启动 Pixiu（端口 8888，加载本示例配置）
```bash
cd /path/to/dubbo-go-pixiu
go run cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/mcp/oauth/pixiu/conf.yaml
```

## 获取访问令牌（client_credentials）

```bash
curl -s -X POST \
  -d "grant_type=client_credentials" \
  -d "client_id=sample-client" \
  -d "client_secret=secret" \
  -d "resource=http://localhost:8888/mcp" \
  -d "scope=read" \
  http://localhost:9000/oauth/token | jq -r .access_token
```

将返回的 `access_token` 用于后续请求：

```bash
TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImRlbW8ta2V5LTEifQ.eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjkwMDAiLCJhdWQiOiJodHRwOi8vbG9jYWxob3N0Ojg4ODgvbWNwIiwic2NvcGUiOiJyZWFkIiwiaWF0IjoxNzU1Mjc1MjMwLCJleHAiOjE3NTUyNzg4MzB9.J9XFrb620F4pBUKuEdh3SZ68x57PUWaYwVGM1jow001Z96d5gfw8tMy5XjcLOUFsPfH9cuqcIx83rrWQx0sj6rYIQc7Rf26JOShAhA0n68Y3SeRNoCkoQs5tT6TpecQvbG08cWe-080r2TKA3mUFrd3NcAHrzrusKEcNEMl65-tuxX9MWl2V1HNwDWZvIywxzcz2_efHsMakyWwYHxM0FHIuVcX_89P7dAneb4pYQz1nkGnRiELgaXm_87WG6JB8bg6FqiFprcMFIqZ7F2q4Yd28tIIzrU7j4OuL008bNy-oZQhcx9McqDXUW2DZvXBSuDLs6e1tHf3d58D0NW2C7A"
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8888/mcp -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

## 说明

- 远程 JWKS：`jwks: "http://localhost:9000/.well-known/jwks.json"`
- 仅保护 `/mcp`：直通后端 `/api` 不受保护
- 令牌字段：`iss=http://localhost:9000`、`aud=http://localhost:8888/mcp`、`scope=read|write`
