# GNUmakefile

The project uses a `GNUmakefile` at the root. Builds run with `--jobs=$(nproc)` by default.

## Everyday targets

| Target                       | Purpose                                                                                                |
| ---------------------------- | ------------------------------------------------------------------------------------------------------ |
| `make build`                 | Build `bin/probod`, `bin/prb`, and `bin/probod-bootstrap` (does not include frontend apps and Relay)   |
| `make build WITH_APPS=1`     | Build `bin/probod`, `bin/prb`, and `bin/probod-bootstrap` (includes frontend apps, codegen, and Relay) |
| `make test`                  | Run tests with race detection and coverage                                                             |
| `make test MODULE=./pkg/foo` | Run tests for a single module                                                                          |
| `make test-verbose`          | Tests with verbose output                                                                              |
| `make test-short`            | Short tests only                                                                                       |
| `make test-bench`            | Run benchmarks                                                                                         |
| `make test-e2e`              | Run console end-to-end tests (requires `bin/probod`)                                                   |
| `make lint`                  | Run Go + JS linters: `vet` + `go-fmt` + `go-fix` + `go-lint` + `lint-js`                              |
| `make lint-swift`            | Opt-in: lint Swift enroll-ui (`swift-fmt` + `swift-lint`; needs Swift + SwiftLint; CI runs this on Linux) |
| `make fmt`                   | Format Go code                                                                                         |
| `make fmt-swift`             | Opt-in: format Swift enroll-ui (`swift format` + SwiftLint `--fix`; needs Swift)                     |
| `make clean`                 | Remove all build artifacts, `node_modules`, generated files, and coverage                              |
| `make help`                  | List targets with `##` doc comments                                                                    |

## Infrastructure

| Target            | Purpose                                                       |
| ----------------- | ------------------------------------------------------------- |
| `make stack-up`   | Start Docker Compose infra (Postgres, step-ca, Keycloak, etc.) |
| `make stack-down` | Stop Docker Compose infra                                     |
| `make stack-ps`   | List running containers                                       |
| `make psql`       | Open a `psql` shell to the dev Postgres database              |

## Codegen

`make generate` runs go code generation (GraphQL + MCP without Relay).
`make generate WITH_APPS=1` runs all code generation (GraphQL + MCP + Relay).

Individual codegen is driven by `go generate`:

- `go generate ./pkg/server/api/console/v1` — Console GraphQL (gqlgen)
- `go generate ./pkg/server/api/connect/v1` — Connect GraphQL (gqlgen)
- `go generate ./pkg/server/api/complianceportal/v1` — Compliance portal GraphQL (gqlgen)
- `go generate ./pkg/server/api/mcp/v1` — MCP (mcpgen)
- `go generate ./pkg/llm` — LLM model registry from OpenRouter (`make genmodels`)

`make relay` merges each service's split `.graphql` schema files into a single `schema.graphql` (via `contrib/merge-graphql-schema.sh`) and runs `relay-compiler`. The merge is required: relay-compiler's `schema` must be a single file, and `schemaExtensions` would mark the fields as client-only (emitting `text: null`), so the split files cannot be fed to Relay directly.

## Coverage

| Target                   | Purpose                                               |
| ------------------------ | ----------------------------------------------------- |
| `make coverage-report`   | Unit test HTML coverage report (`coverage.html`)      |
| `make test-e2e-coverage` | E2E coverage report (`coverage-e2e.html`)             |
| `make coverage-combined` | Combined unit + e2e report (`coverage-combined.html`) |

## Docker

| Target              | Purpose                                           |
| ------------------- | ------------------------------------------------- |
| `make docker-build` | Build the Docker image (`artifact.probo.inc/probo/probo`) |
| `make sbom`         | Source SBOM (CycloneDX)                           |
| `make sbom-docker`  | Docker image SBOM                                 |
| `make scan`         | Vulnerability scan (Grype) on source + Docker     |
| `make scan-license` | License compliance scan (Trivy)                   |

## Sandbox (Lima)

| Target                | Purpose                                    |
| --------------------- | ------------------------------------------ |
| `make sandbox-create` | Create a Lima sandbox VM for this worktree |
| `make sandbox-start`  | Start the VM                               |
| `make sandbox-stop`   | Stop (hibernate) the VM                    |
| `make sandbox-delete` | Delete the VM                              |
| `make sandbox-ssh`    | Open a shell in the VM                     |
| `make sandbox-status` | Show VM status and IP                      |

## Overridable variables

| Variable             | Default                                   | Purpose                                    |
| -------------------- | ----------------------------------------- | ------------------------------------------ |
| `WITH_APPS`          | (unset)                                   | Set to `1` to generate/build frontend apps |
| `CGO_ENABLED`        | `0`                                       | Enable/disable CGO                         |
| `GOOS`               | (host)                                    | Cross-compile target OS                    |
| `TEST_FLAGS`         | `-race -cover -coverprofile=coverage.out` | Extra flags passed to `go test`            |
| `DOCKER_BUILD_FLAGS` | (empty)                                   | Extra flags for `docker build`             |
| `SWIFTLINTCMD`       | `swiftlint`                               | SwiftLint binary                           |
| `SWIFTCMD`           | `swift`                                   | Swift toolchain binary (`swift format`)    |
| `SWIFT_ENROLL_UI`    | `cmd/probo-agent/installer/macos/enroll-ui` | Path to the Swift SPM package            |
