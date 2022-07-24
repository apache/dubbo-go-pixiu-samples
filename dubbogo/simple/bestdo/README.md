include jaeger and http to dubbo
1. cd 到案例总目录
cd samples/dubbogo/simple/

2. 进行环境准备，启动 zk 和准备对应配置文件
./start.sh prepare bestdo

3. 启动 dubbo server
./start.sh startServer bestdo

4. 启动 pixiu 
./start.sh startPixiu bestdo

5. 启动 Client 测试用例
./start.sh startTest bestdo

或者使用 curl 

``` curl -X POST 'http://localhost:8883/UserService/com.dubbogo.pixiu.UserService/GetUserByName' -d '{"types":"string","values":"tc"}' -H 'Content-Type: application/json' -H 'x-dubbo-http1.1-dubbo-version':'1.0.0' -H 'x-dubbo-service-protocol':"dubbo" -H 'x-dubbo-service-version':'1.0.0' -H 'x-dubbo-service-group':'test'```

返回值 ``` {"age":15,"code":1,"iD":"0001","name":"tc","time":"2021-08-01T18:08:41+08:00"} ```
