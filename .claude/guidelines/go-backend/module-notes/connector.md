# Probo — Go Backend — pkg/connector

**Purpose.** Generic 3rd-party SaaS connector framework. Provides
OAuth2 authorization-code, OAuth2 client-credentials, and API-key
authentication. Centralises HMAC-signed stateless `state` token,
SSRF-protected HTTP transport, and per-provider config (URLs, scopes,
auth mode).

> See [patterns.md § 12](../patterns.md#12-connector-oauth2-framework).

**Key files.**

- `connector.go` — `Connector` interface (`Initiate`, `Complete`),
  `Connection` interface (`Type`, `Client`, `Scopes`,
  Marshal/Unmarshal).
- `oauth2.go` — `OAuth2Connector`, `OAuth2Connection`,
  `oauth2Transport` (Bearer-token round tripper),
  `RefreshableClient`, `clientCredentialsClient`.
- `apikey.go` — `APIKeyConnection` (static Bearer key).
- `slack.go` — `SlackConnection` (provider-specific extension).
- `providers.go` — `providerDefinitions` map (SLACK, GITHUB,
  GOOGLE_WORKSPACE, LINEAR, HUBSPOT, DOCUSIGN, NOTION, SENTRY,
  INTERCOM, BREX, CLOUDFLARE, OPENAI, SUPABASE, TALLY, RESEND,
  ONE_PASSWORD) + `ApplyProviderDefaults`.
- `registry.go` — `ConnectorRegistry` (thread-safe map; OAuth2 probe
  URLs).
- `scopes.go` — `ParseScopeString`, `FormatScopeString`, `UnionScopes`.

**Three `TokenEndpointAuth` modes** (chosen per provider in
`providerDefinitions`):

- `post-form` — credentials in form body (most providers).
- `basic-form` — Basic auth header + form body.
- `basic-json` — Basic auth header + JSON body (Notion).

**How to extend (a new OAuth2 provider).**

1. `pkg/connector/providers.go` — add the entry to
   `providerDefinitions` (AuthURL, TokenURL, ExtraAuthParams,
   TokenEndpointAuth, SupportsIncrementalAuth).
2. `pkg/probod/probod.go` — call
   `ApplyProviderDefaults(connector, providerName)` then
   `registry.Register(providerName, connector)` with the deployment
   client ID/secret.
3. GraphQL/MCP enum — add the provider name to the four-surface enum so
   clients can request it (see
   [shared.md § 3](../../shared.md#3-the-four-surface-api-rule)).

**Top pitfalls.**

- `OAuth2Connector.HTTPClient` is required and must be SSRF-protected.
  `ApplyProviderDefaults` injects it; tests must inject
  `httpclient.DefaultClient(WithSSRFProtection(), WithSSRFAllowLoopback())`.
- Three-map edit is easy to half-do — see
  [pitfalls.md § 15](../pitfalls.md).
- `state` token must round-trip safely; never store side-data in a DB
  for OAuth2 state (the framework intentionally avoids that).
