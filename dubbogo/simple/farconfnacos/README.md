环境准备:
1. 安装 docker
2. 安装 goland，会在 goland 中启动 server、pixiu、test

结构图:
见 farconfnacos.png

案例运行步骤:
1. 在 docker 中启动 zk 和 nacos. 其中 zk 作为配置中心；nacos 作为远端配置中心，存储 pixiu 的启动参数
2. 在 nacos 中创建配置，配置内容为 farconfnacos/nacos 目录中文件
3. 设置命令行参数，启动 dubbogo server:
- Working directory:
- D:\mongo\samples\dubbogo\simple\farconfnacos\server

- Environment:
- APP_LOG_CONF_FILE=./profiles/dev/log.yml;DUBBO_GO_CONFIG_PATH=./profiles/dev/server.yml

4. 设置命令行参数，启动 pixiu:
- Package path:
- github.com/apache/dubbo-go-pixiu/cmd/pixiu

- Working directory:
- D:/mongo/dubbo-go-pixiu-develop

- Program arguments:
- gateway start -c D:\mongo\dubbo-go-pixiu-samples\dubbogo\simple\farconfnacos\pixiu\conf.yaml

5. 运行 test，查看结果