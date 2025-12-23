# Sample for OPA filter

English | [中文](README_CN.md)

The OPA filter can provide out-of-the-box authorization capability to ensure service security and stability.

> The filter is based on [Open Policy Agent](https://www.openpolicyagent.org/), see more here [Documentation](https://www.openpolicyagent.org/docs/latest/).

**Recommendation:** Use server mode (remote OPA) for production; embedded mode is for compatibility and simple demos.

### Entrypoint Reminder
- `entrypoint` must match your policy package/rule path (e.g., `data.pixiu.authz.allow`).
- Filter order: OPA must be before the HTTP proxy filter.

### Embedded Mode

- Backend: `go run /path_to/dubbo-go-pixiu-samples/plugins/opa/embedded/server/app/server.go`
- Pixiu: `go run /path_to/dubbo-go-pixiu/cmd/pixiu/*.go gateway start -c /path_to/dubbo-go-pixiu-samples/plugins/opa/embedded/pixiu/conf.yaml`
- Pixiu key config (inline policy):
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
- Go test:
  ```bash
  cd /path_to/dubbo-go-pixiu-samples/plugins/opa/embedded/test
  go test -v .
  # Expected:
  # === RUN   TestEmbeddedUserServiceAllow
  # --- PASS: TestEmbeddedUserServiceAllow
  # === RUN   TestEmbeddedUserServiceDeny
  # --- PASS: TestEmbeddedUserServiceDeny
  # === RUN   TestEmbeddedOtherServiceDeny
  # --- PASS: TestEmbeddedOtherServiceDeny
  # PASS
  ```
- Curl (through Pixiu):
  - Deny: `curl -s http://127.0.0.1:8888/UserService` / `curl -s http://127.0.0.1:8888/OtherService`
  - Allow: `curl -s http://127.0.0.1:8888/UserService -H "Test_header: 1"`

### Server Mode

- OPA + uploader:
  ```bash
  cd /path_to/dubbo-go-pixiu-samples/plugins/opa/server-mode/docker
  docker-compose up
  ```
  - Host port 8182 -> container 8181, policy auto uploaded to `pixiu-authz`.
- Backend: `cd /path_to/dubbo-go-pixiu-samples/plugins/opa/server-mode && go run server/app/*.go`
- Pixiu: `go run /path_to/dubbo-go-pixiu/cmd/pixiu/*.go gateway start -c /path_to/dubbo-go-pixiu-samples/plugins/opa/server-mode/pixiu/conf.yaml`
- Pixiu key config (remote OPA):
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
- OPA policy (auto uploaded by compose): `plugins/opa/server-mode/remote/policy.rego`
  ```rego
  package pixiu.authz
  default allow := false
  allow {
    input.path == "/UserService"
    input.headers["Test_header"][0] == "1"
  }
  ```
- Go test:
  ```bash
  cd /path_to/dubbo-go-pixiu-samples/plugins/opa/server-mode/test
  go test -v .
  # Expected:
  # === RUN   TestServerModeUserServiceAllow
  # --- PASS: TestServerModeUserServiceAllow
  # === RUN   TestServerModeUserServiceDeny
  # --- PASS: TestServerModeUserServiceDeny
  # === RUN   TestServerModeOtherServiceDeny
  # --- PASS: TestServerModeOtherServiceDeny
  # PASS
  ```
- Curl (through Pixiu):
  - Deny: `curl -s http://127.0.0.1:8888/UserService` / `curl -s http://127.0.0.1:8888/OtherService`
  - Allow: `curl -s http://127.0.0.1:8888/UserService -H "Test_header: 1"`
- Curl (direct OPA decision, host port 8182):
  - Allow:
    ```bash
    curl -s -X POST -H "Content-Type: application/json" \
      -d '{"input":{"path":"/UserService","headers":{"Test_header":["1"]}}}' \
      http://127.0.0.1:8182/v1/data/pixiu/authz/allow
    ```
  - Deny:
    ```bash
    curl -s -X POST -H "Content-Type: application/json" \
      -d '{"input":{"path":"/UserService","headers":{}}}' \
      http://127.0.0.1:8182/v1/data/pixiu/authz/allow
    ```
