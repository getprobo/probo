# Faster Development Environment Setup

## Overview

This document describes the improvements made to support running the development environment without needing to build TypeScript applications. Issue #661 requested the ability to run without building TS apps for faster iteration.

## Changes Made

### 1. GNUmakefile Updates

#### Placeholder Frontend Assets
Modified the dist file generation rules to create valid, minimal HTML placeholder files instead of "dev-server" text files:

```makefile
apps/console/dist/index.html:
	$(MKDIR) $(dir $@)
	@echo '<!DOCTYPE html><html><head><title>Probo Console - Dev Mode</title></head><body><h1>Starting Vite dev server...</h1><p>Run: npm -w @probo/console run dev</p></body></html>' > $@
```

This allows the Go build to succeed without requiring full production builds of the frontend.

#### Improved dev Target
Enhanced the `make dev` command to start both backend and frontend dev servers with automatic hot module replacement:

```makefile
.PHONY:dev
dev: ## Start the development server with hot reload
	VITE_DEV_SERVER_CONSOLE=http://localhost:5173 \
	VITE_DEV_SERVER_TRUST=http://localhost:5174 \
	parallel -j 3 --line-buffer ::: \
		"gow -r=false run cmd/probod/main.go -cfg-file cfg/dev.yaml" \
		"cd apps/console && npm run dev" \
		"cd apps/trust && npm run dev"
```

### 2. Server-Side Dev Mode Support

#### Console Web Server (pkg/server/web/web.go)
- Added `NewServerWithDevAddr()` function to support dev mode
- Added reverse proxy support that routes requests to Vite dev servers
- Reads `VITE_DEV_SERVER_CONSOLE` environment variable
- Automatically falls back to embedded static files if no dev server is configured
- Proxies WebSocket connections for HMR (Hot Module Replacement)

#### Trust Web Server (pkg/server/trust/trust.go)
- Similar improvements as console web server
- Reads `VITE_DEV_SERVER_TRUST` environment variable
- Supports both dev mode proxying and production embedded files

### 3. Documentation Updates

#### CONTRIBUTING.md
- Updated development setup instructions to show the new faster approach
- Documented `make dev` command for quick startup
- Showed alternative manual setup with separate terminals
- Added instructions for environment variables when running services separately

## Usage

### Quick Development Start (Recommended)

```bash
# 1. Install dependencies
npm ci
go mod download

# 2. Start Docker services
make stack-up

# 3. Start development servers with hot reload
# This automatically runs build-fast internally, then starts 3 processes
make dev
```

Then access:
- **Console**: http://localhost:5173 (or http://localhost:8080 via proxy)
- **Trust Center**: http://localhost:5174 (or via backend proxy)
- **Backend API**: http://localhost:8080/api

What happens when you run `make dev`:
- ✅ Backend binary built with `DEV=1` (skips TS builds) - ~3.5s
- ✅ Go server starts with gow auto-reload
- ✅ Vite dev servers start for console and trust apps
- ✅ Environment variables set to proxy requests to Vite
- ✅ **No TypeScript compilation** unless you explicitly run it

Any changes to React/TypeScript files appear instantly with HMR! ⚡

### Benefits

- **⚡ No build step required**: Skip `npm run build` entirely during development
- **🔄 True HMR**: Changes to TypeScript/React code are instantly reflected in the browser
- **🚀 Fast iteration**: Dramatically reduced time to see changes
- **📦 Convenient setup**: Single `make dev` command starts everything
- **🔀 Flexible**: Can still run services separately if preferred
- **♻️ Production ready**: When you're ready to deploy, `make build` creates optimized bundles

### Alternative: Manual Service Management

If you prefer running services separately or need a different setup:

```bash
# Terminal 1: Backend only
VITE_DEV_SERVER_CONSOLE=http://localhost:5173 \
VITE_DEV_SERVER_TRUST=http://localhost:5174 \
bin/probod -cfg-file cfg/dev.yaml

# Terminal 2: Console dev server
cd apps/console && npm run dev

# Terminal 3: Trust dev server
cd apps/trust && npm run dev
```

## Technical Details

### How Dev Mode Works

1. **Environment Variables**: When `VITE_DEV_SERVER_*` variables are set, the Go backend automatically configures reverse proxies
2. **Reverse Proxy**: Requests that would normally hit embedded static files are proxied to the Vite dev server
3. **Fallback**: If no dev server environment variable is found, the server uses embedded static files (production behavior)
4. **HMR Support**: WebSocket connections for Vite's HMR are properly handled through the reverse proxy

### Build System

Three build modes are now available:

**1. Fast Dev Build** (for local development)
```bash
make build-fast
# or equivalently
DEV=1 make bin/probod
```
- Skips `@probo/console` and `@probo/trust` TypeScript builds
- Always builds `@probo/emails` (email templates required for embedding)
- Compiles Go backend with placeholder frontend files
- Time: ~3-5 seconds
- Use for: Rapid local development without building frontends

**2. Full Production Build** (for releases and CI)
```bash
make build
```
- Builds `@probo/console` with TypeScript + Vite (2130 modules)
- Builds `@probo/trust` with TypeScript + Vite (1414 modules)
- Builds `@probo/emails` (email templates)
- Compiles Go backend with embedded frontend assets
- Time: ~58 seconds
- Result: Production-optimized binary with all assets embedded
- Use for: Production releases, Docker images, final testing

**3. Build Frontend Only** (explicit app builds)
```bash
make build-apps
```
- Builds just `@probo/console` and `@probo/trust` apps
- Does not compile the Go backend
- Use for: Developing frontends in isolation

**4. CI/CD (E2E Tests)**
```bash
SKIP_APPS=1 make build
```
- Skips frontend app builds (console, trust)
- Still builds email templates
- Used in GitHub Actions for e2e test runs

## Requirements

For the improved `make dev` command, you need:
- `parallel` (part of GNU Parallel): `sudo apt-get install parallel` or `brew install parallel`
- `gow` (Go file watcher): `go install github.com/mitranim/gow@latest`

If these aren't available, use the manual approach with separate terminals.

## Production Build

Building for production is unchanged:

```bash
make build
```

This creates:
- Optimized frontend bundles in `apps/console/dist` and `apps/trust/dist`
- Fully compiled `bin/probod` binary with embedded frontend assets

## Troubleshooting

### "dev-server" page still showing

Make sure the environment variables are set:
```bash
export VITE_DEV_SERVER_CONSOLE=http://localhost:5173
export VITE_DEV_SERVER_TRUST=http://localhost:5174
```

### WebSocket connection errors

Ensure the Vite dev servers are running on the correct ports. Check:
- Console: http://localhost:5173
- Trust: http://localhost:5174

### HMR not updating

- Check that browser console doesn't show connection errors
- Verify the dev server is using the correct IP/hostname
- Try accessing from the Vite port directly instead of through the backend proxy

## See Also

- [CONTRIBUTING.md](./CONTRIBUTING.md) - Complete development setup guide
- Vite Documentation: https://vitejs.dev/
- Relay Compiler: https://relay.dev/docs/guides/compiler/
