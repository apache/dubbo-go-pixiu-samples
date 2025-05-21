# 如何运行

### 目录结构

通常来说，一个示例的目录结构如下所示：

```
http/simple/
├── docker
│   └── docker-compose.yml # docker-compose 文件，用于启动 docker 服务
├── pixiu
│   └── conf.yaml          # pixiu 配置文件
├── server
│   └── app
│       └── server.go      # 服务提供方代码
│─ test
│   └── pixiu_test.go      # 测试用例
└── request.sh             # 请求脚本
```

您可以通过 bash 命令快速运行 dubbo-go-pixiu-samples 的示例以及运行集成测试。

### 通过 命令行 快速开始

*前置条件：需要 docker 环境就绪*

下面我们将使用 `http/simple` 作为示例:

1. **启动注册中心（如果有需要的话，本例不需要运行该步骤）**
   
   ```bash
   make -f igt/Makefile docker-up 
   ```
   
   当看到类似下面的输出信息时，就表明 zookeeper server 启动就绪了。
   
   ```bash
   >  Starting dependency services with ./integrate_test/dockercompose/docker-compose.yml
   Docker Compose is now in the Docker CLI, try `docker compose up`
   
   Creating network "dockercompose_default" with the default driver
   Creating dockercompose_zookeeper_1 ... done
   Creating etcd                      ... done
   Creating nacos-standalone          ... done
   ```
   
   如果要停掉注册中心，可以通过运行以下的命令完成
   
   ```bash
   make -f build/Makefile docker-down
   ```
   
2. **启动 Server**
   
    ```bash
    go run http/simple/server/app/*
    ```
 
3. **运行 Pixiu**

   在每个 sample 中，Pixiu的配置文件存储在```./pixiu/```目录下，如本例的配置文件为```http/simple/pixiu/conf.yaml```。

   如果您需要基于源码运行 Pixiu，您的工作目录需要位于 Pixiu 根目录下
   ```bash
   cd pathto/dubbo-go-pixiu/
   ```
   
   然后运行 Pixiu gateway
   ```bash
   go run cmd/pixiu/*.go gateway start -c /pathto/conf.yaml
   ```
   
   如果输出如下则 Pixiu 已被成功启动

   ```bash
   2025-05-19T12:46:00.104+0800	INFO	server/pixiu_start.go:127	[dubbopixiu go] start by config : &{StaticResources:{Listeners:[0xc0007b7a20] Clusters:[0xc0007cc5a0] Adapters:[] ShutdownConfig:0xc00067fb30 PprofConf:{Enable:false Address:{SocketAddress:{Address:0.0.0.0 Port:8881 ResolverName: Domains:[] CertsDir:} Name:}}} DynamicResources:<nil> Metric:{Enable:false PrometheusPort:0} Node:<nil> Trace:<nil> Wasm:<nil> Config:<nil> Nacos:<nil> Log:<nil>}
   2025-05-19T12:46:00.104+0800	INFO	healthcheck/healthcheck.go:157	[health check] create a health check session for 127.0.0.1:1314
   2025-05-19T12:46:00.105+0800	INFO	tracing/driver.go:76	[dubbo-go-pixiu] no trace configuration in conf.yaml
   2025-05-19T12:46:00.105+0800	INFO	http/http_listener.go:157	[dubbo-go-server] httpListener start at : 0.0.0.0:8888
   ```
   
4. **运行测试用例**

   在每个 sample 中，测试用例存储在```./test/```目录下，如本例的测试用例为```http/simple/test/pixiu_test.go```。

   运行测试用例

   ```bash
   go test -v http/simple/test/pixiu_test.go
   ```

### 运行集成测试

本项目 dubbo-go-pixiu-samples 除了用来展示如何使用 dubbo-go-pixiu 中的功能和特性之外，还被用于 apache/dubbo-go-pixiu 的集成测试。可以按照以下的步骤来运行 `http/simple` 的集成测试:
   
1. 首先确保您的机器已经下载docker
2. 运行集成测试脚本
   ```bash
   ./integrate_test.sh http/simple/
   ```
   
   当以下信息输出时，说明集成测试通过。
   ```bash
   --- PASS: TestGET1 (0.00s)
   PASS
   ok  	github.com/dubbo-go-pixiu/samples/http/simple/test	0.030s
   >  Stopping the application simple
   >  Killing PID: 12551
   ```