# Logging — Structured, PII-free Observability

Library: `go.gearno.de/kit/log` — named, context-aware structured logger with typed fields.

## Golden rule

**Only log opaque identifiers (IDs, correlation IDs, request IDs). Never log PII, PHI, credentials, or any data that can identify or harm a natural person.**

## What must never appear in logs

| Category | Examples | Why |
|----------|----------|-----|
| **PII** (Personally Identifiable Information) | Email addresses, full names, phone numbers, IP addresses of end-users, postal addresses, dates of birth | GDPR / privacy — log the entity ID instead |
| **PHI** (Protected Health Information) | Medical records, health status, insurance IDs | HIPAA / privacy |
| **Credentials & secrets** | Passwords, API keys, tokens (access, refresh, bearer, SCIM), signing secrets, private keys, client secrets | Security — a leaked log line becomes a breach |
| **User-generated content** | Form input, document text, chat messages, file contents | May embed PII/PHI; can also be weaponized for log injection |
| **Financial data** | Credit card numbers, bank accounts, billing addresses | PCI-DSS |

## What is safe to log

| Safe | Example field |
|------|--------------|
| Entity IDs (GIDs, UUIDs) | `log.String("identity_id", identity.ID.String())` |
| Correlation / request IDs | `log.String("correlation_id", correlationID)` |
| Operation names and types | `log.String("graphql_operation_name", operationName)` |
| URL paths (without query strings containing secrets) | `log.String("path", r.URL.Path)` |
| Domain names (public, not user-chosen) | `log.String("domain", domain)` |
| Counts, sizes, durations | `log.Int("count", n)`, `log.Duration("elapsed", d)` |
| Error messages from internal code | `log.Error(err)` |
| State transitions / enum values | `log.String("state", string(newState))` |
| Timestamps | `log.Time("expires_at", cert.ExpiresAt)` |

## Logger wiring and progressive enrichment

Loggers are constructor-injected and progressively enriched with `.Named()` / `.With()` as they flow deeper into the call stack. Every layer that has meaningful context **must** derive a child logger and attach attributes — this builds up a rich, filterable log trail without repeating fields on every call site.

### Deriving loggers

- **`.Named(subsystem)`** — adds a dot-separated name prefix to every log line. Use at service/component boundaries.
- **`.With(fields...)`** — returns a new logger with permanent structured fields. Use when an identifier or attribute is known for the lifetime of that scope (a request, a job iteration, a connection).

Always assign the derived logger to a new variable or field — never mutate the parent:

```go
// Service constructor — name the subsystem
func NewRenewer(logger *log.Logger, ...) *Renewer {
	return &Renewer{
		logger: logger.Named("renewer"),
	}
}

// Worker iteration — attach the entity being processed
func (r *Renewer) renewDomain(ctx context.Context, domain CustomDomain) {
	logger := r.logger.With(
		log.String("domain", domain.Domain),
		log.String("custom_domain_id", domain.ID.String()),
	)

	logger.InfoCtx(ctx, "starting certificate renewal")
	// ... all subsequent logs in this function carry domain + custom_domain_id
	logger.InfoCtx(ctx, "certificate renewed successfully")
}
```

### Enrichment chain

The root logger created in `probod` flows through the system, gaining context at each layer:

```
probod (root)
  → .Named("http.server")
    → .Named("api")
      → .With(log.String("correlation_id", id))
        → .Named("certmanager")
          → .With(log.String("domain", d))
```

A log line emitted at the leaf carries **all** ancestor attributes automatically. This means:
- You never need to repeat `correlation_id` or `organization_id` at inner call sites
- Filtering by any attribute in the chain works across the entire request span
- Adding a new attribute at one layer enriches every log line below it

### Where to derive

| Boundary | Derive with | Typical attributes |
|----------|------------|-------------------|
| Service/component constructor | `.Named("subsystem")` | — |
| HTTP middleware / request entry | `.With(...)` | `correlation_id`, `identity_id`, `path` |
| Worker job iteration | `.With(...)` | entity ID being processed |
| Agent / tool execution | `.Named("agent").With(...)` | `agent`, tool name |
| Retry / loop body | `.With(...)` | `attempt`, iteration key |

### Fallback

In HTTP handlers where no constructor-injected logger is available (e.g. panic recovery), use `httpserver.LoggerFromContext(ctx)` — it returns the request-scoped logger with all middleware-attached attributes.

## Structured field helpers

Use typed field constructors — never `fmt.Sprintf` into a log message for structured data:

| Helper | Use for |
|--------|---------|
| `log.String(key, val)` | String values (IDs, domains, operation names) |
| `log.Int(key, val)` | Integer counts |
| `log.Int64(key, val)` | Large integers (byte sizes, database counts) |
| `log.Bool(key, val)` | Flags |
| `log.Float64(key, val)` | Floating-point metrics |
| `log.Duration(key, val)` | `time.Duration` values |
| `log.Time(key, val)` | `time.Time` values |
| `log.Error(err)` | Error values (key is automatically `"error"`) |
| `log.Any(key, val)` | Last resort — prefer a typed helper when one exists |

## Context-aware methods

Always prefer the `*Ctx` variants to propagate trace/request context:

```go
logger.InfoCtx(ctx, "message", log.String("key", "value"))
logger.WarnCtx(ctx, "message", log.String("key", "value"))
logger.ErrorCtx(ctx, "message", log.Error(err))
```

Use the non-context `Info` / `Error` only at process startup/shutdown where no request context exists.

## Rules

- **Always derive, never repeat.** When you enter a new scope that has a meaningful identifier (a request, a job, an entity), derive a child logger with `.Named()` or `.With()` and use it for all subsequent calls in that scope. Never pass the same attribute as an inline field on every log call — attach it once on the derived logger.
- **Log IDs, not values.** Instead of `log.String("email", user.Email)`, write `log.String("identity_id", user.ID.String())`.
- **Treat error descriptions from external sources as untrusted.** OAuth `error_description`, OIDC provider messages, and SAML responses may contain user data or be provider-controlled. Log a sanitized error code, not the full description.
- **Guard `log.Any`.** The `log.Any` helper serializes arbitrary values. Never pass structs that may contain sensitive fields. Prefer explicit field selection.
- **Never log raw HTTP bodies or query strings.** Query strings may carry `code`, `token`, or `state` parameters. Log `r.URL.Path` only.
- **Never log `fmt.Errorf` messages that embed secrets.** If an error wraps sensitive context (e.g. a decryption failure message that echoes input), strip it before logging.
- **GraphQL errors are semi-public.** `log.Any("errors", resp.Errors)` is acceptable because GraphQL errors are already filtered for the client, but never add raw input variables to log context.
- **Keep log messages static.** The message string should be a fixed human-readable sentence. Dynamic data goes in structured fields, not in `fmt.Sprintf` message templates.
