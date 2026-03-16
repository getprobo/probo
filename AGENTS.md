# AGENTS.md

## Build & Development

| Command | Purpose |
|---|---|
| `make build` | Build `bin/probod` (includes frontend apps and codegen) |
| `SKIP_APPS=1 make build` | Build `bin/probod` without frontend apps (faster for backend-only work) |
| `make test` | Run tests with race detection and coverage |
| `make test MODULE=./pkg/foo` | Run tests for a single module |
| `make test-verbose` | Tests with verbose output |
| `make lint` | Vet + Go lint + npm lint |
| `make fmt` | Format Go code |
| `make dev` | Start dev server (Go + console hot-reload) |
| `make test-e2e` | Run console end-to-end tests (requires `bin/probod`) |
| `make deadcode` | Detect dead code — run after removing or renaming exported functions |
| `make stack-up` / `make stack-down` | Start / stop Docker compose infra |
| `make psql` | Open psql shell to dev database |

GraphQL and MCP codegen is triggered by `go generate`:
- `go generate ./pkg/server/api/console/v1`
- `go generate ./pkg/server/api/connect/v1`
- `go generate ./pkg/server/api/trust/v1`
- `go generate ./pkg/server/api/mcp/v1`

## Reference Documentation

Detailed guides for specific subsystems live in `contrib/claude/`:
- [`contrib/claude/relay.md`](contrib/claude/relay.md) — Frontend Relay client (queries, fragments, mutations, pagination)
- [`contrib/claude/graphql.md`](contrib/claude/graphql.md) — Go GraphQL backend (gqlgen, @goModel, connection types, cursor pagination)
- [`contrib/claude/commit.md`](contrib/claude/commit.md) — Commit message conventions
- [`contrib/claude/license.md`](contrib/claude/license.md) — ISC license header (all file types)
- [`contrib/claude/go-testing.md`](contrib/claude/go-testing.md) — Go test conventions (parallel, require vs assert, naming)
- [`contrib/claude/go-worker.md`](contrib/claude/go-worker.md) — Go worker pattern (poll-based, bounded concurrency, FOR UPDATE SKIP LOCKED)
- [`contrib/claude/go-service.md`](contrib/claude/go-service.md) — Go service orchestration (Run, graceful shutdown, crash propagation)
- [`contrib/claude/sandbox.md`](contrib/claude/sandbox.md) — Lima sandbox environments (create, manage, access services)
- [`contrib/claude/release.md`](contrib/claude/release.md) — Release process (version bump, changelog, tag, push)
- [`contrib/claude/sandbox.md`](contrib/claude/sandbox.md) — Lima sandbox environments (create, manage, access services)

## API Surface Rules

Every feature must be exposed through **all three interfaces**: GraphQL, MCP, and CLI. When adding a new endpoint or editing an existing type, keep all three in sync:

- **GraphQL** — `pkg/server/api/console/v1/schema.graphql` (+ codegen)
- **MCP** — `pkg/server/api/mcp/v1/` (+ codegen)
- **CLI** — `cmd/`

If you add a mutation in GraphQL, add the corresponding MCP tool and CLI command. If you rename or change a type, update it everywhere.

Every new Go API endpoint must have end-to-end tests in `e2e/`.

## Project

- Module: `go.probo.inc/probo`
- Router: `github.com/go-chi/chi/v5`
- Database: `go.gearno.de/kit/pg` — all raw SQL lives in `pkg/coredata`, never elsewhere
- HTTP server: `go.gearno.de/kit/httpserver`
- HTTP client: `go.gearno.de/kit/httpclient`
- Logging: `go.gearno.de/kit/log`
- Tracing: OpenTelemetry (`go.opentelemetry.io/otel`)
- UUID: `go.gearno.de/crypto/uuid` (never use `github.com/google/uuid`)
- Pointers: `go.gearno.de/x/ref` for pointer helpers (`ref.UnrefOrZero`, etc.)
- Tests: `github.com/stretchr/testify` (`require` for fatal, `assert` for non-fatal)
- Go version: 1.26 — use `new(expr)` to create pointers to values (e.g. `new(1)`, `new("foo")`, `new(time.Now())`) instead of helper functions or temporary variables

## Go Style

### Grouped declarations

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

### One argument per line

A function call is either entirely on one line or fully expanded with one argument per line. Never mix the two styles.

```go
// Good — short enough to fit on one line
id := gid.New(tenantID, "Foo")

// Good — multiple arguments, one per line
svc, err := foo.NewService(
	ctx,
	db,
	logger,
	foo.Config{
		Interval: 10 * time.Second,
		MaxRetry: 3,
	},
)

// Bad — mixed inline and multiline
svc, err := foo.NewService(ctx, db, logger, foo.Config{
	Interval: 10 * time.Second,
})
```

### Import ordering

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

### Receiver names

Short receivers: usually single-letter matching the type (`s` for Service, `c` for Client, `p` for Provider).

### Error handling

Wrap errors with `fmt.Errorf` using lowercase messages starting with `cannot`:

```go
return nil, fmt.Errorf("cannot load trust center: %w", err)
return nil, fmt.Errorf("cannot create SAML service: %w", err)
```

Sentinel errors in grouped `var ()` blocks. Custom error types implement `Unwrap() error`. Use `errors.Is` / `errors.As` for checks.

### Naming

- Constructors: `New*` (e.g. `NewService`, `NewServer`, `NewBridge`)
- Config structs: `*Config` suffix (e.g. `APIConfig`, `PgConfig`, `TrustCenterConfig`)
- Request structs: `*Request` suffix (e.g. `UpdateTrustCenterRequest`)
- Unexported types for internal data: lowercase (e.g. `vendorInfo`, `ctxKey`)

### Functional options and Config structs

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

### Interfaces

Define interfaces in the consumer package. Keep them small. Verify satisfaction at compile time:

```go
var (
	_ unit.Configurable = (*Implm)(nil)
	_ unit.Runnable     = (*Implm)(nil)
)
```

### Context

Always first parameter. Private struct keys for context values:

```go
type ctxKey struct{ name string }
var trustCenterIDKey = &ctxKey{name: "trust_center_id"}
```

### Logging

Named, context-aware structured logging with typed fields. **Never log PII, PHI, or other sensitive data** (e.g. emails, names, passwords, tokens, health records). Log opaque identifiers (IDs, request IDs) instead.

```go
l.InfoCtx(ctx, "HTTP request to trust center custom domain, redirecting to HTTPS",
	log.String("domain", domain),
	log.String("path", r.URL.Path),
	log.String("to", httpsURL),
)
```

