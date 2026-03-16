# Sandbox Environments

## When to use

Use a sandbox when you need to:
- Run `make stack-up` (Docker services: Postgres, SeaweedFS, etc.)
- Test changes end-to-end with `make dev` or `make test-e2e`
- Build the full binary with `make build`
- Run any command that requires Docker or the full service stack

## Quick reference

```bash
# Create a sandbox (first time only)
./contrib/lima/sandbox.sh create

# Start an existing sandbox
./contrib/lima/sandbox.sh start

# Run commands inside the sandbox
./contrib/lima/sandbox.sh exec -- make stack-up
./contrib/lima/sandbox.sh exec -- make build
./contrib/lima/sandbox.sh exec -- make dev
./contrib/lima/sandbox.sh exec -- make test

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

## Common workflows

**Build and test:**
```bash
./contrib/lima/sandbox.sh exec -- make stack-up
./contrib/lima/sandbox.sh exec -- make build
./contrib/lima/sandbox.sh exec -- make test
```

**Run e2e tests:**
```bash
./contrib/lima/sandbox.sh exec -- make stack-up
./contrib/lima/sandbox.sh exec -- make test-e2e
```

**Restart after code changes:**
Code changes are reflected immediately (virtiofs mount). Just re-run the
relevant make target — no need to restart the VM.

```bash
./contrib/lima/sandbox.sh exec -- make dev
```
