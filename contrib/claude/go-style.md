# Go Style

## Project and dependencies

- HTTP server: `go.gearno.de/kit/httpserver`
- HTTP client: `go.gearno.de/kit/httpclient`
- Tracing: OpenTelemetry (`go.opentelemetry.io/otel`)
- Pointers: Go 1.26 — use `new(expr)` to create pointers to values (e.g. `new(1)`, `new("foo")`, `new(time.Now())`). Use `go.gearno.de/x/ref` only for dereference helpers (`ref.UnrefOrZero`, etc.)

## Grouped declarations

Use `type ()`, `const ()`, and `var ()` blocks to group related declarations. Use explicit typed values for string enums, not `iota`.

```go
type (
	CreateFooRequest struct {
		Name   string
		Active bool
	}

	UpdateFooRequest struct {
		ID     gid.GID
		Name   *string
		Active *bool
	}
)

const (
	NameMaxLength    = 100
	ContentMaxLength = 5000
)

var (
	_ Reader = (*FileReader)(nil)
	_ Writer = (*FileWriter)(nil)
)
```

## Call expressions and argument lists

In the [Go spec](https://go.dev/ref/spec#Calls), a **call** is a primary expression `f(a1, a2, … an)` where `f` is the **function value** (or **method value**) and `a1` … `an` are **arguments** passed to the matching parameters.

Treat the **argument list** as either single-line or multiline — never mixed:

- **Single-line call** — the entire call, from the callee through the closing `)`, fits on one source line. Any argument may be a short expression (including a one-line composite literal or conversion).
- **Multiline call** — if any argument is written across multiple lines (e.g. a multi-line **composite literal**, **function literal**, or other expression that contains a line break), then **every** argument must start on its own line: one argument per line at the top level of that argument list. The closing `)` is on its own line after the last argument (with a trailing comma after the final argument when the list is multiline).

Do not place some arguments on the same line as the opening `(` while others continue on following lines.

```go
// Good — entire call on one line
id := gid.New(tenantID, "Foo")

// Good — multiline argument list; each argument on its own line
svc, err := foo.NewService(
	ctx,
	db,
	logger,
	foo.Config{
		Interval: 10 * time.Second,
		MaxRetry: 3,
	},
)

// Good — function literal argument is multiline, so the name argument is on its own line too
t.Run(
	"handoff with custom tool name",
	func(t *testing.T) {
		t.Parallel()
		// ...
	},
)

// Bad — mixed: first arguments on the callee line, last argument is a multiline composite literal
svc, err := foo.NewService(ctx, db, logger, foo.Config{
	Interval: 10 * time.Second,
})
```

The same rule applies to **method calls** `x.M(a1, …)` — the receiver is already bound; the rule applies to the **argument list** after the method name.

## Import ordering

Two groups separated by a blank line: stdlib, then everything else (third-party and internal sorted together alphabetically).

```go
import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/trust"
)
```

## Receiver names

Short receivers: usually single-letter matching the type (`s` for Service, `c` for Client, `p` for Provider).

## Error handling

Wrap errors with `fmt.Errorf` using lowercase messages starting with `cannot`:

```go
return nil, fmt.Errorf("cannot load trust center: %w", err)
return nil, fmt.Errorf("cannot create SAML service: %w", err)
```

Sentinel errors in grouped `var ()` blocks. Custom error types implement `Unwrap() error`. Use `errors.Is` for sentinel checks. Use `errors.AsType[T](err)` (generic form) instead of `errors.As(err, &ptr)` for type assertions:

```go
// Good
if e, ok := errors.AsType[*ValidationError](err); ok {
	// use e
}

// Bad — avoid the two-argument form
var ve *ValidationError
if errors.As(err, &ve) {
	// use ve
}
```

## Naming

- Constructors: `New*` (e.g. `NewService`, `NewServer`, `NewBridge`)
- Config structs: `*Config` suffix (e.g. `APIConfig`, `PgConfig`, `TrustCenterConfig`)
- Request structs: `*Request` suffix (e.g. `UpdateTrustCenterRequest`)
- Unexported types for internal data: lowercase (e.g. `vendorInfo`, `ctxKey`)

## Functional options and Config structs

Use `Config` structs when a constructor has many required parameters. Use functional options (`With*` functions) for optional configuration.

```go
type Option func(*Bridge)

func WithDryRun(dryRun bool) Option {
	return func(s *Bridge) {
		s.dryRun = dryRun
	}
}

func NewBridge(provider provider.Provider, client *scimclient.Client, opts ...Option) *Bridge {
	s := &Bridge{provider: provider, scimClient: client}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
```

## Interfaces

Define interfaces in the consumer package. Keep them small. Verify satisfaction at compile time:

```go
var (
	_ unit.Configurable = (*Implm)(nil)
	_ unit.Runnable     = (*Implm)(nil)
)
```

## Context

Always first parameter. Private struct keys for context values:

```go
type ctxKey struct{ name string }
var trustCenterIDKey = &ctxKey{name: "trust_center_id"}
```

## Logging

`go.gearno.de/kit/log` — named, context-aware structured logging with typed fields. **Never log PII, PHI, or other sensitive data** (e.g. emails, names, passwords, tokens, health records). Log opaque identifiers (IDs, request IDs) instead. See [`contrib/claude/logging.md`](logging.md) for the full guide (allowed/forbidden data, field helpers, wiring patterns).

```go
l.InfoCtx(
	ctx,
	"HTTP request to trust center custom domain, redirecting to HTTPS",
	log.String("domain", domain),
	log.String("path", r.URL.Path),
	log.String("to", httpsURL),
)
```
