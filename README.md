# Dubbo-go-pixiu Samples

![CI](https://github.com/apache/dubbo-go-samples/workflows/CI/badge.svg)

samples for [dubbo-go-pixiu](https://github.com/apache/dubbo-go-pixiu)

[中文 🇨🇳](./README_CN.md)

## What It Contains

- dubbogo/simple: this directory contains some simple samples for dubbogo and pixiu
  - dubbogo/simple/bestdo: include jaeger and http to dubbo
  - dubbogo/simple/body: http to dubbo with api_config.yaml
  - dubbogo/simple/csrf: csrf protection
  - dubbogo/simple/direct: http to dubbo with direct generic call
  - dubbogo/simple/nacos_farconf: pixiu uses nacos as a remote config center
  - dubbogo/simple/jaeger: pixiu with jaeger
  - dubbogo/simple/jwt: jwt authentication
  - dubbogo/simple/nacos: http to dubbo with nacos registry
  - dubbogo/simple/prometheus: pixiu with prometheus
  - dubbogo/simple/dubboproxy: dubbo to http transform and http to dubbo transform
  - dubbogo/simple/resolve: convert http requests to dubbo requests, following the default http to dubbo conversion rules
  - dubbogo/simple/traffic: traffic splitting and canary release
  - dubbogo/simple/triple: http to triple
  - dubbogo/simple/zookeeper: pixiu with dubbo using zookeeper as registry center

- dubbohttpproxy: dubbo to http transform and http to dubbo transform
- dubbotripleproxy: example of inter-conversion of dubbo and triple protocol requests

- grpc/simple: this sample demonstrates how to use pixiu as a gateway for a standard gRPC service, supporting unary, client-side streaming, server-side streaming, and bidirectional streaming RPCs.

- http/grpc: convert http requests to grpc requests, support configuring proto files or dynamically obtaining proto information from a grpc server with reflection enabled.
- http/simple: this directory contains common Http request proxying features that serve as common API gateways

- llm: simple sample for pixiu to call llm

- mcp: demonstrates MCP (Model Context Protocol) filter to expose HTTP APIs as LLM tools
  - mcp/simple: basic MCP service integration example showing how to convert HTTP APIs to MCP tools
  - mcp/oauth: MCP OAuth authorization example demonstrating how to protect MCP endpoints with OAuth2, supporting PKCE authorization code flow

- plugins: this directory contains some plugins for pixiu
  - plugins/ratelimit: rate limit plugin for pixiu

- seata: this demonstrates how to configure the Seata filter to interact with the Seata TC for distributed transaction coordination.

- shutdown: this directory demonstrates how to gracefully shut down
  - shutdown/dubbo: this demonstrates how to gracefully shut down the Pixiu server with dubbo listener.
  - shutdown/http: this demonstrates how to gracefully shut down the Pixiu server with http listener.
  - shutdown/http2: this demonstrates how to gracefully shut down the Pixiu server with http2 listener.
  - shutdown/triple: this demonstrates how to gracefully shut down the Pixiu server with triple listener.

- springcloud: Http proxy function, get cluster information from spring cloud service registry, dynamic management of cluster and route function
- xds: pixiu with xds

## How To Run

Please refer [How To Run](HOWTO.md) for the instructions.

## Dubbo-go-pixiu ecosystem other projects

- [pixiu-admin](https://github.com/dubbo-go-pixiu/pixiu-admin) Dubbo-go-pixiu Admin is a comprehensive management platform for the dubbo-go-pixiu Gateway. It provides a centralized control panel for configuring, monitoring, and managing gateway resources through both a web-based user interface and RESTful APIs.
- [pixiu-api](https://github.com/dubbo-go-pixiu/pixiu-api) Dubbo-go-pixiu API is a gateway framework that serves as the API configuration model for the dubbo-go-pixiu ecosystem. It provides a set of APIs and configuration models to build, configure, and extend API gateway functionality for microservices, with a particular focus on integration with Dubbo services.
- [benchmark](https://github.com/dubbo-go-pixiu/benchmark) The benchmark system allows users to measure and analyze key performance metrics such as latency, throughput, and Queries Per Second (QPS) under various load conditions to evaluate the efficiency of the protocol conversion process.

## How to contribute

If you want to add more samples, pls. read on:
1. Create new sub directory and give it an appropriate name for your new sample. Pls follow the layout of the existing sample if you are not sure how to organize your code.
2. Make sure your sample work as expected before submit PR, and make sure GitHub CI passes after PR is submitted. Pls refer to the existing sample on how to test the sample.
3. Pls provide README.md to explain your samples.