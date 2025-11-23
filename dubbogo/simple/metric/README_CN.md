# Pixiu Metric 示例

本示例演示如何在 Dubbo-Go-Pixiu 中使用新的 metric 功能来收集和暴露 Prometheus 指标。

## 功能说明

新的 metric 功能提供：
- 自动收集 HTTP 请求指标（请求数、延迟、错误率等）
- 通过 Prometheus 格式暴露指标
- 支持与 Grafana 集成进行可视化监控

## 配置说明

Metric 功能需要两个配置部分：

### 1. HTTP Filter 配置

在 `http_filters` 中添加 `dgp.filter.http.metric` filter：

```yaml
http_filters:
  - name: dgp.filter.http.metric
    config:
  # 其他 filters...
```

### 2. 全局 Metric 配置

在配置文件根级别添加 `metric` 配置块：

```yaml
metric:
  enable: true
  prometheus_port: 9091
```

- `enable`: 启用 metric 功能
- `prometheus_port`: Prometheus 指标暴露的端口

## 运行示例

### 前置条件

确保已安装：
- Docker 和 Docker Compose
- Go 1.19+

### 步骤 1: 启动依赖服务

启动 Zookeeper、Prometheus 和 Grafana：

```bash
cd dubbogo/simple/metric
docker-compose -f docker/docker-compose.yml up -d
```

### 步骤 2: 启动 Dubbo 服务提供者

```bash
export CONF_PROVIDER_FILE_PATH=dubbogo/simple/metric/server/profiles/dev/server.yml
export APP_LOG_CONF_FILE=dubbogo/simple/metric/server/profiles/dev/log.yml

cd dubbogo/simple/metric/server/app
go run .
```

### 步骤 3: 启动 Pixiu 网关

在新的终端窗口中：

```bash
cd /path/to/dubbo-go-pixiu
go run ./cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/dubbogo/simple/metric/pixiu/conf.yaml
```

### 步骤 4: 测试 API

发送一些请求以生成指标数据：

```bash
# 发送测试请求
curl -X GET "http://localhost:8888/api/v1/test-dubbo/user/tc?age=18"

# 多次请求以生成更多数据
for i in {1..10}; do
  curl -X GET "http://localhost:8888/api/v1/test-dubbo/user/user$i?age=$((20+i))"
  sleep 0.5
done
```

### 步骤 5: 查看指标

访问 Prometheus 指标端点：

```bash
curl http://localhost:9091/
```

您将看到类似以下的 Prometheus 格式指标：

```
# HELP pixiu_http_requests_total Total number of HTTP requests
# TYPE pixiu_http_requests_total counter
pixiu_http_requests_total{method="GET",path="/api/v1/test-dubbo/user/tc",status="200"} 10

# HELP pixiu_http_request_duration_seconds HTTP request latencies in seconds
# TYPE pixiu_http_request_duration_seconds histogram
...
```

### 步骤 6: 使用 Grafana 可视化

1. 访问 Grafana: http://localhost:3000
2. 默认用户名/密码: `admin` / `admin`
3. 添加 Prometheus 数据源:
   - URL: `http://prometheus:9090`
4. 创建仪表板来可视化指标

## 运行测试

```bash
cd dubbogo/simple/metric
go test -v ./test/
```

## 清理环境

```bash
docker-compose -f docker/docker-compose.yml down
```

## 参考

- [Dubbo-Go-Pixiu 文档](https://github.com/apache/dubbo-go-pixiu)
- [Prometheus 文档](https://prometheus.io/docs/)
- [Grafana 文档](https://grafana.com/docs/)


