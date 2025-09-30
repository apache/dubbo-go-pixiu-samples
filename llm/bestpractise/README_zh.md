# **Dubbo-go-pixiu LLM 示例**

## 1. **简介**

本示例演示了如何使用 Dubbo-go-pixiu 进行 LLM 调用，并使用 Prometheus 和 Grafana 进行数据观测。

## 2. **准备工作**


1. 将你的 DeepSeek API 添加到 `.env` 文件中，更多详情请参阅 [deepseek](https://platform.deepseek.com)。

    ```shell
    cp pathto/dubbo-go-pixiu-samples/llm/bestpractise/test/.env.example pathto/dubbo-go-pixiu-samples/llm/bestpractise/test/.env
    ```
2. 根据真实情况修改 prometheus.yml 文件中的 IP 地址。

3. Docker Compose
    ```shell
    docker-compose up -d
    ```
    
### **运行 Pixiu 服务器**

通过执行以下命令运行Pixiu服务器：

```shell
cd pathto/dubbo-go-pixiu
go run ./cmd/pixiu/*.go gateway start -c pathto/dubbo-go-pixiu-samples/llm/bestpractise/pixiu/conf.yaml
```

### **运行客户端代码**

```shell
cd pathto/dubbo-go-pixiu-samples/llm/bestpractise
go run ./go-client/client.go
```

### **查看 Grafana 仪表盘**

打开浏览器，访问 `http://localhost:3000`，使用默认用户名和密码 `admin` 登录。登录后，上传 `grafana.json` 作为仪表盘，将数据源设置为 Prometheus，监控 LLM 调用的相关指标。