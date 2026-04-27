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

## Cancellation semantics

`ctx.Done()` is a **graceful-suspend signal**, not a hard abort. When the
caller cancels `ctx`, `Run`/`RunStreamed`/`Resume`/`Restore` finish their
in-flight LLM call and tool, persist a checkpoint via the configured
`Checkpointer`, and return `*SuspendedError`. Internally `coreLoop`
shadows the incoming ctx with `context.WithoutCancel(ctx)` and uses the
shadow for every downstream call so the cancel never kills work
in-progress; only the at-boundary check observes the original ctx.

Implications:

- A `context.WithTimeout` becomes a "max wall-clock budget then suspend"
  — strictly better than today's "deadline kills work outright."
- There is no in-process hard-abort path. Callers that genuinely need
  to kill a run terminate the process; stale recovery handles the row.
- Tools that need their own deadline must derive it themselves
  (`context.WithTimeout(ctx, ...)` *inside* the tool body).

The supervisor (`pkg/probo/agent_run_handler.go`) maps a SIGTERM-driven
shutdown broadcast onto a per-run `cancelRun(ErrSuspendForCheckpoint)`,
so the same contract drives both the public Go API and the worker
infrastructure path.

## Limits

- Max turns: 10 (default)
- Max tool depth: 16 (default)
- Depth tracking prevents infinite recursion in handoffs
