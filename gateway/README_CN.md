# Gateway 示例

[English](./README.md)

本示例演示了如何在 Kubernetes Gateway API 资源下使用 Apache Dubbo-go-pixiu。示例会部署 Pixiu Gateway 数据面，并将 HTTP 请求路由到 Dubbo Triple helloworld 后端服务。

示例包括：

- Pixiu Gateway 控制器、`GatewayClass` 和 `Gateway`。
- helloworld 后端 `Service` 和 `Deployment`。
- 将请求转发到后端服务的 `HTTPRoute`。
- 允许跨 namespace backend 引用的 `ReferenceGrant`。
- 作用于网关的 Pixiu `PixiuClusterPolicy` 和 `PixiuFilterPolicy` 资源。

## 前置条件

运行本示例前，请确保已经准备好：

- 一个可用的 Kubernetes 集群，例如 minikube。
- 已配置好访问该集群的 `kubectl`。
- Gateway API CRD v1.5.0 或更高版本。本示例使用
  `gateway.networking.k8s.io/v1` 的 `ReferenceGrant`。
- Pixiu 的 `PixiuClusterPolicy` 和 `PixiuFilterPolicy` CRD。

在项目根目录下安装所需 CRD：

```sh
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.5.0/standard-install.yaml
kubectl apply -k 'https://github.com/apache/dubbo-go-pixiu/controllers/config/crd?ref=develop'
```

## 如何运行

以下命令均在项目根目录 (`dubbo-go-pixiu-samples`) 下执行。

### 1. 部署 Pixiu Gateway

```sh
kubectl apply -f gateway/k8s/pixiu-gateway.yaml
```

该 manifest 会创建 Pixiu Gateway 控制器、`GatewayClass` 和 `Gateway` 资源。

### 2. 部署 Helloworld 后端服务

```sh
kubectl apply -f gateway/helloworld/deployment.yaml
```

后端服务运行在 `helloworld` namespace 中，并监听 `20000` 端口。

### 3. 应用 HTTP 路由和 Pixiu Policy

```sh
kubectl apply -f gateway/http/http.yaml
```

该 manifest 会创建 `HTTPRoute`、位于 `helloworld` namespace 的 `ReferenceGrant`，以及作用于 `Gateway default/pixiu` 的 Pixiu policy。

### 4. 检查资源

```sh
kubectl get pods -A
kubectl get httproute -n default
kubectl get referencegrant -n helloworld
```

你应该能看到 Pixiu Gateway 控制器、Pixiu 数据面和 helloworld 后端 Pod 都处于运行状态。

### 5. 发送请求

转发 Pixiu 数据面端口：

```sh
kubectl port-forward -n default pod/<pixiu-pod-name> 8888:8888
```

打开另一个终端并发送请求：

```sh
curl -v \
  -H 'Content-Type: application/json' \
  --data '{"name":"Dubbo"}' \
  http://127.0.0.1:8888/greet.GreetService/Greet
```

预期响应如下：

```json
{"greeting":"Dubbo"}
```
