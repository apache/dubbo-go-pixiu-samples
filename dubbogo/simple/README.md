# Pixiu Dubbogo 简单案例

### 运行方法

以下是运行 Pixiu Dubbogo 简单案例的步骤，以benchmark为例：

1. cd 到案例总目录
    ```bash
    cd samples/dubbogo/simple/
    ```
2. 根据实际情况修改 `dubbogo/simple/benchmark/pixiu/conf.yaml:34` 中的配置文件路径等参数。 

3. 进行环境准备，启动 zk 和准备对应配置文件
    ```bash
    ./start.sh prepare benchmark
    ```

4. 启动 dubbo server
    ```bash
    ./start.sh startServer benchmark
    ```

5. 启动 pixiu 
    ```bash
    ./start.sh startPixiu benchmark
    ```

6. 启动 Client 测试用例
    ```bash
    ./start.sh startTest benchmark
    ```

    或者使用 curl 来进行测试
    
    ```bash
    curl -s -X GET 127.0.0.1:8881/api/v1/test-dubbo/user/tc?age=66 
    ```
    
    预期返回值 ``` {"age":55,"code":1,"iD":"0001","name":"tc","time":"2021-08-01T18:08:41+08:00"} ```
