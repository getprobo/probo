# Probo — Go Backend — pkg/trust

**Purpose.** Trust-center domain logic — the public-facing compliance
artefact pages served by the Trust HTTP/HTTPS server. Owns the
**visibility model** that gates each artefact (public, NDA-required,
authenticated, ...) and the access-grant + email flow that issues
access to NDA-gated content.

**Key files.**

- `pkg/trust/service.go` — `Service`, visibility model, grant flow.
- `pkg/trust/grant.go` — `GrantByIDs` and related grant helpers.
- `pkg/trust/logo.go` — `GenerateLogoURL`.

**How to extend.**

- A new visibility level: extend the enum, add a migration in
  `pkg/coredata`, update the API mappings (GraphQL trust schema,
  cookie-banner public endpoint if applicable).
- A new artefact type: define the entity in `pkg/coredata`, add a
  visibility column, surface it via `pkg/server/api/trust/v1`.

**Top pitfalls.**

- `pkg/trust/grant.go` line 387 has an inverted `shouldSendEmail`
  condition — emails go out when they should not, or the reverse. See
  [pitfalls.md § 8](../pitfalls.md). Add an e2e Mailpit assertion when
  you touch this file.
- `GenerateLogoURL` swallows errors; trust pages render with a missing
  logo. Update the signature to return an error and propagate. See
  [pitfalls.md § 9](../pitfalls.md).
- Trust-center HTTP server is a separate subsystem in
  `pkg/probod/probod.go` (uses `errgroup` internally) — see
  [probod.md](./probod.md).
