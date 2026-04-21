# Configuration Propagation

When a configuration field is added, renamed, or removed in the Go config structs, **all** downstream consumers must be updated in the same change. The config struct in `pkg/probod/` is the source of truth.

## Files to update (checklist)

| # | File | Role |
|---|------|------|
| 1 | `pkg/probod/*.go` | Go config structs — source of truth |
| 2 | `pkg/probod/probod.go` `New()` | Default values for new fields |
| 3 | `pkg/bootstrap/builder.go` | Env-var → struct mapping (`Build()` method) |
| 4 | `pkg/bootstrap/builder.go` | Required-env validation (`validateRequired()`) |
| 5 | `cfg/dev.yaml` | Local development config |
| 6 | `e2e/console/testdata/config.yaml` | E2E test config |
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
  │     ├─► cfg/dev.yaml              (static YAML, local dev)
  │     ├─► e2e/console/testdata/     (static YAML, tests)
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
4. **`cfg/dev.yaml`** uses safe, non-production defaults (plaintext passwords, `localhost`, `secure: false`).
5. **`e2e/console/testdata/config.yaml`** mirrors `cfg/dev.yaml` but with test-specific values (different ports, `probod_test` DB, shorter intervals).
6. **`provision.sh`** only sets env vars that differ from `builder.go` defaults (e.g. `PROBOD_BASE_URL`, `AUTH_COOKIE_DOMAIN`, `AUTH_COOKIE_SECURE`). If the new field's default is acceptable in the sandbox, no env var is needed.
7. **Helm `values.yaml`** exposes the field under the appropriate `probo.*` key with a sensible default. `values-production.yaml.example` includes it only when the production value differs or the user must set it.
8. **Optional features** (custom domains, SAML, connectors, tracing) are gated by `{{- if }}` blocks in the Helm templates; follow the same pattern for new optional fields.
9. **Bootstrap tests** (`pkg/bootstrap/builder_test.go`) must cover the new env var mapping.
