# OPA 插件示例

[English](README.md) | 中文

OPA 过滤器可以提供开箱即用的授权能力，确保服务的安全性和稳定性。

> 该过滤器基于 [Open Policy Agent](https://www.openpolicyagent.org/)，更多内容请参阅 [官方文档](https://www.openpolicyagent.org/docs/latest/)。

### 创建 API:

- 创建一个简单的 Http API，参考 [创建一个简单的Http API](../../dubbogo/http/README.md)，然后我们获得了一个可访问的路径。

- 测试命令: `curl http://localhost:8888/UserService`

### 设置过滤器

- 第一步，定义策略。OPA 过滤器需要在配置中内联 Rego 策略。例如：

```yaml
            policy: |
              package http.authz

              default allow = false

              allow {
                input.method == "GET"
                input.path == "/status"
              }
            entrypoint: "data.http.authz.allow"
```

### 设置 **entrypoint**

- 第二步，设置 entrypoint。entrypoint 必须同包匹配

```yaml
       		entrypoint: "data.http.authz.allow"
```



- 第三步，确保过滤器顺序。**OPA 过滤器必须放在 HTTP proxy 过滤器之前**，如下所示：

```yaml
filters:
  - name: dgp.filter.httpconnectionmanager
    config:
      route_config:
        # ... 你的路由
      http_filters:
        - name: dgp.filter.http.opa
          config:
            policy: |
              package http.authz

              default allow = false

              allow {
                input.method == "GET"
                input.path == "/status"
              }
            entrypoint: "data.http.authz.allow"
        # HTTP proxy 过滤器必须在 OPA 过滤器之后
        - name: dgp.filter.http.proxy
          config:
```

### 测试:

```cmd
go run /path_to/dubbo-go-pixiu/cmd/pixiu/*.go gateway start -c /path_to/dubbo-go-pixiu-samples/plugins/opa/pixiu/conf.yaml
go run /path_to/dubbo-go-pixiu-samples/plugins/opa/server/app/*
```

##### Go Test

```
go test -v /path_to/dubbo-go-pixiu-samples/plugins/opa/test
=== RUN   TestUserServiceAllow
--- PASS: TestUserServiceAllow (0.02s)
=== RUN   TestUserServiceDeny
--- PASS: TestUserServiceDeny (0.00s)
=== RUN   TestOtherServiceDeny
--- PASS: TestOtherServiceDeny (0.00s)
PASS
```

###### 使用 Curl 测试

- 拒绝请求:

```bash
curl -s http://127.0.0.1:8888/OtherService
```

​	预期输出:

```
null
```

- 允许请求:

```
curl -s http://127.0.0.1:8888/UserService -H "Test_header: 1"
```

​	预期输出:

```
{"message":"UserService","result":"pass"}
```