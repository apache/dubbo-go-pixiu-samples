# Traffic Filter

## Canary Distribution

### Start Server

```shell
cd server/app
go run main.go
```

### Start Pixiu

```shell
go run cmd/pixiu/*.go gateway start -c samples/dubbogo/simple/traffic/pixiu/conf.yaml
```

### Test

```shell
# default (v2, canary-weight 100)
curl http://localhost:8888/user

# canary to v1
curl -H "canary-by-header: v1" http://localhost:8888/user

# canary to v3
curl -H "canary-by-header: v3" http://localhost:8888/user
```

## Header Route

### Start Server

```shell
cd server/app
go run main.go
```

### Start Pixiu

```shell
go run cmd/pixiu/*.go gateway start -c samples/dubbogo/simple/traffic/pixiu/header-conf.yaml
```

### Test

```shell
curl -H "X-A: t1" -H "X-B: t4" http://localhost:8888/user
curl -H "X-B: t4" -H "X-C: t1" http://localhost:8888/user
curl -H "REG: tt" http://localhost:8888/user
```
