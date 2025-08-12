# **Dubbo-go-pixiu LLM 示例**

## 1. **简介**

本示例演示了如何使用 Dubbo-go-pixiu 进行 LLM 调用。

## 2. **准备工作**


1. 将你的 DeepSeek API 添加到 `.env` 文件中，更多详情请参阅 [deepseek](https://platform.deepseek.com)。

    ```shell
    $ cp pathto/dubbo-go-pixiu-samples/llm/test/.env.example pathto/dubbo-go-pixiu-samples/llm/test/.env
    ```

2. Docker Compose
    ```shell
    docker-compose up -d
    ```
    
### **运行 Pixiu 服务器**

通过执行以下命令运行Pixiu服务器：

```shell
$ cd pathto/dubbo-go-pixiu
$ go run ./cmd/pixiu/*.go gateway start -c pathto/dubbo-go-pixiu-samples/llm/pixiu/conf.yaml
```

### **运行客户端代码**

```shell
$ cd pathto/dubbo-go-pixiu-samples/llm
$ go run ./go-client/client.go
```