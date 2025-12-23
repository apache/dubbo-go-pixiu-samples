# Dubbo-Go-Pixiu-Samples

![CI](https://github.com/apache/dubbo-go-samples/workflows/CI/badge.svg)
![License](https://img.shields.io/badge/license-Apache--2.0-green.svg)

Examples for [dubbo-go-pixiu](https://github.com/apache/dubbo-go-pixiu)

**English** | [中文](README_CN.md)

---

**Dubbo-Go-Pixiu-Samples** is a collection of examples built on **Dubbo-Go-Pixiu**, demonstrating how to use Dubbo-Go-Pixiu as an API gateway to handle various protocol conversions and microservice integration scenarios.
This project includes multiple samples covering conversions such as HTTP to Dubbo, gRPC to HTTP, Dubbo to HTTP, and integrations with common microservice components like Jaeger, Prometheus, and Nacos.

👉 **Quick Start:** Want to get hands-on quickly? Check out our [HOWTO Guide](HOWTO.md) to learn how to run the sample code.

## Contents

* **dubbogo/simple**: Basic examples of dubbogo and pixiu

  * `bestdo`: Includes Jaeger and HTTP-to-Dubbo conversion
  * `body`: HTTP to Dubbo using `api_config.yaml`
  * `csrf`: CSRF protection
  * `direct`: Direct generic HTTP-to-Dubbo call
  * `nacos_farconf`: Pixiu using Nacos as a remote configuration center
  * `jaeger`: Pixiu integration with Jaeger
  * `jwt`: JWT authentication
  * `nacos`: HTTP to Dubbo using Nacos as a registry center
  * `prometheus`: Pixiu integration with Prometheus
  * `dubboproxy`: Dubbo-to-HTTP and HTTP-to-Dubbo conversion
  * `resolve`: Converts HTTP requests to Dubbo requests following default conversion rules
  * `traffic`: Traffic splitting and canary release
  * `triple`: HTTP to Triple protocol
  * `zookeeper`: Pixiu integration with Dubbo using Zookeeper as the registry center

* **dubbohttpproxy**: Dubbo-to-HTTP and HTTP-to-Dubbo conversion

* **dubbotripleproxy**: Conversion between Dubbo2 and Triple protocol requests

* **grpc/simple**: Demonstrates how to use Pixiu as a gateway for standard gRPC services, supporting unary calls, client streaming, server streaming, and bidirectional streaming.

* **http/grpc**: Converts HTTP requests to gRPC requests, supporting configuration via proto files or dynamic retrieval from a gRPC server with reflection enabled.

* **http/simple**: Common HTTP proxy examples demonstrating typical API gateway functionality.

* **llm**: Examples for pixiu-ai-gateway

  * `bestpractice`: Shows how to use pixiu-ai-gateway as a unified LLM entry point, supporting model fallback, retry on failure, and Prometheus + Grafana monitoring.
  * `nacos`: Demonstrates using Nacos as the service registry for pixiu-ai-gateway LLM services.

* **mcp**: Demonstrates the MCP (Model Context Protocol) filter that exposes HTTP APIs as LLM tools.

  * `simple`: Basic MCP service integration example showing how to convert HTTP APIs into MCP tools.
  * `oauth`: MCP OAuth authorization example demonstrating OAuth2 protection for MCP endpoints, supporting the PKCE authorization code flow.
  * `nacos`: MCP Nacos integration example showing how to use Nacos as both the registry and configuration center for MCP servers.

* **plugins**: Pixiu plugin examples

  * `ratelimit`: Pixiu rate limiting plugin
  * `opa`: Pixiu Open Policy Agent (OPA) integration example for policy-based access control (embedded Rego sample and server-mode sample)

* **seata**: Demonstrates how to configure the Seata filter to interact with Seata TC for distributed transaction coordination

* **shutdown**: Demonstrates graceful shutdown

  * `dubbo`: Graceful shutdown for Pixiu services with Dubbo listener
  * `http`: Graceful shutdown for Pixiu services with HTTP listener
  * `http2`: Graceful shutdown for Pixiu services with HTTP/2 listener
  * `triple`: Graceful shutdown for Pixiu services with Triple listener

* **springcloud**: HTTP proxy using Spring Cloud service registry for dynamic cluster and route management

* **tools**: Development and testing utilities

  * `authserver`: OAuth2 authorization server implementation providing full authorization code flow with PKCE, JWT token generation, and validation

* **xds**: Pixiu integration with xDS

## Other Projects in the Dubbo-Go-Pixiu Ecosystem

* **[pixiu-admin](https://github.com/apache/dubbo-go-pixiu/tree/develop/admin)**
  Dubbo-Go-Pixiu Admin is a comprehensive management platform for the Dubbo-Go-Pixiu gateway. It provides a centralized control plane for configuring, monitoring, and managing gateway resources via a web-based UI and RESTful APIs.

* **[pixiu-api](https://github.com/dubbo-go-pixiu/pixiu-api)**
  Dubbo-Go-Pixiu API provides the API models for the ecosystem and integrates with Pixiu Admin.

* **[benchmark](https://github.com/apache/dubbo-go-pixiu/tree/develop/tools/benchmark)**
  The benchmarking system allows users to measure and analyze key performance metrics—such as latency, throughput, and QPS—under various load conditions to evaluate protocol conversion efficiency.

## Contributing

If you’d like to add new examples, please follow these steps:

1. Choose a proper name for your example and create a subdirectory. Refer to existing examples for directory structure guidance.
2. Ensure all examples run successfully locally before submitting your PR, and confirm CI tests pass on GitHub.
3. Provide both English and Chinese versions of your example’s README.md file.

## License

This project is licensed under the [Apache License 2.0](LICENSE).