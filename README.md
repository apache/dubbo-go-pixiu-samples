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
  - dubbogo/simple/farconfnacos: pixiu with nacos remote Configuration Center
  - dubbogo/simple/jaeger: pixiu with jaeger
  - dubbogo/simple/jwt: jwt authentication
  - dubbogo/simple/nacos: http to dubbo with nacos registry
  - dubbogo/simple/prometheus: pixiu with prometheus
  - dubbogo/simple/dubboproxy: dubbo to http transform and http to dubbo transform
  - dubbogo/simple/resolve: http to dubbo with auto resolve protocol
  - dubbogo/simple/traffic: traffic splitting and canary release
  - dubbogo/simple/triple: http to triple
  - dubbogo/simple/zookeeper: pixiu with dubbo using zookeeper as registry center

- dubbohttpproxy: dubbo to http transform and http to dubbo transform
- dubbotripleproxy: dubbo to triple transform and triple to dubbo transform

- grpc: grpc proxy

- http/grpc: http to grpc transform
- http/simple: simple http proxy

- llm: simple sample for pixiu to call llm

- plugins: this directory contains some plugins for pixiu
  - plugins/ratelimit: rate limit plugin for pixiu

- seata: This demonstrates how to configure the Seata filter to interact with the Seata TC for distributed transaction coordination.

- shutdown: this directory demonstrates how to gracefully shut down
  - shutdown/dubbo: This demonstrates how to gracefully shut down the Pixiu server with dubbo listener.
  - shutdown/http: This demonstrates how to gracefully shut down the Pixiu server with http listener.
  - shutdown/http2: This demonstrates how to gracefully shut down the Pixiu server with http2 listener.
  - shutdown/triple: This demonstrates how to gracefully shut down the Pixiu server with triple listener.

- springcloud: http proxy with spring cloud registry
- xds: pixiu with xds

## How To Run

Please refer [How To Run](HOWTO.md) for the instructions.

## How to contribute

If you want to add more samples, pls. read on:
1. Create new sub directory and give it an appropriate name for your new sample. Pls follow the layout of the existing sample if you are not sure how to organize your code.
2. Make sure your sample work as expected before submit PR, and make sure GitHub CI passes after PR is submitted. Pls refer to the existing sample on how to test the sample.
3. Pls provide README.md to explain your samples.