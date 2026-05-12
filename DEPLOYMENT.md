# Govrly Deployment on Render — Complete Guide

## Overview
Deploy the open-source Govrly GRC platform on Render (Docker web service) with AWS RDS PostgreSQL.

## Pre-requisites
- GitHub repo with the Govrly source code
- Render account
- AWS account (for RDS PostgreSQL)
- Cloudflare R2 bucket (for file storage)

---

## Step 1: Create Dockerfile.render

The original `Dockerfile` expects a pre-built binary. We need a multi-stage Dockerfile that builds from source.

**IMPORTANT:** Use Alpine (not Ubuntu) as the runtime base. Ubuntu causes `Operation not permitted` errors on Render due to security policies (`no-new-privileges:true`). The binary must live in `/app/` with `--chown`, not in `/usr/local/bin/` with `chmod +x`.

```dockerfile
# Stage 1: Build frontend assets
FROM node:24-alpine AS frontend
RUN apk add --no-cache findutils
WORKDIR /app
COPY package.json package-lock.json turbo.json ./
COPY apps/ apps/
COPY packages/ packages/
COPY pkg/server/api/connect/v1/schema.graphql pkg/server/api/connect/v1/schema.graphql
COPY pkg/server/api/console/v1/schema.graphql pkg/server/api/console/v1/schema.graphql
COPY pkg/server/api/trust/v1/schema.graphql pkg/server/api/trust/v1/schema.graphql
RUN npm ci
RUN npx turbo run relay
RUN npx turbo run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS backend
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/apps/console/dist apps/console/dist
COPY --from=frontend /app/apps/trust/dist apps/trust/dist
COPY --from=frontend /app/packages/emails/dist packages/emails/dist
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'main.version=dev' -X 'main.env=prod'" \
    -o /probod ./cmd/probod

# Stage 3: Runtime (Alpine — required for Render)
FROM alpine:latest
RUN apk --no-cache add ca-certificates openssl bash && \
    addgroup -g 1000 probo && \
    adduser -D -u 1000 -G probo probo && \
    mkdir -p /etc/probod && \
    chown probo:probo /etc/probod
WORKDIR /app
COPY --from=backend --chown=probo:probo /probod ./probod
COPY --chown=probo:probo entrypoint.sh ./entrypoint.sh
COPY --chown=probo:probo rds-ca-bundle.pem ./rds-ca-bundle.pem
USER probo
EXPOSE 8080
CMD ["bash", "./entrypoint.sh"]
```

---

## Step 2: Fix eslint.config.ts type errors

Both `apps/trust/tsconfig.node.json` and `apps/console/tsconfig.node.json` include `eslint.config.ts` in compilation. Type incompatibility between `@typescript-eslint` and `@eslint/core`.

**Fix:** Add `"exclude": ["eslint.config.ts"]` to both tsconfig.node.json files.

---

## Step 3: Add missing disposable email blocklist

```bash
curl -sL "https://raw.githubusercontent.com/disposable-email-domains/disposable-email-domains/master/disposable_email_blocklist.conf" \
  -o pkg/validator/data/disposable-email-domains/disposable_email_blocklist.conf
```

Commit this file.

---

## Step 4: Fix entrypoint.sh

Change last line from `exec probod` to `exec ./probod`:
```bash
exec ./probod -cfg-file "$CONFIG_FILE" "$@"
```

---

## Step 5: Create .dockerignore

```
.git
node_modules
bin
coverage
*.md
.env
.env.*
.DS_Store
.idea
.vscode
compose
compose.yaml
compose.prod.yaml
contrib
e2e
docs
sbom.json
sbom-docker.json
```

---

## Step 6: Download RDS CA certificate

Match the region to your RDS instance:
```bash
curl -sL "https://truststore.pki.rds.amazonaws.com/<REGION>/<REGION>-bundle.pem" -o rds-ca-bundle.pem
```

---

## Step 7: AWS RDS Setup

1. Create PostgreSQL instance
2. Enable public access
3. Security group: inbound PostgreSQL (5432) from 0.0.0.0/0
4. Default database is `postgres` (DB identifier ≠ database name)

---

## Step 8: Render Setup (Manual)

1. Web Service → Docker
2. Dockerfile Path: `./Dockerfile.render`

---

## Step 9: Environment Variables

### Secrets (generate each with `openssl rand -base64 32`):
- PROBOD_ENCRYPTION_KEY
- AUTH_COOKIE_SECRET
- AUTH_PASSWORD_PEPPER
- TRUST_AUTH_TOKEN_SECRET

### App config:
- PROBOD_BASE_URL = https://your-app.onrender.com
- API_ADDR = :10000
- API_CORS_ALLOWED_ORIGINS = https://your-app.onrender.com
- AUTH_COOKIE_DOMAIN = your-app.onrender.com
- AUTH_COOKIE_SECURE = true

### PostgreSQL:
- PG_ADDR = your-rds-endpoint:5432
- PG_USERNAME = postgres
- PG_PASSWORD = <password>
- PG_DATABASE = postgres
- PG_POOL_SIZE = 20
- PG_CA_BUNDLE_PATH = /app/rds-ca-bundle.pem

### Telemetry (disable):
- TRACING_ADDR = none
- METRICS_ADDR = none

### R2 Storage:
- AWS_ENDPOINT = https://<account-id>.r2.cloudflarestorage.com
- AWS_BUCKET = <bucket-name>
- AWS_REGION = auto
- AWS_ACCESS_KEY_ID = <key>
- AWS_SECRET_ACCESS_KEY = <secret>
- AWS_USE_PATH_STYLE = true

### SMTP (when ready):
- SMTP_ADDR = <smtp server>
- SMTP_TLS_REQUIRED = true

---

## Performance Note

**IMPORTANT:** Place Render and RDS in the **same region**. Cross-region latency (e.g., Render in us-west, RDS in ap-south-1) adds ~300ms per database query, causing 10+ second page loads.

For Egypt/Saudi users, **eu-central-1 (Frankfurt)** is ideal for both Render and RDS (~50ms from Egypt).

---

## All Issues & Fixes

| # | Error | Fix |
|---|-------|-----|
| 1 | `"/linux/amd64/probod": not found` | Use Dockerfile.render with multi-stage build |
| 2 | eslint.config.ts type errors | Exclude from tsconfig.node.json |
| 3 | `disposable_email_blocklist.conf: no matching files found` | Download and commit the file |
| 4 | `Operation not permitted` (exit 126) | Alpine + --chown + /app/ path + CMD |
| 5 | `dial tcp 127.0.0.1:5432: connection refused` | Set PG_ADDR to RDS endpoint |
| 6 | `argument list too long` | Bake CA cert in image, use PG_CA_BUNDLE_PATH |
| 7 | `no pg_hba.conf entry... no encryption` | Provide RDS CA bundle for SSL |
| 8 | `certificate signed by unknown authority` | Use correct region's CA bundle |
| 9 | `database "govrly" does not exist` | Set PG_DATABASE=postgres |
| 10 | `traces export: connection refused` spam | Set TRACING_ADDR=none |
| 11 | Render detects port 80 | Use API_ADDR=:10000 |
| 12 | Email links → localhost | Set PROBOD_BASE_URL correctly |
| 13 | `findutils` missing | apk add findutils in frontend stage |
| 14 | `go generate` fails | Don't run it — generated files already committed |
