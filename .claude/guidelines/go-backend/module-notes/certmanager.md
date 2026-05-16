# Probo — Go Backend — pkg/certmanager

**Purpose.** ACME (Let's Encrypt / Pebble) + custom-domain TLS for the
Trust Center server. Provides dynamic `GetCertificate` for
`tls.Config`, an HTTP-01 challenge handler, and background provisioner
+ renewer goroutines.

**Key files.**

- `pkg/certmanager/service.go` — `Service`, `GetCertificate`, ACME
  challenge handling.
- `pkg/certmanager/provisioner.go` — background loop that issues new
  certs for newly-added custom domains.
- `pkg/certmanager/renewer.go` — background loop that renews certs
  before expiry.
- `pkg/coredata/custom_domain*.go` — domain rows + cert storage.

**How to use.** Wired in `pkg/probod/probod.go` `runTrustCenterServer`,
which is the **only** subsystem using `errgroup.WithContext` (cert
provisioner, renewer, HTTP server, HTTPS server form a unit — see
[probod.md](./probod.md) and
[pitfalls.md § 17](../pitfalls.md)).

**Top pitfalls.**

- ACME rate limits — back off aggressively on failures; the renewer
  must respect the upstream's `Retry-After`.
- Pebble (dev ACME) returns a different cert chain than Let's Encrypt;
  e2e tests against Pebble must trust the Pebble root.
- Don't add the cert subsystems to the top-level `sync.WaitGroup` — keep
  them inside `runTrustCenterServer`'s errgroup so they fail together.
