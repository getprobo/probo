# GNUmakefile

The project uses a `GNUmakefile` at the root. Builds run with `--jobs=$(nproc)` by default.

## Everyday targets

| Target | Purpose |
|---|---|
| `make build` | Build `bin/probod`, `bin/prb`, and `bin/probod-bootstrap` (includes frontend apps, codegen, and Relay) |
| `SKIP_APPS=1 make build` | Build without frontend apps (faster for backend-only work) |
| `make test` | Run tests with race detection and coverage |
| `make test MODULE=./pkg/foo` | Run tests for a single module |
| `make test-verbose` | Tests with verbose output |
| `make test-short` | Short tests only |
| `make test-bench` | Run benchmarks |
| `make test-e2e` | Run console end-to-end tests (requires `bin/probod`) |
| `make lint` | Run all linters: `vet` + `go-fmt` + `go-fix` + `go-lint` + `npm-lint` |
| `make fmt` | Format Go code (`go fmt ./...`) |
| `make clean` | Remove all build artifacts, `node_modules`, generated files, and coverage |
| `make help` | List targets with `##` doc comments |

## Infrastructure

| Target | Purpose |
|---|---|
| `make stack-up` | Start Docker Compose infra (Postgres, Pebble, Keycloak, etc.) |
| `make stack-down` | Stop Docker Compose infra |
| `make stack-ps` | List running containers |
| `make psql` | Open a `psql` shell to the dev Postgres database |

## Codegen

`make generate` runs all code generation (GraphQL + MCP + Relay). Individual codegen is driven by `go generate`:

- `go generate ./pkg/server/api/console/v1` — Console GraphQL (gqlgen)
- `go generate ./pkg/server/api/connect/v1` — Connect GraphQL (gqlgen)
- `go generate ./pkg/server/api/trust/v1` — Trust GraphQL (gqlgen)
- `go generate ./pkg/server/api/mcp/v1` — MCP (mcpgen)
- `go generate ./pkg/llm` — LLM model registry from OpenRouter (`make genmodels`)

`make relay` merges split `.graphql` schema files and runs `relay-compiler`.

## Coverage

| Target | Purpose |
|---|---|
| `make coverage-report` | Unit test HTML coverage report (`coverage.html`) |
| `make test-e2e-coverage` | E2E coverage report (`coverage-e2e.html`) |
| `make coverage-combined` | Combined unit + e2e report (`coverage-combined.html`) |

## Docker

| Target | Purpose |
|---|---|
| `make docker-build` | Build the Docker image (`ghcr.io/getprobo/probo`) |
| `make sbom` | Source SBOM (CycloneDX) |
| `make sbom-docker` | Docker image SBOM |
| `make scan` | Vulnerability scan (Grype) on source + Docker |
| `make scan-license` | License compliance scan (Trivy) |

## Sandbox (Lima)

| Target | Purpose |
|---|---|
| `make sandbox-create` | Create a Lima sandbox VM for this worktree |
| `make sandbox-start` | Start the VM |
| `make sandbox-stop` | Stop (hibernate) the VM |
| `make sandbox-delete` | Delete the VM |
| `make sandbox-ssh` | Open a shell in the VM |
| `make sandbox-status` | Show VM status and IP |

## Overridable variables

| Variable | Default | Purpose |
|---|---|---|
| `SKIP_APPS` | (unset) | Set to `1` to skip frontend app builds |
| `CGO_ENABLED` | `0` | Enable/disable CGO |
| `GOOS` | (host) | Cross-compile target OS |
| `TEST_FLAGS` | `-race -cover -coverprofile=coverage.out` | Extra flags passed to `go test` |
| `DOCKER_BUILD_FLAGS` | (empty) | Extra flags for `docker build` |
| `E2E_CONFIG` | `e2e/console/testdata/config.yaml` | E2E test config path |
