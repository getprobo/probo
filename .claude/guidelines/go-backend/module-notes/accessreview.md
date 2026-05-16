# Probo — Go Backend — pkg/accessreview

**Purpose.** Provider-driver registry for access-review integrations
across 15+ SaaS apps (GitHub, Okta, Google Workspace, Slack, AWS,
Linear, Notion, ...). Each driver implements the `Driver` interface
(list users, list group memberships, etc.) backed by a
`connector.Connection`.

**Key files.**

- `pkg/accessreview/driver.go` — `Driver` interface, **switch-based**
  `NewDriver(provider, conn) (Driver, error)` registry.
- `pkg/accessreview/<provider>/` — one sub-package per provider.

**How to extend (a new provider).**

1. Create `pkg/accessreview/<provider>/driver.go` implementing the
   `Driver` interface.
2. Add a `case` to `NewDriver` in `pkg/accessreview/driver.go`.
3. Register the provider in `pkg/connector/providers.go` (OAuth2 URLs
   and probe URL).
4. Add it to the GraphQL/MCP enum exposing connector providers.

**Why switch-based instead of init-side-effect registration?**

The `accessreview` package deliberately uses an explicit switch instead
of `init()` registration. The reason is *auditable diff*: adding a new
provider is a one-line `case` addition that reviewers can spot on
sight, with no hidden ordering or import-side-effect surprises. The
only place we use `init`-style registration is `pkg/connector`, and
only because it must wire deployment-supplied client IDs/secrets at
startup.

**Top pitfalls.**

- Forgetting the `case` in `NewDriver` after adding the sub-package —
  runtime "unsupported provider" error at first use. The build won't
  catch this.
- Skipping the connector registry edit — the OAuth2 flow has nowhere
  to go. See [pitfalls.md § 15](../pitfalls.md).
