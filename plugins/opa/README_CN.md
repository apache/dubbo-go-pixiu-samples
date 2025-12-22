# OPA 插件示例

[English](README.md) | 中文

OPA 过滤器提供基于策略的鉴权能力。

> 过滤器基于 [Open Policy Agent](https://www.openpolicyagent.org/)，更多见官方文档：https://www.openpolicyagent.org/docs/latest/。

**推荐：** 生产环境优先使用 Server 模式（远程 OPA），嵌入模式主要用于兼容和简单演示。

### Entrypoint 提醒
- `entrypoint` 必须与策略包名/规则路径匹配（如 `data.pixiu.authz.allow`）。
- 过滤器顺序：OPA 必须放在 HTTP proxy 过滤器之前。

### 嵌入模式

- 后端：`go run /path_to/dubbo-go-pixiu-samples/plugins/opa/embedded/server/app/server.go`
- Pixiu：`go run /path_to/dubbo-go-pixiu/cmd/pixiu/*.go gateway start -c /path_to/dubbo-go-pixiu-samples/plugins/opa/embedded/pixiu/conf.yaml`
- Pixiu 关键配置（内联策略）：
  ```yaml
  http_filters:
    - name: dgp.filter.http.opa
      config:
        policy: |
          package pixiu
          default allow := false
          allow {
            input.path == "/UserService"
            input.headers["Test_header"][0] == "1"
          }
        entrypoint: data.pixiu.allow
    - name: dgp.filter.http.httpproxy
      config: {}
  clusters:
    - name: "user"
      endpoints:
        - socket_address:
            address: 127.0.0.1
            port: 1314
  ```
- Go test：
  ```bash
  cd /path_to/dubbo-go-pixiu-samples/plugins/opa/embedded/test
  go test -v .
  # 预期：
  # === RUN   TestEmbeddedUserServiceAllow
  # --- PASS: TestEmbeddedUserServiceAllow
  # === RUN   TestEmbeddedUserServiceDeny
  # --- PASS: TestEmbeddedUserServiceDeny
  # === RUN   TestEmbeddedOtherServiceDeny
  # --- PASS: TestEmbeddedOtherServiceDeny
  # PASS
  ```
- Curl（通过 Pixiu）：
  - 拒绝：`curl -s http://127.0.0.1:8888/UserService` / `curl -s http://127.0.0.1:8888/OtherService`
  - 允许：`curl -s http://127.0.0.1:8888/UserService -H "Test_header: 1"`

### Server 模式

- OPA + 上传器：
  ```bash
  cd /path_to/dubbo-go-pixiu-samples/plugins/opa/server-mode/docker
  docker-compose up
  ```
  - 宿主 8182 -> 容器 8181，`policy-uploader` 自动上传策略 `pixiu-authz`。
- 后端：`cd /path_to/dubbo-go-pixiu-samples/plugins/opa/server-mode && go run server/app/*.go`
- Pixiu：`go run /path_to/dubbo-go-pixiu/cmd/pixiu/*.go gateway start -c /path_to/dubbo-go-pixiu-samples/plugins/opa/server-mode/pixiu/conf.yaml`
- Pixiu 关键配置（远程 OPA）：
  ```yaml
  http_filters:
    - name: dgp.filter.http.opa
      config:
        server_url: "http://127.0.0.1:8182"
        decision_path: "/v1/data/pixiu/authz/allow"
        timeout_ms: 500
    - name: dgp.filter.http.httpproxy
      config: {}
  clusters:
    - name: "user"
      endpoints:
        - socket_address:
            address: 127.0.0.1
            port: 1314
  ```
- OPA 策略（compose 自动上传）：`plugins/opa/server-mode/remote/policy.rego`
  ```rego
  package pixiu.authz
  default allow := false
  allow {
    input.path == "/UserService"
    input.headers["Test_header"][0] == "1"
  }
  ```
- Go test：
  ```bash
  cd /path_to/dubbo-go-pixiu-samples/plugins/opa/server-mode/test
  go test -v .
  # 预期：
  # === RUN   TestServerModeUserServiceAllow
  # --- PASS: TestServerModeUserServiceAllow
  # === RUN   TestServerModeUserServiceDeny
  # --- PASS: TestServerModeUserServiceDeny
  # === RUN   TestServerModeOtherServiceDeny
  # --- PASS: TestServerModeOtherServiceDeny
  # PASS
  ```
- Curl（通过 Pixiu）：
  - 拒绝：`curl -s http://127.0.0.1:8888/UserService` / `curl -s http://127.0.0.1:8888/OtherService`
  - 允许：`curl -s http://127.0.0.1:8888/UserService -H "Test_header: 1"`
- Curl（直接 OPA 决策，宿主 8182）：
  - 允许：
    ```bash
    curl -s -X POST -H "Content-Type: application/json" \
      -d '{"input":{"path":"/UserService","headers":{"Test_header":["1"]}}}' \
      http://127.0.0.1:8182/v1/data/pixiu/authz/allow
    ```
  - 拒绝：
    ```bash
    curl -s -X POST -H "Content-Type: application/json" \
      -d '{"input":{"path":"/UserService","headers":{}}}' \
      http://127.0.0.1:8182/v1/data/pixiu/authz/allow
    ```
