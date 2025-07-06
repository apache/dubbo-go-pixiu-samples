
# Pixiu Dubbo-go 快速入门示例

本文档将引导您如何运行一个基于 Pixiu 网关和 Dubbo-go 的简单服务示例。

---

### 1. 运行步骤

#### 第 1 步：进入示例目录

首先，克隆项目并进入本示例所在的目录。
```bash
# 假设您已在项目根目录
cd samples/dubbogo/simple/
````

#### 第 2 步：环境准备

执行以下脚本来启动运行所需的依赖服务（如 Zookeeper），并准备相关配置文件。
> **提示**: 请根据实际情况修改 `benchmark/pixiu/conf.yaml` 文件中的地址。

```bash
# 此命令会准备 benchmark 案例所需的环境
./start.sh prepare benchmark
```


#### 第 3 步：启动后端 Dubbo 服务

启动作为服务提供者的 Dubbo-go 应用。

```bash
./start.sh startServer benchmark
```

#### 第 4 步：启动 Pixiu 网关

在独立的终端中启动 Pixiu 网关。

```bash
./start.sh startPixiu benchmark
```

-----

### 2. 验证服务

您可以通过两种方式测试网关是否成功代理了后端服务。

#### 方式 A：使用 cURL 直接测试

打开一个新的终端，执行以下 cURL 命令：

```bash
curl -s -X GET '[http://127.0.0.1:8881/api/v1/test-dubbo/user/tc?age=66](http://127.0.0.1:8881/api/v1/test-dubbo/user/tc?age=66)'
```

**预期响应:**

您应该会看到类似下面的 JSON 输出，证明 Pixiu 已成功调用后端服务并返回结果。

```json
{
    "age": 55,
    "code": 1,
    "iD": "0001",
    "name": "tc",
    "time": "2021-08-01T18:08:41+08:00"
}
```

#### 方式 B：运行客户端测试脚本

您也可以使用预置的测试脚本来发起调用。

```bash
./start.sh startTest benchmark
```

-----

### 3. 清理环境

测试完成后，可以运行以下命令来停止所有本次示例启动的服务。

```bash
./start.sh clean
```
