# This sample for OPA filter.

English | [中文](README_CN.md)

The OPA filter can provide out-of-the-box authorization capability to ensure service security and stability.

> The filter is based on [Open Policy Agent](https://www.openpolicyagent.org/), see more here [Documentation](https://www.openpolicyagent.org/docs/latest/).

### Create the API:

- Create a simple Http API, Reference the [Create a simple Http API](../../dubbogo/http/README.md), then we got a path.

- just test it: `curl http://localhost:8888/UserService -H "Test_header: 1"`

### Define Policy

- The first step, Define the policy. The OPA filter requires a Rego policy provided inline. Example:

```yaml
            policy: |
              package http.authz

              default allow = false

              allow {
                input.method == "GET"
                input.path == "/status"
              }
```

### Define Entrypoint

- The second step, Define entrypoint.The entrypoint must match the package

```yaml
       		entrypoint: "data.http.authz.allow"
```



- The third step, Ensure filter order. **The OPA filter must be placed before the HTTP proxy filter** in `http_filters`:

```yaml
filters:
  - name: dgp.filter.httpconnectionmanager
    config:
      route_config:
        # ... your routes
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
        # HTTP proxy filter should be after OPA filter
        - name: dgp.filter.http.proxy
          config:
```



### Test:

```cmd
go run /path_to/dubbo-go-pixiu/cmd/pixiu/*.go gateway start -c /path_to/dubbo-go-pixiu-samples/plugins/opa/pixiu/conf.yaml
```

```
go run /path_to/dubbo-go-pixiu-samples/plugins/opa/server/app/*
```



##### Go Test

```
go test -v /path_to/dubbo-go-pixiu-samples/plugins/opa/test
```



```
=== RUN   TestUserServiceAllow
--- PASS: TestUserServiceAllow (0.02s)
=== RUN   TestUserServiceDeny
--- PASS: TestUserServiceDeny (0.00s)
=== RUN   TestOtherServiceDeny
--- PASS: TestOtherServiceDeny (0.00s)
PASS
```



###### Use Curl to Test

- Denied request:

```bash
curl -s http://127.0.0.1:8888/OtherService
```

​	Expected result:

```
null
```



- Allowed request:

```
curl -s http://127.0.0.1:8888/UserService -H "Test_header: 1"
```

​	Expected result:

```
{"message":"UserService","result":"pass"}
```

