# **Dubbo-go-pixiu LLM 示例**

## 1. **简介**

本示例演示了如何使用 Dubbo-go-pixiu 进行 LLM 调用。

## 2. **准备工作**

```shell
$ cp pathto/dubbo-go-pixiu-samples/llm/test/.env.example pathto/dubbo-go-pixiu-samples/llm/test/.env
```

将你的 DeepSeek API 添加到 `.env` 文件中，更多详情请参阅 [deepseek](https://platform.deepseek.com)。

### **运行 Pixiu 服务器**

通过执行以下命令运行Pixiu服务器：

```shell
$ go run pathto/dubbo-go-pixiu/cmd/pixiu/*.go gateway start -c pathto/dubbo-go-pixiu-samples/llm/pixiu/conf.yaml
```

### **运行测试代码**

```shell
$ go test -v pathto/dubbo-go-pixiu-samples/llm/test/ -count=1
```