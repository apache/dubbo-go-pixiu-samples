# Dubbo-go-pixiu 示例

![CI](https://github.com/apache/dubbo-go-samples/workflows/CI/badge.svg)

[dubbo-go-pixiu](https://github.com/apache/dubbo-go-pixiu) 的示例

[中文 🇨🇳](./README_CN.md)

## 包含内容

- dubbogo/simple：此目录包含 dubbogo 和 pixiu 的一些简单示例
  - dubbogo/simple/bestdo：包含 jaeger 和 http 到 dubbo
  - dubbogo/simple/body：http 到 dubbo，使用 api_config.yaml
  - dubbogo/simple/csrf：CSRF 保护
  - dubbogo/simple/direct：http 到 dubbo 的直接泛化调用
  - dubbogo/simple/farconfnacos：pixiu 使用 nacos 远端配置中心
  - dubbogo/simple/jaeger：pixiu 集成 jaeger
  - dubbogo/simple/jwt：JWT 认证
  - dubbogo/simple/nacos：http 到 dubbo，使用 nacos 作为注册中心
  - dubbogo/simple/prometheus：pixiu 集成 prometheus
  - dubbogo/simple/dubboproxy：dubbo 到 http 转换和 http 到 dubbo 转换
  - dubbogo/simple/resolve：http 到 dubbo 的协议自动解析
  - dubbogo/simple/traffic：流量切分和金丝雀发布
  - dubbogo/simple/triple：http 到 triple
  - dubbogo/simple/zookeeper：pixiu 集成 dubbo，使用 zookeeper 作为注册中心

- dubbohttpproxy：dubbo 到 http 转换和 http 到 dubbo 转换
- dubbotripleproxy：dubbo 到 triple 转换和 triple 到 dubbo 转换

- grpc：gRPC 代理

- http/grpc：http 到 gRPC 转换
- http/simple：简单的 HTTP 代理

- llm：pixiu 调用 LLM 的简单示例

- plugins：此目录包含 pixiu 的一些插件
  - plugins/ratelimit：pixiu 的 ratelimit 插件

- seata：示了如何配置 Seata filter 与 Seata TC 交互对分布式事务进行协调

- shutdown：此目录演示优雅关闭
  - shutdown/dubbo：演示如何优雅关闭带有 dubbo listener 的 Pixiu 服务。
  - shutdown/http：演示如何优雅关闭带有 http listener 的 Pixiu 服务。
  - shutdown/http2：演示如何优雅关闭带有 http2 listener 的 Pixiu 服务。
  - shutdown/triple：演示如何优雅关闭带有 triple listener 的 Pixiu 服务。

- springcloud：使用 Spring Cloud 注册中心的 HTTP 代理
- xds：pixiu 集成 xds

## 如何运行

请参考 [如何运行](HOWTO_CN.md) 获取相关说明。

## 如何贡献

如果您希望增加新的用例，请继续阅读:

1. 为您的示例起合适的名字并创建子目录。如果您不太确定如何做，请参考现有示例摆放目录结构
2. 提交 PR 之前请确保在本地运行通过，提交 PR 之后请确保 GitHub 上的集成测试通过。请参考现有示例增加对应的测试
3. 请提供示例相关的 README.md 的中英文版本
