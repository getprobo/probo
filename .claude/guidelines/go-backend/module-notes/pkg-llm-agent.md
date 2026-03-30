# Probo -- Go Backend -- pkg/llm and pkg/agent (LLM Integration)

> Module-specific patterns that differ from stack-wide conventions.
> For stack-wide patterns, see [patterns.md](../patterns.md) and [conventions.md](../conventions.md).

## Architecture

The LLM integration is split into three layers:

1. **pkg/llm** -- Provider-agnostic LLM abstraction (hexagonal pattern, unique in this codebase)
2. **pkg/agent** -- LLM agent orchestration framework (tool dispatch, guardrails, streaming)
3. **pkg/agents** -- Domain-specific agent implementations (changelog, vendor assessment)

## pkg/llm -- Provider Abstraction

### Hexagonal architecture (unique in codebase)

Unlike the flat pattern used elsewhere, `pkg/llm` uses a hexagonal (ports-and-adapters) architecture:

- **Core types** (root package): `Provider` interface, `ChatCompletionRequest/Response`, `Message`, `Part`, `Tool`, error types
- **Adapters** (sub-packages): `anthropic/`, `openai/`, `bedrock/` each implement `Provider`
- **Instrumented client** (root package): `Client` wraps any `Provider` with logging and OTel tracing

### Provider interface

```go
// See: pkg/llm/provider.go
type Provider interface {
    ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
    ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (ChatCompletionStream, error)
}
```

### Canonical error types

All provider adapters map vendor-specific errors to four canonical types:

| Error type | Meaning |
|-----------|---------|
| `ErrRateLimit` | API rate limit exceeded |
| `ErrContextLength` | Token limit exceeded |
| `ErrContentFilter` | Content policy violation |
| `ErrAuthentication` | Invalid API credentials |

### OTel GenAI semantic conventions

LLM calls are traced with GenAI semantic conventions (semconv v1.37.0). Spans named `chat {model}` with attributes like `gen_ai.operation.name`, `gen_ai.usage.input_tokens`, etc. No PII or message content is ever logged -- only model name, token counts, and duration.

### Streaming -- always close

Streaming calls must always call `stream.Close()` even on early exit. The OTel span is only finalized by `Next()` returning false or `Close()`. Forgetting `Close()` leaks the span.

### Provider-specific gotchas

- **Anthropic** requires `MaxTokens` to be set (returns `ErrContextLength` if nil)
- **Bedrock** does not support `ToolChoiceNone` (tools are silently omitted)
- **Bedrock** silently drops `FilePart` and `ImagePart` in user messages
- **OpenAI** is the only adapter supporting `ResponseFormat` (JSON schema mode)

## pkg/agent -- Agent Framework

### Construction via functional options

Agents are configured declaratively:

```go
// See: pkg/agent/agent.go
agent := agent.New(
    "my-agent",
    "You are a helpful assistant.",
    llmClient,
    agent.WithTools(myTool),
    agent.WithLogger(logger),
    agent.WithModel("claude-sonnet-4-20250514"),
    agent.WithInputGuardrails(promptInjectionGuard),
    agent.WithOutputGuardrails(sensitiveDataGuard),
)
```

### Default logger discards output

The agent's default logger writes to `io.Discard`. You must pass `WithLogger(myLogger)` to see any diagnostics. This is different from service packages where loggers are always wired.

### Tool interface

Tools implement a two-tier interface: `ToolDescriptor` (name + JSON schema) and `Tool` (extends with `Execute`). `FunctionTool[P]` is a generic constructor that auto-generates JSON schema from the parameter type:

```go
// See: pkg/agent/tool.go
tool := agent.FunctionTool("search", "Search the web", func(ctx context.Context, params SearchParams) (agent.ToolResult, error) {
    // ...
})
```

### Guardrails

- **Input guardrails** check messages before LLM call (e.g., `PromptInjectionGuardrail`)
- **Output guardrails** check each assistant message (e.g., `SensitiveDataGuardrail`, `SystemPromptLeakGuardrail`)
- `PromptInjectionGuardrail` **fails open** on classifier error (defense-in-depth philosophy)
- `SensitiveDataGuardrail` has broad patterns (`select `, `update `) that may cause false positives

### Approval / human-in-the-loop

Tool calls matching `ApprovalConfig` raise `InterruptedError`. Runs must be resumed via `Resume()`, not re-run via `Run()`. The `InterruptedError` carries opaque state that cannot be reconstructed.

### Streaming events

Events are sent on a buffered channel (size 64). If the consumer reads too slowly, events are **dropped silently** -- `trySendEvent` does not block.

## pkg/agents -- Domain Agents

Thin facades over `pkg/agent`:

- `GenerateChangelog(ctx, oldContent, newContent)` -- single-turn text diff summary
- `AssessVendor(ctx, websiteURL)` -- structured vendor info via `RunTyped[vendorInfo]`

Each method creates a new `agent.Agent` per call (stateless, no session). System prompts are unexported `const` strings.

**Known issue:** The logger passed to `NewAgent` is stored but never forwarded to inner `agent.New` calls -- inner agents default to `io.Discard` logging.

## Testing Pattern

LLM and agent tests use mock implementations rather than live API calls:

```go
// See: pkg/llm/llm_test.go
type mockProvider struct {
    chatCompletionFunc func(...) (*llm.ChatCompletionResponse, error)
}

// See: pkg/agent/agent_test.go
// mockProvider with pre-scripted responses
// mockChatStream for streaming tests
// recordingHook for verifying hook invocations
```

No per-provider integration tests exist. The `pkg/agents` package has no tests at all.
