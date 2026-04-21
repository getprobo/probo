# Configuration Propagation

When a configuration field is added, renamed, or removed in the Go config structs, **all** downstream consumers must be updated in the same change. The config struct in `pkg/probod/` is the source of truth.

## Files to update (checklist)

| # | File | Role |
|---|------|------|
| 1 | `pkg/probod/*.go` | Go config structs — source of truth |
| 2 | `pkg/probod/probod.go` `New()` | Default values for new fields |
| 3 | `pkg/bootstrap/builder.go` | Env-var → struct mapping (`Build()` method) |
| 4 | `pkg/bootstrap/builder.go` | Required-env validation (`validateRequired()`) |
| 5 | `GNUmakefile` (`dev-config` target) | Env vars fed to `probod-bootstrap` to regenerate `cfg/dev.yaml` (file itself is gitignored) |
| 6 | `e2e/internal/testutil/testutil.go` | E2E env-var map fed to `bootstrap.NewBuilder` |
| 7 | `contrib/lima/provision.sh` | Sandbox env vars passed to `probod-bootstrap` |
| 8 | `contrib/helm/charts/probo/values.yaml` | Helm default values |
| 9 | `contrib/helm/charts/probo/values-production.yaml.example` | Helm production template |
| 10 | `contrib/helm/charts/probo/templates/deployment.yaml` | Helm deployment — maps values → env vars |
| 11 | `contrib/helm/charts/probo/templates/secret.yaml` | Helm secret — sensitive values |

## Flow

```
Go struct (pkg/probod/)
  │
  ├─► probod New() defaults
  │
  ├─► bootstrap builder.go (env var → struct)
  │     │
  │     ├─► GNUmakefile dev-config    (env vars → probod-bootstrap → cfg/dev.yaml)
  │     ├─► e2e/internal/testutil/    (env map → bootstrap.Build, tests)
  │     ├─► contrib/lima/provision.sh  (env vars → probod-bootstrap)
  │     └─► Helm chart
  │           ├─ values.yaml           (user-facing knobs)
  │           ├─ values-production.yaml.example
  │           ├─ templates/deployment.yaml (values → env vars)
  │           └─ templates/secret.yaml     (sensitive values)
  │
  └─► probod.go Run() (wiring into services)
```

## Rules

1. **Never add a Go config field without updating every file in the checklist.**
2. **Env var naming** — follow the existing convention in `builder.go`: `SECTION_FIELD_NAME` (e.g. `AUTH_COOKIE_DOMAIN`, `CUSTOM_DOMAINS_RENEWAL_INTERVAL`).
3. **Secrets** go through `secret.yaml` and are referenced via `secretKeyRef` in `deployment.yaml`. Non-secret values are set inline.
4. **`make dev-config`** writes `cfg/dev.yaml` via `probod-bootstrap` with safe, non-production defaults (plaintext passwords, `localhost`, `secure: false`). The generated file and the per-dev OAuth2 signing key (`cfg/.dev-oauth2-signing-key.pem`) are both gitignored. The recipe sources `.env` at the repo root if present so devs can override any env var without editing the `GNUmakefile`; keep `.env.example` in sync when you add or rename env vars.
5. **`e2e/internal/testutil/testutil.go`** builds the e2e config through `bootstrap.NewBuilder` with a test-only env-var map (different ports, `probod_test` DB, shorter intervals). Any new field whose test value differs from the bootstrap default must be added to that map.
6. **`provision.sh`** only sets env vars that differ from `builder.go` defaults (e.g. `PROBOD_BASE_URL`, `AUTH_COOKIE_DOMAIN`, `AUTH_COOKIE_SECURE`). If the new field's default is acceptable in the sandbox, no env var is needed.
7. **Helm `values.yaml`** exposes the field under the appropriate `probo.*` key with a sensible default. `values-production.yaml.example` includes it only when the production value differs or the user must set it.
8. **Optional features** (custom domains, SAML, connectors, tracing) are gated by `{{- if }}` blocks in the Helm templates; follow the same pattern for new optional fields.
9. **Bootstrap tests** (`pkg/bootstrap/builder_test.go`) must cover the new env var mapping.
