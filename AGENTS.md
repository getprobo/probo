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
- [`contrib/claude/relay.md`](contrib/claude/relay.md) — Relay cursor pagination (cursor format, keyset pagination, schema types)
- [`contrib/claude/graphql.md`](contrib/claude/graphql.md) — Frontend Relay client (queries, fragments, mutations, pagination)
- [`contrib/claude/commit.md`](contrib/claude/commit.md) — Commit message conventions
- [`contrib/claude/license.md`](contrib/claude/license.md) — ISC license header for Go files

## API Surface Rules

Every feature must be exposed through **all three interfaces**: GraphQL, MCP, and CLI. When adding a new endpoint or editing an existing type, keep all three in sync:

- **GraphQL** — `pkg/server/api/console/v1/schema.graphql` (+ codegen)
- **MCP** — `pkg/server/api/mcp/v1/` (+ codegen)
- **CLI** — `cmd/`

If you add a mutation in GraphQL, add the corresponding MCP tool and CLI command. If you rename or change a type, update it everywhere.

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

### Tests

Black-box test packages (`package foo_test`). White-box (`package foo`) only when testing unexported functions. Test names follow `TestFunctionName_Scenario`. Use `t.Run` for subtests with lowercase descriptive names. Table-driven tests for parameterized cases.

**Parallel tests:** Always call `t.Parallel()` at both the top-level test and each subtest.

**`require` vs `assert`:**
- `require` — stops the test immediately on failure. Use for **preconditions** that would make subsequent assertions meaningless: `require.NoError`, `require.Error`, `require.ErrorAs`, `require.Len`, `require.NotNil`, `require.True` (as a guard).
- `assert` — logs failure but continues the test. Use for the **actual values** being verified: `assert.Equal`, `assert.Contains`, `assert.True`, `assert.False`, `assert.Nil`.

Rule of thumb: if a failure would cause a nil-pointer panic or make every following assertion nonsensical, use `require`; otherwise use `assert`.

```go
func TestRun_Handoff(t *testing.T) {
	t.Parallel()

	t.Run(
		"handoff with custom tool name and description",
		func(t *testing.T) {
			t.Parallel()

			// ... setup ...

			result, err := triage.Run(
				context.Background(),
				[]llm.Message{
					userMessage("How much is my invoice?"),
				},
			)

			require.NoError(t, err)
			assert.Equal(t, "Your invoice is $42.", result.FinalMessage().Text())
			assert.Equal(t, "billing", result.LastAgent.Name())
		},
	)
}
```

**Helpers and mocks:** Define mock types and helper functions (e.g. `stopResponse`, `userMessage`) at the top of the test file, not inline in each test.

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

### Service `Run()` orchestration

A top-level `Run` method starts child subsystems (workers, servers) as goroutines via `sync.WaitGroup.Go`. Each child gets its own cancellable context created with `context.WithCancel(context.WithoutCancel(ctx))` so that a parent cancellation does not kill in-flight work — the parent explicitly calls each `stop*` function and then `wg.Wait()` for a controlled shutdown.

When a child crashes, it calls `cancel(fmt.Errorf("… crashed: %w", err))` to signal the parent.

```go
func (impl *Implm) Run(ctx context.Context, l *log.Logger) error {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	// Start a worker
	workerCtx, stopWorker := context.WithCancel(context.WithoutCancel(ctx))
	worker := NewFooWorker(pgClient, l.Named("foo-worker"))
	wg.Go(
		func() {
			if err := worker.Run(workerCtx); err != nil {
				cancel(fmt.Errorf("foo worker crashed: %w", err))
			}
		},
	)

	// Start a server
	serverCtx, stopServer := context.WithCancel(context.WithoutCancel(ctx))
	defer stopServer()
	wg.Go(
		func() {
			if err := impl.runServer(serverCtx, l); err != nil {
				cancel(fmt.Errorf("server crashed: %w", err))
			}
		},
	)

	<-ctx.Done()

	stopServer()
	stopWorker()

	wg.Wait()

	return context.Cause(ctx)
}
```

### Workers

Background workers follow a poll-based pattern with bounded concurrency. The struct holds a `*pg.Client`, a `*log.Logger`, and tuning knobs (`interval`, `staleAfter`, `maxConcurrency`). Use functional options (`With*` functions) for the tuning knobs with sensible defaults.

The `Run(ctx context.Context) error` method loops with a `select` on `ctx.Done()` and `time.After(interval)`. On each tick it recovers stale rows, then drains available work via `processNext`. Work items are claimed inside a transaction with `FOR UPDATE SKIP LOCKED`, marked as processing, then handled concurrently in goroutines bounded by a semaphore channel. Use `context.WithoutCancel` for work that must complete even after shutdown, and `sync.WaitGroup` with `defer wg.Wait()` to ensure in-flight goroutines finish before `Run` returns.

```go
type (
	FooWorker struct {
		pg             *pg.Client
		logger         *log.Logger
		interval       time.Duration
		staleAfter     time.Duration
		maxConcurrency int
	}

	FooWorkerOption func(*FooWorker)
)

func NewFooWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...FooWorkerOption,
) *FooWorker {
	w := &FooWorker{
		pg:             pgClient,
		logger:         logger,
		interval:       10 * time.Second,
		staleAfter:     5 * time.Minute,
		maxConcurrency: 5,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (w *FooWorker) Run(ctx context.Context) error {
	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, w.maxConcurrency)
	)
	defer wg.Wait()

LOOP:
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(w.interval):
		nonCancelableCtx := context.WithoutCancel(ctx)
		w.recoverStaleRows(nonCancelableCtx)
		for {
			if err := w.processNext(ctx, sem, &wg); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					w.logger.ErrorCtx(nonCancelableCtx, "cannot claim item", log.Error(err))
				}
				break
			}
		}
		goto LOOP
	}
}

func (w *FooWorker) processNext(ctx context.Context, sem chan struct{}, wg *sync.WaitGroup) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	var (
		item coredata.FooItem
		now  = time.Now()
		nonCancelableCtx = context.WithoutCancel(ctx)
	)

	if err := w.pg.WithTx(
		nonCancelableCtx,
		func(tx pg.Conn) error {
			if err := item.LoadNextPendingForUpdateSkipLocked(nonCancelableCtx, tx); err != nil {
				return err
			}
			item.Status = coredata.FooStatusProcessing
			item.UpdatedAt = now
			return item.Update(nonCancelableCtx, tx, coredata.NewNoScope())
		},
	); err != nil {
		<-sem
		return err
	}

	wg.Add(1)
	go func(item coredata.FooItem) {
		defer wg.Done()
		defer func() { <-sem }()

		if err := w.handle(nonCancelableCtx, &item); err != nil {
			w.logger.ErrorCtx(nonCancelableCtx, "cannot process item", log.Error(err))
		}
	}(item)

	return nil
}
```
