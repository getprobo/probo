# HTTP Client

Use `go.gearno.de/kit/httpclient` for every outbound HTTP call. Never use `http.DefaultClient` or a bare `&http.Client{}`.

## SSRF protection is the default

Every call goes through `httpclient.DefaultClient(...)` / `httpclient.DefaultPooledClient(...)` / `httpclient.DefaultPooledTransport(...)` with `httpclient.WithSSRFProtection()` enabled. This applies whenever the destination URL is:

- Customer-supplied (webhook endpoint, OAuth2 token URL, SCIM bridge URL, connector-provided base URL)
- Reached through a customer-supplied connector (OAuth2/APIKey connection clients)
- A hardcoded third-party provider host (Slack, Linear, GitHub, Google Workspace, Sentry, …) — defense in depth, and public IPs pass the check unchanged

```go
client := httpclient.DefaultPooledClient(
    httpclient.WithLogger(logger),
    httpclient.WithSSRFProtection(),
)
```

What the option does:

- Rejects dials to loopback, RFC 1918 private, RFC 6598 CGNAT, link-local, multicast, unspecified, ULA, IPv4-mapped IPv6, and IETF-reserved ranges. Check runs on the resolved peer IP at connect time, defeating DNS rebinding.
- On `DefaultClient` / `DefaultPooledClient`, also refuses redirects whose scheme, host, or port differs from the original.

## When to omit SSRF protection

Only when the target is an **internal service you actively intend to reach** (sandbox-local service, sidecar, in-cluster endpoint with a known private IP). These cases are rare in this codebase — confirm the intent in code review. Do not disable it "just to make a test pass."

For tests that need to hit an `httptest` server on loopback, add `httpclient.WithSSRFAllowLoopback()` on top of `WithSSRFProtection()` (or inject a loopback-friendly client into the component under test). Production callers must not use the loopback exemption.

## Connector wiring

`OAuth2Connector.HTTPClient` is a required field for the token-exchange request. Set it via `connector.ApplyProviderDefaults`, which wires an SSRF-protected client. Don't re-introduce a nil fallback in `CompleteWithState` — callers must provide the client explicitly.

`OAuth2Connection.ClientWithOptions`, `RefreshableClient`, `clientCredentialsClient` and `APIKeyConnection.Client` already append `WithSSRFProtection()` internally; additional caller options layer on top.

## Summary

- Customer-reachable URL → `WithSSRFProtection()`, always.
- Third-party SaaS → `WithSSRFProtection()`, always (public IPs pass).
- Local/internal service you meant to dial → document why and omit.
- Test against `httptest` → `WithSSRFProtection()` + `WithSSRFAllowLoopback()`.
