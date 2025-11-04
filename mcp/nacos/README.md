# Dubbo-go-pixiu MCP with Nacos Example

English | [中文](./README_zh.md)

This example demonstrates how to use Nacos 3.0+ as the MCP (Model Configuration Protocol) server for the Dubbo-go-pixiu gateway.

## Prerequisites

- Go programming environment
- Nacos 3.0 or higher installed and running

## Steps

### 1. Configure Nacos MCP Service

First, we need to configure and publish an MCP service in the Nacos console.

1.  **Login to Nacos Console**: Access `http://<nacos-server-ip>:8080/nacos`.
2.  **Navigate to MCP Management**: Find and click on "MCP Management" in the left sidebar.
3.  **Create MCP Server**:
    *   Click "MCP List" -> "Create MCP Server".
    *   **Type**: Select `streamable`.
    *   **Tools**: Select "Import from OpenAPI", then upload the `mcp/nacos/mcptools/mcptools.yaml` file.
4.  **Correct Backend Address (Important)**:
    *   After a successful upload, Nacos will automatically parse `mcptools.yaml` and generate a list of tools.
    *   **Note**: There is a known issue in Nacos 3.0 where the imported backend address `http://` might be incorrectly changed to `http:/`. Please manually check and correct the backend address for all tools to ensure it is `http://localhost:8081`.
5.  **Publish the Service**:
    *   After confirming all configurations are correct, click "Publish" to start the MCP Server.
    *   *Note: Pixiu currently only supports connecting to a single MCP Server.*

### 2. Start the Mock Backend Server

This server provides the API interfaces defined in the OpenAPI file.

```bash
cd mcp/simple/server/app
go run .
```

After starting successfully, you will see output similar to the following:

```
🚀 Mock Backend Server starting on :8081
📚 Available endpoints:
  GET  /api/users/{id}        - Get user by ID
  GET  /api/users/search      - Search users
  POST /api/users             - Create user
  GET  /api/users/{id}/posts  - Get user posts
  GET  /api/health            - Health check
  GET  /                      - Root endpoint
```
*Please execute the next step in a new terminal window, keeping this server running.*

### 3. Start the Pixiu Gateway

Now, start the Pixiu gateway, which will connect to the Nacos MCP service and route requests based on the fetched configuration.

```shell
cd /path/to/dubbo-go-pixiu
go run cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/mcp/nacos/pixiu/conf.yaml
```

### 4. Install and Launch the MCP Inspector Client

```shell
npx @modelcontextprotocol/inspector
```

Open the Inspector interface in your browser and connect to `http://localhost:8888/mcp` to start testing.
