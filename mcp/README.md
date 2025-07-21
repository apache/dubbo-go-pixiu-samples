# Dubbo-go-pixiu MCP Sample

## Introduction

This sample demonstrates how to use the MCP (Model Context Protocol) filter in Dubbo-go-pixiu to expose backend HTTP API services as MCP tools for Large Language Models (LLM) to call.

MCP is a protocol designed for LLMs to interact with external tools, implemented based on JSON-RPC 2.0. Through the MCP filter, LLMs can discover and call backend service APIs without modifying existing backend code.

## Quick Start

### 1. Start Backend Service

```shell
cd mcp/server
go run server.go
```

The backend service will start on port 8081, providing user management related APIs.

### 2. Start Pixiu Gateway

```shell
cd /path/to/dubbo-go-pixiu
go run cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/configs/pixiu-mcp-tools-test.yaml
```

The Pixiu gateway will start on port 8888, with MCP endpoint at `/mcp`.

### 3. Install and Start MCP Inspector Client

```shell
npx @modelcontextprotocol/inspector
```

Open the Inspector interface in your browser and connect to `http://localhost:8888/mcp`.

After successful connection, you can test the available tools.

## Configuration

The configuration file used in this sample is located at `configs/pixiu-mcp-tools-test.yaml`, with main configurations including:

### Server Information

```yaml
server_info:
  name: "MCP Tools Test Server"
  version: "1.0.0"
  description: "MCP Server for testing all tools functionality"
  instructions: "Use these tools to interact with the mock server API"
```

### Tool Configuration

Each tool corresponds to a backend API endpoint:

```yaml
tools:
  - name: "get_user"                    # Tool name
    description: "Get user by ID"       # Tool description
    cluster: "mock-backend"             # Target cluster
    request:
      method: "GET"                     # HTTP method
      path: "/api/users/{id}"           # API path
      timeout: "10s"                    # Timeout
    args:                               # Parameter definition
      - name: "id"
        type: "integer"
        in: "path"
        description: "User ID"
        required: true
```

### Parameter Location Description

- `path`: Path parameters, such as `id` in `/api/users/{id}`
- `query`: Query parameters, such as `?page=1&limit=10`
- `body`: Request body parameters, used for POST/PUT requests
- `header`: Request header parameters

## Backend Service

The sample backend service (`mcp/server/server.go`) provides a complete user management API, including:

- User CRUD operations
- User search functionality
- User post management
- Health check interface

After the service starts, all available API endpoints will be displayed in the console.

## Testing

This sample provides test cases to verify the correctness of MCP functionality.

### Run Tests

1. Ensure services are started (follow the steps in Quick Start section)
2. Run tests:

```shell
cd mcp/test
go test -v
```

### Test Cases

The test suite contains the following test cases:

- **TestServiceAvailability** - Check service availability
- **TestMCPInitialize** - Test MCP initialization
- **TestToolsList** - Test getting tools list
- **TestGetUser** - Test get user tool
- **TestSearchUsers** - Test search users tool
- **TestCreateUser** - Test create user tool
- **TestHealthCheck** - Test health check tool

## Troubleshooting

1. **Connection failed**: Ensure both Pixiu gateway and backend service are started
2. **Tool call failed**: Check if cluster names and paths in configuration file are correct
3. **Parameter error**: Ensure passed parameters match the type and format defined in tool definition
4. **Test failed**: Check if services are running normally and ports are not occupied

For more information, please refer to [MCP Official Documentation](https://github.com/modelcontextprotocol/specification).