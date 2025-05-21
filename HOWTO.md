# How to Run

[中文 🇨🇳](./HOWTO_CN.md)

### Directory Structure

Generally, the directory structure of an example is as follows:

```

http/simple/
├── docker
│   └── docker-compose.yml # docker-compose file for starting docker services
├── pixiu
│   └── conf.yaml          # Pixiu configuration file
├── server
│   └── app
│       └── server.go      # Service provider code
│─ test
│   └── pixiu_test.go     # Test cases
└── request.sh             # Request script

```

You can quickly run the examples and integration tests of dubbo-go-pixiu-samples using bash commands.

### 1. Quick Start via Command Line

*Prerequisite: Docker environment needs to be ready.*

Below, we will use the `http/simple` example:

1.  **Start the Registry Center (if needed; this step is not required for this example)**

    ```bash
    make -f igt/Makefile docker-up
    ```

    When you see output similar to the following, it indicates that the zookeeper server has started successfully.

    ```bash
    >  Starting dependency services with ./integrate_test/dockercompose/docker-compose.yml
    Docker Compose is now in the Docker CLI, try `docker compose up`

    Creating network "dockercompose_default" with the default driver
    Creating dockercompose_zookeeper_1 ... done
    Creating etcd                      ... done
    Creating nacos-standalone          ... done
    ```

    To stop the registry center, you can run the following command:

    ```bash
    make -f build/Makefile docker-down
    ```

2.  **Start the Server**

    ```bash
    go run http/simple/server/app/*
    ```

3.  **Run Pixiu**

    In each sample, Pixiu's configuration file is stored in the ```./pixiu/``` directory. For this example, the configuration file is ```http/simple/pixiu/conf.yaml```.

    If you need to run Pixiu from the source code, your working directory needs to be the Pixiu root directory.

    ```bash
    cd pathto/dubbo-go-pixiu/
    ```

    Then run the Pixiu gateway:

    ```bash
    go run cmd/pixiu/*.go gateway start -c /pathto/conf.yaml
    ```

    If the following output appears, Pixiu has been started successfully:

    ```bash
    2025-05-19T12:46:00.104+0800 INFO   server/pixiu_start.go:127  [dubbopixiu go] start by config : &{StaticResources:{Listeners:[0xc0007b7a20] Clusters:[0xc0007cc5a0] Adapters:[] ShutdownConfig:0xc00067fb30 PprofConf:{Enable:false Address:{SocketAddress:{Address:0.0.0.0 Port:8881 ResolverName: Domains:[] CertsDir:} Name:}}} DynamicResources:<nil> Metric:{Enable:false PrometheusPort:0} Node:<nil> Trace:<nil> Wasm:<nil> Config:<nil> Nacos:<nil> Log:<nil>}
    2025-05-19T12:46:00.104+0800 INFO   healthcheck/healthcheck.go:157 [health check] create a health check session for 127.0.0.1:1314
    2025-05-19T12:46:00.105+0800 INFO   tracing/driver.go:76   [dubbo-go-pixiu] no trace configuration in conf.yaml
    2025-05-19T12:46:00.105+0800 INFO   http/http_listener.go:157  [dubbo-go-server] httpListener start at : 0.0.0.0:8888
    ```

4.  **Run Test Cases**

    In each sample, test cases are stored in the ```./test/``` directory. For this example, the test case is ```http/simple/test/pixiu_test.go```.

    Run the test case:

    ```bash
    go test -v http/simple/test/pixiu_test.go
    ```

### Running Integration Tests

This project, dubbo-go-pixiu-samples, is not only used to demonstrate how to use the features and functionalities of dubbo-go-pixiu but is also used for the integration tests of apache/dubbo-go-pixiu. You can follow the steps below to run the integration test for `http/simple`:

1.  First, ensure that Docker is downloaded on your machine.
2.  Run the integration test script:

    ```bash
    ./integrate_test.sh http/simple/
    ```

    When the following information is output, it indicates that the integration test has passed:

    ```bash
    --- PASS: TestGET1 (0.00s)
    PASS
    ok   [github.com/dubbo-go-pixiu/samples/http/simple/test](https://github.com/dubbo-go-pixiu/samples/http/simple/test) 0.030s
    >  Stopping the application simple
    >  Killing PID: 12551
    ```