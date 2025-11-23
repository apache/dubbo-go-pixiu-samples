# Pixiu Metric Sample

This sample demonstrates how to use the new metric feature in Dubbo-Go-Pixiu to collect and expose Prometheus metrics.

## Features

The new metric feature provides:
- Automatic collection of HTTP request metrics (request count, latency, error rate, etc.)
- Metrics exposed in Prometheus format
- Integration with Grafana for visualization

## Configuration

The metric feature requires two configuration parts:

### 1. HTTP Filter Configuration

Add the `dgp.filter.http.metric` filter in `http_filters`:

```yaml
http_filters:
  - name: dgp.filter.http.metric
    config:
  # other filters...
```

### 2. Global Metric Configuration

Add a `metric` configuration block at the root level of the configuration file:

```yaml
metric:
  enable: true
  prometheus_port: 9091
```

- `enable`: Enable the metric feature
- `prometheus_port`: Port to expose Prometheus metrics

## Running the Sample

### Prerequisites

Ensure you have installed:
- Docker and Docker Compose
- Go 1.19+

### Step 1: Start Dependencies

Start Zookeeper, Prometheus, and Grafana:

```bash
cd dubbogo/simple/metric
docker-compose -f docker/docker-compose.yml up -d
```

### Step 2: Start Dubbo Service Provider

```bash
export CONF_PROVIDER_FILE_PATH=dubbogo/simple/metric/server/profiles/dev/server.yml
export APP_LOG_CONF_FILE=dubbogo/simple/metric/server/profiles/dev/log.yml

cd dubbogo/simple/metric/server/app
go run .
```

### Step 3: Start Pixiu Gateway

In a new terminal window:

```bash
cd /path/to/dubbo-go-pixiu
go run ./cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/dubbogo/simple/metric/pixiu/conf.yaml
```

### Step 4: Test the API

Send some requests to generate metric data:

```bash
# Send test request
curl -X GET "http://localhost:8888/api/v1/test-dubbo/user/tc?age=18"

# Send multiple requests to generate more data
for i in {1..10}; do
  curl -X GET "http://localhost:8888/api/v1/test-dubbo/user/user$i?age=$((20+i))"
  sleep 0.5
done
```

### Step 5: View Metrics

Access the Prometheus metrics endpoint:

```bash
curl http://localhost:9091/
```

You will see Prometheus format metrics similar to:

```
# HELP pixiu_http_requests_total Total number of HTTP requests
# TYPE pixiu_http_requests_total counter
pixiu_http_requests_total{method="GET",path="/api/v1/test-dubbo/user/tc",status="200"} 10

# HELP pixiu_http_request_duration_seconds HTTP request latencies in seconds
# TYPE pixiu_http_request_duration_seconds histogram
...
```

### Step 6: Visualize with Grafana

1. Access Grafana: http://localhost:3000
2. Default username/password: `admin` / `admin`
3. Add Prometheus data source:
   - URL: `http://prometheus:9090`
4. Create dashboards to visualize metrics

## Running Tests

```bash
cd dubbogo/simple/metric
go test -v ./test/
```

## Cleanup

```bash
docker-compose -f docker/docker-compose.yml down
```

## References

- [Dubbo-Go-Pixiu Documentation](https://github.com/apache/dubbo-go-pixiu)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)


