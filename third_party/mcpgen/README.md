# mcpgen

## Overview

mcpgen is a code generator for Model Context Protocol (MCP) servers in Go, inspired by [gqlgen](https://github.com/99designs/gqlgen).

mcpgen takes a schema-first approach to building MCP servers. Define your tools, resources, and prompts in a YAML configuration file with JSON Schema definitions, and mcpgen generates type-safe Go code including:

- Type-safe Go structs from JSON Schemas
- MCP server boilerplate with the official [go-sdk](https://github.com/modelcontextprotocol/go-sdk)
- Handler function stubs ready for your business logic

## Features

- **Schema-First Development**: Define MCP primitives (tools, resources, prompts) in YAML with JSON Schema
- **Type-Safe Code Generation**: Generate Go structs from JSON Schema Draft 2020-12
- **Custom Type Mapping**: Use your own Go types instead of generated ones (like gqlgen)
- **Omittable Fields**: Distinguish between "not set", "null", and "value" with `go.probo.inc/mcpgen/omittable` (like gqlgen's `@goField(omittable: true)`)
- **Official SDK Integration**: Uses the official `modelcontextprotocol/go-sdk`
- **Handler Preservation**: Regeneration preserves your handler implementations
- **gqlgen-Inspired**: Familiar workflow if you've used gqlgen

## Installation

```bash
go install go.probo.inc/mcpgen@latest
```

Or build from source:

```bash
git clone https://github.com/probo-inc/mcpgen
cd mcpgen
go build -o mcpgen
```

## Quick Start

### 1. Initialize a new project

```bash
mcpgen init my-mcp-server
cd my-mcp-server
```

This creates:
```
my-mcp-server/
├── mcpgen.yaml           # Configuration file
├── schemas/              # JSON Schema definitions
│   └── example_input.json
├── main.go               # Entry point
└── README.md
```

### 2. Define your MCP primitives

Edit `mcpgen.yaml`:

```yaml
server:
  name: my-mcp-server
  version: 1.0.0

tools:
  - name: calculate
    description: Perform arithmetic operations
    input_schema: schemas/calculate_input.json

resources:
  - uri: docs://readme
    name: Project README
    description: The project README file
    mime_type: text/markdown

prompts:
  - name: greeting
    description: A friendly greeting
    arguments:
      - name: name
        description: Name of person to greet
        required: false
```

### 3. Create JSON Schemas

Define schemas in the `schemas/` directory:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "operation": {
      "type": "string",
      "enum": ["add", "subtract", "multiply", "divide"]
    },
    "a": {
      "type": "number",
      "description": "First operand"
    },
    "b": {
      "type": "number",
      "description": "Second operand"
    }
  },
  "required": ["operation", "a", "b"]
}
```

### 4. Generate code

```bash
mcpgen generate
```

This generates:
- `generated/models.go` - Type-safe Go structs
- `generated/server.go` - MCP server setup
- `generated/resolver.go` - Handler stubs (first time only)

### 5. Implement handlers

Edit `generated/resolver.go`:

```go
func (r *Resolver) Calculate(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
    operation := args["operation"].(string)
    a := args["a"].(float64)
    b := args["b"].(float64)

    var result float64
    switch operation {
    case "add":
        result = a + b
    case "subtract":
        result = a - b
    case "multiply":
        result = a * b
    case "divide":
        if b == 0 {
            return nil, nil, fmt.Errorf("division by zero")
        }
        result = a / b
    }

    return &mcp.CallToolResult{
        Content: []mcp.Content{
            &mcp.TextContent{
                Text: fmt.Sprintf("Result: %f", result),
            },
        },
    }, map[string]any{"result": result}, nil
}
```

### 6. Build and run

```bash
go mod init my-mcp-server
go mod tidy
go build -o server
./server
```

## Configuration Reference

### Server Configuration

```yaml
server:
  name: my-server       # Required: Server name
  version: 1.0.0        # Required: Server version
```

### Code Generation Options

```yaml
exec:
  filename: generated/server.go  # Server code output
  package: generated             # Package name

model:
  filename: generated/models.go  # Models output
  package: generated             # Package name

resolver:
  filename: generated/resolver.go  # Resolver stubs output
  type: Resolver                   # Resolver type name
  package: generated               # Package name
  preserve_resolver: true          # Don't overwrite on regeneration
```

### Tools

```yaml
tools:
  - name: tool_name                      # Required: Tool identifier
    description: Tool description        # Optional: Human-readable description
    input_schema: schemas/input.json     # Required: JSON Schema for input
    output_schema: schemas/output.json   # Optional: JSON Schema for output
```

### Resources

Static resources:

```yaml
resources:
  - uri: docs://readme               # Required: Resource URI
    name: README                     # Required: Display name
    description: Project README      # Optional
    mime_type: text/markdown        # Optional
```

Resource templates (dynamic URIs):

```yaml
resources:
  - uri_template: users://{id}/profile  # Required: URI template
    name: User Profile                  # Required
    description: User profile data      # Optional
    mime_type: application/json        # Optional
    uri_params:                        # Parameters from template
      - name: id
        type: string
        description: User ID
```

### Prompts

```yaml
prompts:
  - name: prompt_name           # Required: Prompt identifier
    description: Description    # Optional
    arguments:                  # Optional: Prompt arguments
      - name: arg_name
        description: Arg description
        required: true
```

## Commands

### `mcpgen init [name]`

Initialize a new MCP server project with example configuration.

```bash
mcpgen init my-server
```

### `mcpgen generate`

Generate code from `mcpgen.yaml` configuration.

```bash
mcpgen generate

# Specify custom config file
mcpgen generate --config custom-config.yaml
```

### `mcpgen version`

Print mcpgen version.

```bash
mcpgen version
```

## How It Works

1. **Configuration Loading**: mcpgen reads your `mcpgen.yaml` file
2. **Schema Loading**: JSON Schemas are loaded and `$ref` references resolved
3. **Type Generation**: Go structs are generated from JSON Schemas
4. **Server Generation**: MCP server boilerplate is generated with tool/resource/prompt registration
5. **Resolver Generation**: Handler stubs are generated (only if they don't exist)

## MCP Primitives

### Tools

Tools let LLMs interact with external systems. Each tool has:
- **Name**: Unique identifier (alphanumeric, underscore, dash, dot)
- **Description**: What the tool does
- **Input Schema**: JSON Schema defining parameters (required)
- **Output Schema**: JSON Schema for result validation (optional)

### Resources

Resources provide context to LLMs via URIs:
- **Static Resources**: Fixed URI (e.g., `docs://readme`)
- **Resource Templates**: Dynamic URIs (e.g., `users://{id}/profile`)

### Prompts

Prompts are reusable templates for LLM interactions with optional arguments.

## Comparison with gqlgen

| Feature | gqlgen | mcpgen |
|---------|--------|--------|
| **Schema Language** | GraphQL SDL | JSON Schema |
| **Protocol** | GraphQL | MCP (JSON-RPC 2.0) |
| **Core Primitives** | Queries, Mutations, Subscriptions | Tools, Resources, Prompts |
| **Generation** | Resolvers, models | Handlers, models |
| **Schema-first** | ✅ | ✅ |
| **Preserve implementations** | ✅ | ✅ |
| **Type safety** | ✅ | ✅ |

## Custom Type Mapping

You can use your own Go types instead of generated ones, similar to gqlgen's model binding.

### Using Schema Annotations (Recommended)

Add `go.probo.inc/mcpgen/type` annotations in your JSON Schema:

```yaml
components:
  schemas:
    # Use time.Time for timestamps
    Timestamp:
      type: string
      format: date-time
      go.probo.inc/mcpgen/type: time.Time

    # Use UUID package
    UUID:
      type: string
      format: uuid
      go.probo.inc/mcpgen/type: github.com/google/uuid.UUID

    # Use your own domain models
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
      go.probo.inc/mcpgen/type: github.com/myorg/models.User
```

When you reference these schemas, mcpgen will:
- Skip generating types for them
- Use your custom types instead
- Automatically add necessary imports

See [docs/custom-types.md](docs/custom-types.md) for full documentation.

## Examples

See the `examples/` directory for complete working examples.

## Development

### Building

```bash
go build -o mcpgen
```

### Testing

```bash
go test ./...
```

## Contributing

Contributions welcome. Please submit a Pull Request.

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Inspired by [gqlgen](https://github.com/99designs/gqlgen)
- Uses the official [Model Context Protocol Go SDK](https://github.com/modelcontextprotocol/go-sdk)
