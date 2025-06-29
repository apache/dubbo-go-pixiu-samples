# **Dubbo-go-pixiu Nacos 远端配置中心示例**

## 简介

本示例演示了使用 Nacos 作为 dubbo-go-pixiu 的远端配置中心的使用示例。

远端配置中心使用 Nacos，其结构图如：

![farconfnacos.png](farconfnacos.png)

## 案例运行步骤:

1. 在 docker 中启动 Nacos
    ```bash
    cd pathto/dubbo-go-pixiu-samples/dubbogo/simple/farconfnacos/docker
    docker-compose up
    ```
2. 配置Nacos作为远端配置中心
    - 在浏览器中访问```http://172.22.142.171:8848/nacos/```, 初始化登陆密码并登入
    - 修改 `dubbogo/simple/farconfnacos/pixiu/conf.yaml` 文件中的 `client-config` 部分，确保其内容与您刚才初始化的账号密码一致
      ```yaml
      client-config:
       username: {your_username}
       password: {your_password}
      ```
   - 点击 配置管理>配置列表>创建配置 来创建新配置，`Data ID` 为 `dubbo-go-pixiu`，`Group` 为 `DEFAULT_GROUP`。将`dubbogo/simple/farconfnacos/nacos/nacos.yaml`中的内容复制到配置内容对话框，点击发布

3. 设置命令行参数，启动 dubbo-go-pixiu

    ```bash
    cd pathto/dubbo-go-pixiu
    go run cmd/pixiu/*.go gateway start -c pathto/dubbo-go-pixiu-samples/dubbogo/simple/farconfnacos/pixiu/conf.yaml
    ```
   
    查看控制台输出，若看到如下日志，表示远端配置已成功加载，可以看出`idle_timeout``read_timeout``write_timeout`分别为我们在配置文件中设定的值：

    ```
    2025-05-28T11:39:22.982+0800	INFO	config/config_load.go:137	http config:map[idle_timeout:4s read_timeout:5s write_timeout:6s] true
    ```
