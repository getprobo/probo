# Fix: 500 Internal Server Error on GraphQL

## Symptom

When loading the console app at `http://localhost:5173`, you see:

```
POST http://localhost:5173/api/connect/v1/graphql net::ERR_ABORTED 500 (Internal Server Error)
```

The React app crashes with `InternalServerError: INTERNAL_SERVER_ERROR` in `ViewerMembershipLayoutLoader`.

## Cause

The Go backend (`probod`) is not running. The Vite dev server proxies API requests to the backend, so if `probod` is down, all GraphQL calls return 500.

## Fix

### 1. Make sure Docker infra is running

```bash
make stack-up
```

Verify all containers are healthy:

```bash
docker compose ps
```

You should see `postgres`, `keycloak`, `mailpit`, `seaweedfs`, etc. all with status `Up`.

### 2. Build and start the backend

```bash
SKIP_APPS=1 make build
bin/probod -cfg-file cfg/dev.yaml
```

> **Important:** `probod` requires the `-cfg-file cfg/dev.yaml` flag. Without it, it panics with:
> ```
> panic: cannot parse encryption key: key must be 32 bytes for AES-256, got 0 bytes after base64 decoding
> ```

### 3. Verify the backend is responding

```bash
curl -s http://localhost:8080/api/connect/v1/graphql \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"query":"{__typename}"}'
```

Expected response:

```json
{"data":{"__typename":"Query"}}
```

### 4. Refresh the browser

Go to `http://localhost:5173` and the app should load normally.

## Quick Reference

| Service        | Port  | Start Command                        |
|----------------|-------|--------------------------------------|
| Docker infra   | —     | `make stack-up`                      |
| Backend        | 8080  | `bin/probod -cfg-file cfg/dev.yaml`  |
| Vite (console) | 5173  | `npm run dev` (from `apps/console/`) |
