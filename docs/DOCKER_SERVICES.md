# Docker Services

Local development stack started with `make stack-up` (runs `docker compose up -d`).
Full service definitions live in `compose.yaml`; CI overlays `compose.github-action.yaml`.

## Services

| Service | Ports (host:container) | Purpose | Config / data |
| --- | --- | --- | --- |
| postgres | 5432:5432 | Primary database. Init script creates `probod` superuser, `probod` and `probod_test` databases. | `compose/postgres/01_probod.sh`, volume `postgres-data` |
| seaweedfs | 8333:8333, 9333:9333, 8888:8888 | S3-compatible object store for file uploads (S3 API on 8333). Dev identity `probod` is defined in `s3.json`. | `compose/seaweedfs/s3.json`, volume `seaweedfs-data` |
| grafana | 3001:3000 | Metrics and traces dashboards (anonymous admin access in dev). | `compose/grafana/provisioning`, volume `grafana-data` |
| prometheus | 9191:9191 | Metrics collection (`prometheus.yaml` scrapes prometheus and tempo). | `compose/prometheus/prometheus.yaml`, volume `prometheus-data` |
| loki | 3100:3100 | Log aggregation backend for Grafana. | No local volume |
| tempo | 3200:3200, 4317:4317, 4318:4318 | Trace storage (OTLP gRPC on 4317, HTTP on 4318). | `compose/tempo/tempo.yaml`, volume `tempo-data` |
| mailpit | 1025:1025 (SMTP), 8025:8025 (UI) | Local mail catcher for transactional emails. | No local volume |
| chrome | 9222:9222 | Headless Chrome used for document rendering and exports. | No local volume |
| acme-http-01-proxy | 80:80, 9000:9000 | Caddy proxy for ACME HTTP-01 challenges to probod on host `:10080`. | `compose/caddy/Caddyfile` |
| step-ca | Shares `acme-http-01-proxy` network | Local certificate authority ("Probo Local CA") for development TLS. Issues the root CA at `compose/step-ca/certs/root_ca.crt`. | `compose/step-ca`, no named volume |
| keycloak | 8082:8080 | Local identity provider for SSO development (dev admin `admin` / `admin`). Realm imported from `probo-realm.json.tmpl`. | `compose/keycloak`, volume `keycloak-data` |

## Useful commands

```sh
make stack-up    # start the stack as a daemon
make stack-ps    # list stack containers
make stack-down  # stop the stack
make psql        # open psql to the postgres container
```

## Notes

- All credentials above are local-development defaults defined in `compose.yaml`
  and `compose/postgres/01_probod.sh`. Never reuse them outside local dev.
- `make stack-up` also waits for the step-ca root certificate
  (`compose/step-ca/certs/root_ca.crt`) before finishing; see the `stack-up`
  target in `GNUmakefile`.
- Fork pull requests run tests against origin registries instead of the
  `artifact.probo.inc` mirror; see `.github/workflows/make.yaml`.
