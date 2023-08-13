# Traffic Filter quick start

## Canary Distribution

### Start Http Server

```shell
cd server
go run server.go
```

```shell
cd server/v1
go run server.go
```

```shell
cd server/v2
go run server.go
```

### Start Pixiu

```shell
go run cmd/pixiu/*.go gateway start -c samples/dubbogo/simple/traffic/pixiu/canary-conf.yaml
```

### Start test

```shell
curl http://localhost:8888/user
curl -H "canary-by-header: v1" http://localhost:8888/user
```

## Header Route

### Start Http Server

```shell
cd server/v1
go run server.go
```

```shell
cd server/v2
go run server.go
```

```shell
cd server/v3
go run server.go
```

### Start Pixiu

```shell
go run cmd/pixiu/*.go gateway start -c samples/dubbogo/simple/traffic/pixiu/header-conf.yaml
```

### Start test

```shell
curl http://localhost:8888/user
curl -H "X-C: t1" http://localhost:8888
curl -H "REG: tt" http://localhost:8888
curl -H "X-A: t1" http://localhost:8888/user
```
