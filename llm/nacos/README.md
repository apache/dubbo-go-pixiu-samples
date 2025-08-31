# **Dubbo-go-pixiu LLM Sample**

## 1. **Introduction**

This sample demonstrates register service through nacos, and how to make a llm call using Dubbo-go-pixiu.

## 2. **Preparation**

1. Add your DeepSeek API to `.env` file, see [deepseek](https://platform.deepseek.com) for more details.

    ```shell
    $ cp pathto/dubbo-go-pixiu-samples/llm/nacos/.env.example pathto/dubbo-go-pixiu-samples/llm/nacos/.env
    ```
   
2. Edit the `prometheus.yml` file to set the correct IP address.

3. Docker Compose
   Service registration code and nacos service are started through Docker Compose.
    ```shell
    docker-compose up -d
    ```

### **Run the Pixiu Server**

Run the Pixiu server by executing:

```shell
$ cd pathto/dubbo-go-pixiu
$ go run ./cmd/pixiu/*.go gateway start -c pathto/dubbo-go-pixiu-samples/llm/nacos/pixiu/conf.yaml
```

### **Run the client code**

```shell
$ cd pathto/dubbo-go-pixiu-samples/llm/nacos
$ go run ./go-client/client.go
```