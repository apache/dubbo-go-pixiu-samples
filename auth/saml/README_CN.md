# SAML 认证示例

[English](./README.md) | 中文

## 概览

这个示例演示如何使用 `dgp.filter.http.auth.saml` 过滤器保护 Pixiu 的
HTTP 路由。

示例包含三部分：

- Keycloak：作为 SAML Identity Provider（IdP）
- Pixiu：作为 SAML Service Provider（SP）
- 一个简单后端服务：回显 Pixiu 转发过来的 SAML 属性

登录成功后，Pixiu 会把 SAML 属性转换成请求头并转发给后端：

- `email -> X-User-Email`
- `displayName -> X-User-Name`

后端会把这些字段以 JSON 返回，方便确认整个 SAML 流程已经打通。

## 文件结构

```text
auth/
└── saml/
    ├── certs/
    │   ├── sp.crt
    │   └── sp.key
    ├── docker/
    │   ├── docker-compose.yml
    │   └── docker-health-check.sh
    ├── pixiu/
    │   └── conf.yaml
    ├── server/
    │   └── app/
    │       └── server.go
    └── test/
        └── pixiu_test.go
```

## 前置条件

- 已安装 Docker
- 已安装 Go
- 本地已有 `dubbo-go-pixiu` 源码，用于从源码启动 Pixiu
- 如果你想渲染示例配置或运行完整集成测试，需要 GNU Make 和 Bash

## 第 1 步：启动 Keycloak

```bash
cd auth/saml/docker
docker compose up -d
./docker-health-check.sh
```

启动后，Keycloak 地址为 [http://localhost:18080](http://localhost:18080)。

默认管理员账号：

- 用户名：`admin`
- 密码：`admin`

## 第 2 步：在 Keycloak 中配置 realm 和 SAML client

打开 Keycloak 管理界面后，创建以下资源。

### 创建 realm

- Realm 名称：`pixiu`

### 创建测试用户

- 用户名：`alice`
- 邮箱：`alice@example.com`
- 名：`Alice`
- 姓：`Pixiu`
- 密码：`alice123`

### 创建 SAML client

- Client 类型 / 协议：`SAML`
- Client ID：`pixiu-saml-sp`
- Name：`Pixiu SAML SP`
- Valid redirect URIs：`http://localhost:8888/*`
- Master SAML Processing URL：`http://localhost:8888/saml/acs`
- Home URL：`http://localhost:8888/app`

### 添加协议映射

为了让 Pixiu 能把用户属性转发给后端，请添加下面两个 mapper：

1. email mapper
   - Mapper type：`User Property`
   - User property：`email`
   - SAML Attribute Name：`email`

2. display name mapper
   - Mapper type：`User Property`
   - User property：`firstName`
   - SAML Attribute Name：`displayName`

完成后，Keycloak 会在下面这个地址发布 IdP metadata：

```text
http://localhost:18080/realms/pixiu/protocol/saml/descriptor
```

这个地址与 `pixiu/conf.yaml` 中配置的 `idp_metadata_url` 一致。

## 第 3 步：启动后端服务

```bash
cd auth/saml
go run server/app/*.go
```

后端会监听 `http://localhost:1314`。

## 第 4 步：渲染 Pixiu 配置

这个示例的配置里使用了 `$PROJECT_DIR` 来引用证书路径，因此要先生成
渲染后的最终配置：

```bash
cd /path/to/dubbo-go-pixiu-samples
make PROJECT_DIR=$(pwd)/auth/saml \
  PIXIU_DIR=/path/to/dubbo-go-pixiu \
  PROJECT_NAME=saml \
  BASE_DIR=$(pwd)/auth/saml/dist \
  -f igt/Makefile config
```

执行后会生成：

`auth/saml/dist/<os>_<arch>/pixiuconf/conf.yaml`

## 第 5 步：启动 Pixiu

使用上一步渲染后的配置启动 Pixiu：

```bash
cd /path/to/dubbo-go-pixiu
go run cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/auth/saml/dist/<os>_<arch>/pixiuconf/conf.yaml
```

## 第 6 步：验证示例

### 检查 SP metadata 接口

```bash
curl http://localhost:8888/saml/metadata
```

返回的 XML 中应该包含：

- `pixiu-saml-sp`
- `AssertionConsumerService`

### 检查受保护路由

浏览器打开 [http://localhost:8888/app](http://localhost:8888/app)。

预期流程如下：

1. Pixiu 把未登录请求重定向到 Keycloak
2. 使用 `alice / alice123` 登录
3. Keycloak 把 `SAMLResponse` POST 到 Pixiu 的 ACS 接口
4. Pixiu 完成校验并重定向回 `/app`
5. 后端返回类似下面的 JSON：

```json
{
  "message": "saml login success",
  "email": "alice@example.com",
  "name": "Alice"
}
```

## 第 7 步：运行 smoke test

```bash
go test -v ./auth/saml/test
```

这些测试会检查：

- 示例所需文件是否存在
- Pixiu 配置中是否包含 SAML 过滤器
- metadata 接口是否可访问
- 未登录访问 `/app` 时是否会跳转到 Keycloak

## 第 8 步：运行完整集成测试

```bash
cd /path/to/dubbo-go-pixiu-samples
./integrate_test.sh auth/saml
```

## 说明

- 这个示例面向本地 HTTP 开发环境。
- 配置里开启了 `allow_idp_initiated: true`，方便在本地 HTTP 场景下调试。
- 生产环境建议使用 HTTPS，并且除非确实需要，否则不要开启 `allow_idp_initiated`。
