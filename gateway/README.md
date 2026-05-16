# Gateway Sample

[中文](./README_CN.md)

This sample demonstrates how to use Apache Dubbo-go-pixiu with Kubernetes
Gateway API resources. It deploys a Pixiu Gateway data plane and routes HTTP
requests to a Dubbo Triple helloworld backend.

The example includes:

- A Pixiu Gateway controller, `GatewayClass`, and `Gateway`.
- A helloworld backend `Service` and `Deployment`.
- An `HTTPRoute` that forwards requests to the backend service.
- A `ReferenceGrant` that allows the cross-namespace backend reference.
- Pixiu `PixiuClusterPolicy` and `PixiuFilterPolicy` resources for the gateway.

## Prerequisites

Before running this sample, ensure that you have:

- A running Kubernetes cluster, such as minikube.
- `kubectl` configured to access the cluster.
- Gateway API CRDs v1.5.0 or later. This sample uses
  `gateway.networking.k8s.io/v1` `ReferenceGrant`.
- Pixiu CRDs for `PixiuClusterPolicy` and `PixiuFilterPolicy`.

Install the required CRDs from the project root:

```sh
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.5.0/standard-install.yaml
kubectl apply -k 'https://github.com/apache/dubbo-go-pixiu/controllers/config/crd?ref=develop'
```

## How to Run

Run the following commands from the project root (`dubbo-go-pixiu-samples`).

### 1. Deploy Pixiu Gateway

```sh
kubectl apply -f gateway/k8s/pixiu-gateway.yaml
```

This manifest creates the Pixiu Gateway controller, `GatewayClass`, and
`Gateway` resources.

### 2. Deploy the Helloworld Backend

```sh
kubectl apply -f gateway/helloworld/deployment.yaml
```

The backend service runs in the `helloworld` namespace and listens on port
`20000`.

### 3. Apply the HTTP Route and Pixiu Policies

```sh
kubectl apply -f gateway/http/http.yaml
```

This manifest creates the `HTTPRoute`, the `ReferenceGrant` in the `helloworld`
namespace, and the Pixiu policies that target `Gateway default/pixiu`.

### 4. Check the Resources

```sh
kubectl get pods -A
kubectl get httproute -n default
kubectl get referencegrant -n helloworld
```

You should see the Pixiu Gateway controller, the Pixiu data plane, and the
helloworld backend pod running.

### 5. Send a Request

Forward the Pixiu data plane port:

```sh
kubectl port-forward -n default pod/<pixiu-pod-name> 8888:8888
```

Open another terminal and send a request:

```sh
curl -v \
  -H 'Content-Type: application/json' \
  --data '{"name":"Dubbo"}' \
  http://127.0.0.1:8888/greet.GreetService/Greet
```

The response should look like:

```json
{"greeting":"Dubbo"}
```
