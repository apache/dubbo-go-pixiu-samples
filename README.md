# samples
samples for [dubbo-go-pixiu](https://github.com/apache/dubbo-go-pixiu)


## What It Contains

- dubbogo/simple/nacos: http to dubbo with nacos registry
- dubbogo/simple/triple: http to triple
- dubbogo/simple/jaeger: pixiu with jaeger
- dubbogo/simple/direct: http to dubbo with direct generic call  
- dubbogo/simple/body: http to dubbo with api_config.yaml
- dubbogo/simple/resolve: http to dubbo with auto resolve protocol

- grpc: grpc proxy
- http/grpc: http to grpc transform
- http/simple: http proxy

- springcloud: http proxy with spring cloud registry
- xds: pixiu with xds

## How to run

#### cd samples dir

```
cd dubbogo/simple
```

we can use start.sh to run samples quickly. for more info, execute command as below for more help

```
./start.sh [action] [project]
./start.sh help
```

we run body samples below step

#### prepare config file and docker

prepare command will prepare dubbo-server and pixiu config file and start docker container needed

```
./start.sh prepare body
```

if prepare config file manually, notice:
- modify $PROJECT_DIR in conf.yaml to absolute path in your compute

#### start dubbo or http server

```
./start.sh startServer body
```

#### start pixiu

```
./start.sh startPixiu body
```

if run pixiu manually in pixiu project, use command as below.

```
 go run pixiu/*.go gateway start -c /[absolute-path]/dubbo-go-pixiu/samples/dubbogo/simple/body/pixiu/conf.yaml
```


#### Try a request

use curl to send request

```bash
curl -X POST 'localhost:8881/api/v1/test-dubbo/user' -d '{"id":"0003","code":3,"name":"dubbogo","age":99}' --header 'Content-Type: application/json' 
```

or just run unit test

```bash
./start.sh startTest body

```

#### Clean

```
./start.sh clean body
```