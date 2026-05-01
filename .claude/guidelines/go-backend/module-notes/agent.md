# Probo — Go Backend — pkg/agent (and tools sub-packages)

**Purpose.** In-house agent orchestration framework. `pkg/agent` provides
the `coreLoop` (LLM tool-call iteration), `FunctionTool` registration,
`Handoff` between agents, and `Checkpointer` (pluggable; the production
implementation is `coredata.PGCheckpointer`).
`pkg/agent/tools/{browser,search,security}` are first-party tool
implementations.

**Key files.**

- `pkg/agent/agent.go` — `Agent`, `New`, `Run`, `coreLoop`.
- `pkg/agent/tool.go` — `Tool` interface, `FunctionTool` adapter.
- `pkg/agent/handoff.go` — agent-to-agent handoff plumbing.
- `pkg/agent/checkpoint.go` — `Checkpoint` and `Checkpointer` interfaces;
  PGCheckpointer lives in `pkg/coredata/agent_run.go`.
- `pkg/agent/tools/browser/` — chromedp-backed navigation tools
  (`NavigateToURLTool`, screenshot, click, ...).
- `pkg/agent/tools/search/` — web search tool.
- `pkg/agent/tools/security/csp.go` — CSP-header validator tool.

**How to extend (a new tool).**

1. Create a struct that implements `agent.Tool` (or build a
   `FunctionTool` from a typed `func(ctx, In) (Out, error)`).
2. Register the tool with the agent at construction time.
3. If the tool dials user-influenced URLs, **use
   `httpclient.WithSSRFProtection()`** and validate the URL with
   `netcheck.ValidatePublicURL` before dispatch. Disable redirects on
   the http client to avoid TOCTOU (or rely on httpclient's per-dial
   guard).

**Top pitfalls.**

- `pkg/agent/tools/search` uses bare `http.Client` — SSRF gap. See
  [pitfalls.md § 4](../pitfalls.md).
- `pkg/agent/tools/security/csp.go` skips
  `netcheck.ValidatePublicURL`. See [pitfalls.md § 5](../pitfalls.md).
- `NavigateToURLTool` follows redirects after validation — TOCTOU. See
  [pitfalls.md § 6](../pitfalls.md).
- Long-running tool calls must respect ctx cancellation — the worker
  driving the agent already passes a deadline.
