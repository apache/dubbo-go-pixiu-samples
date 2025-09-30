# **Dubbo-go-pixiu LLM Sample**

## 1. **Introduction**

This sample demonstrates how to make a llm call using Dubbo-go-pixiu, and use prometheus and grafana to discover metrics.

## 2. **Preparation**

1. Add your DeepSeek API to `.env` file, see [deepseek](https://platform.deepseek.com) for more details.

    ```shell
    cp pathto/dubbo-go-pixiu-samples/llm/bestpractise/test/.env.example pathto/dubbo-go-pixiu-samples/llm/bestpractise/test/.env
    ```
   
2. Edit the `prometheus.yml` file to set the correct IP address.

3. Docker Compose
    ```shell
    docker-compose up -d
    ```

### **Run the Pixiu Server**

Run the Pixiu server by executing:

```shell
cd pathto/dubbo-go-pixiu
go run ./cmd/pixiu/*.go gateway start -c pathto/dubbo-go-pixiu-samples/llm/bestpractise/pixiu/conf.yaml
```

### **Run the client code**

```shell
cd pathto/dubbo-go-pixiu-samples/llm/bestpractise
go run ./go-client/client.go
```

### **View Grafana Dashboard**

Open your browser and go to `http://localhost:3000`, log in with the default username and password `admin`. After logging in, upload `grafana.json` as a dashboard, set the data source to Prometheus, and monitor the relevant metrics of LLM calls.