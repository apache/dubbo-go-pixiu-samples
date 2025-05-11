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
$ go run pathto/dubbo-go-pixiu/cmd/pixiu/*.go gateway start -c pathto/dubbo-go-pixiu-samples/llm/pixiu/conf.yaml
```

### **Run the test code**

```shell
$ go test -v pathto/dubbo-go-pixiu-samples/llm/test/ -count=1
```