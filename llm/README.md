# **Dubbo-go-pixiu LLM Sample**

## 1. **Introduction**

This sample demonstrates how to make a llm call using Dubbo-go-pixiu.

## 2. **Preparation**

```shell
$ cp pathto/dubbo-go-pixiu-samples/llm/test/.env.example pathto/dubbo-go-pixiu-samples/llm/test/.env
```

Add your DeepSeek API to `.env` file, see [deepseek](https://platform.deepseek.com) for more details.

### **Run the Pixiu Server**

Run the Pixiu server by executing:

```shell
$ cd pathto/dubbo-go-pixiu
$ go run ./cmd/pixiu/*.go gateway start -c pathto/dubbo-go-pixiu-samples/llm/pixiu/conf.yaml
```

### **Run the client code**

```shell
$ cd pathto/dubbo-go-pixiu-samples/llm
$ go run ./go-client/client.go
```