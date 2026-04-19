# Agent (`pkg/agent`)

LLM agent orchestration framework.

## Agent construction

```go
agent := agent.NewAgent(
    "agent-name",
    "System instructions here",
    agent.WithTools(tool1, tool2),
    agent.WithHandoffs(otherAgent),
    agent.WithModel(model),
)
```

Functional options: `WithTools`, `WithHandoffs`, `WithInstructions`, `WithModel`, `WithModelSettings`, `WithMCPServers`, `WithInputGuardrails`, `WithOutputGuardrails`, `WithApproval`, `WithSession`.

## Execution

```go
result, err := agent.Run(ctx, messages)
result.FinalMessage().Text()  // final output
result.LastAgent              // agent that produced the result
```

Typed output via `RunTyped[T](ctx, agent, messages)` — validates against JSON Schema.

## Tool interface

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() jsonschema.Schema
    Execute(ctx context.Context, input json.RawMessage) (string, error)
}
```

## Agent-as-tool

`agent.AsTool(name, description)` wraps an agent as a tool for composition.

## Limits

- Max turns: 10 (default)
- Max tool depth: 16 (default)
- Depth tracking prevents infinite recursion in handoffs
