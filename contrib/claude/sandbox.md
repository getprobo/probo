# Sandbox Environments

## When to use

Use a sandbox when you need to:
- Run the full service stack (Docker services, probod, console)
- Test changes end-to-end with `make test-e2e`
- Build the full binary with `make build`
- Run any command that requires Docker or the full service stack

## Quick reference

```bash
# Create a sandbox (first time only)
./contrib/lima/sandbox.sh create

# Start an existing sandbox
./contrib/lima/sandbox.sh start

# Build and start the app (probo-stack starts automatically on boot)
./contrib/lima/sandbox.sh exec -- make build
./contrib/lima/sandbox.sh exec -- sudo systemctl start probod probo-console

# Get the VM IP and service URLs
./contrib/lima/sandbox.sh status

# Interactive shell
./contrib/lima/sandbox.sh ssh

# Stop (shutdown — preserves disk and Docker images, but running processes are lost)
./contrib/lima/sandbox.sh stop

# Delete entirely
./contrib/lima/sandbox.sh delete
```

## Accessing services

After `sandbox.sh status`, use the VM IP to access services from the host:

| Service | URL |
|---|---|
| Console | `http://<vm-ip>:5173` |
| API | `http://<vm-ip>:8080` |
| Grafana | `http://<vm-ip>:3001` |
| Mailpit | `http://<vm-ip>:8025` |
| Keycloak | `http://<vm-ip>:8082` |
| PostgreSQL | `psql -h <vm-ip> -U probod` |

## Auto-generated configuration

During provisioning, the sandbox automatically generates:

- **`/etc/probod/config.yml`** — probod config with the VM IP as cookie domain, `secure: false`, and correct CORS origins
- **`apps/console/.env`** and **`apps/trust/.env`** — `VITE_API_URL` pointing to the VM IP

Probod config is at `/etc/probod/config.yml`.

### Custom environment variables

To inject developer-specific secrets (SSO, API keys, etc.) into the sandbox, create a `.sandbox.env` file at the repo root:

```bash
# .sandbox.env (gitignored — never committed)
AUTH_SAML_IDP_METADATA_URL=https://login.example.com/metadata
AUTH_OIDC_CLIENT_ID=my-client-id
AUTH_OIDC_CLIENT_SECRET=s3cret
```

This file is sourced during provisioning before `probod-bootstrap` runs. Any variable set here overrides the defaults. The sandbox must be recreated (`delete` + `create`) for changes to take effect.

## Systemd services

The sandbox provisions three systemd services:

| Service | Description | Starts on boot |
|---|---|---|
| `probo-stack` | Docker Compose stack (Postgres, SeaweedFS, Keycloak, etc.) | Yes |
| `probod` | Probo API server (depends on `probo-stack`) | No |
| `probo-console` | Console frontend dev server | No |

`probo-stack` starts automatically when the VM boots. `probod` and `probo-console` must be started manually after building.

Manage them with `systemctl`:
```bash
./contrib/lima/sandbox.sh exec -- sudo systemctl start probod probo-console
./contrib/lima/sandbox.sh exec -- sudo systemctl stop probod
./contrib/lima/sandbox.sh exec -- sudo systemctl restart probod
./contrib/lima/sandbox.sh exec -- sudo systemctl status probod
./contrib/lima/sandbox.sh exec -- sudo journalctl -u probod -f
```

## Common workflows

**Start the app:**
```bash
./contrib/lima/sandbox.sh exec -- make build
./contrib/lima/sandbox.sh exec -- sudo systemctl start probod probo-console
```

**Run tests:**
```bash
./contrib/lima/sandbox.sh exec -- make build
./contrib/lima/sandbox.sh exec -- make test
```

**Run e2e tests:**
```bash
./contrib/lima/sandbox.sh exec -- make test-e2e
```

**Restart after code changes:**
Code changes are reflected immediately (virtiofs mount). Just rebuild and
restart probod — no need to restart the VM.

```bash
./contrib/lima/sandbox.sh exec -- make build
./contrib/lima/sandbox.sh exec -- sudo systemctl restart probod
```
