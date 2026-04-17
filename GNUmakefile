MAKEFLAGS := --jobs=$(shell nproc)

CAT ?=	cat
CP ?=	cp
DOCKER ?=	docker
GO ?=	go
GRYPE ?=	grype
TRIVY ?=	trivy
MKCERT ?=	mkcert
MKDIR ?=	mkdir -p
NPM ?=	npm
NPX ?=	npx
OPENSSL ?=	openssl
SED ?= sed
SYFT ?=	syft
TAIL ?= tail
ECHO ?= echo
GOLINTCMD ?= golangci-lint

DOCKER_BUILD_FLAGS?=
DOCKER_BUILD=	DOCKER_BUILDKIT=1 $(DOCKER) build $(DOCKER_BUILD_FLAGS)

DOCKER_COMPOSE=	$(DOCKER) compose -f compose.yaml $(DOCKER_COMPOSE_FLAGS)

VERSION=	0.170.0
LDFLAGS=	-ldflags "-X 'main.version=$(VERSION)' -X 'main.env=prod'"
GCFLAGS=	-gcflags="-e"

CGO_ENABLED?=	0
GOOS?=

GO_BASE=	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go
GO_BUILD=	$(GO_BASE) build $(LDFLAGS) $(GCFLAGS)
GO_GENERATE=	$(GO_BASE) generate
GO_TEST=	$(GO_BASE) tool gotestsum -- $(TEST_FLAGS)
GO_VET=	$(GO_BASE) vet
GO_TOOL=	$(GO_BASE) tool

TEST_FLAGS?=	-race -cover -coverprofile=coverage.out

E2E_CONFIG ?= $(CURDIR)/e2e/console/testdata/config.yaml
E2E_COVER_DIR ?= $(CURDIR)/coverage/e2e

DOCKER_IMAGE_NAME=	ghcr.io/getprobo/probo
DOCKER_TAG_NAME?=	latest

PROBOD_BIN_EXTRA_DEPS=
PROBOD_BIN=	bin/probod
PROBOD_E2E_BIN=	bin/probod-e2e
PROBOD_SRC=	cmd/probod/main.go

PRB_BIN=	bin/prb
PRB_SRC=	cmd/prb/main.go

PROBOD_BOOTSTRAP_BIN=	bin/probod-bootstrap
PROBOD_BOOTSTRAP_SRC=	cmd/probod-bootstrap/main.go

ifndef SKIP_APPS
PROBOD_BIN_EXTRA_DEPS += \
	@probo/console \
	@probo/trust
endif

.PHONY: all
all: build

.PHONY: lint
lint: lint-go lint-js

.PHONY: lint-go
lint-go: vet go-fmt go-fix go-lint

.PHONY: lint-js
lint-js: npm-lint

.PHONY: vet
vet: generate apps/console/dist/index.html apps/trust/dist/index.html @probo/emails
	$(GO_VET) ./...

.PHONY: npm-lint
npm-lint:
	$(NPM) run lint

.PHONY: go-fmt
go-fmt:
	@output="$$(gofmt -l apps cmd packages pkg e2e)"; \
	if [ -n "$$output" ]; then \
		echo "error: 'gofmt' found unformatted files:"; \
		echo "$$output"; \
		exit 1; \
	fi

.PHONY: go-fix
go-fix: generate apps/console/dist/index.html apps/trust/dist/index.html @probo/emails
	@output="$$($(GO_BASE) fix -diff -omitzero=false ./apps/... ./cmd/... ./packages/... ./pkg/... ./e2e/...)"; \
	if [ -n "$$output" ]; then \
		echo "error: 'go fix' suggests changes; please apply them"; \
		echo "$$output"; \
		exit 1; \
	fi

.PHONY: go-lint
go-lint: generate
	$(GOLINTCMD) run ./...

.PHONY: test
test: generate
test: CGO_ENABLED=1
test: ## Run tests with race detection and coverage (usage: make test [MODULE=./pkg/some/module])
	$(GO_TEST) $(if $(MODULE),$(MODULE),$(shell $(GO) list ./... | grep -v /e2e/))

.PHONY: test-verbose
test-verbose: TEST_FLAGS+=-v
test-verbose: test ## Run tests with verbose output

.PHONY: test-short
test-short: TEST_FLAGS+=-short
test-short: test ## Run short tests only

.PHONY: coverage-report
coverage-report: test ## Generate HTML coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-bench
test-bench: TEST_FLAGS+=-bench=.
test-bench: test ## Run benchmark tests

.PHONY: test-e2e
test-e2e: CGO_ENABLED=1
test-e2e: bin/probod-e2e ## Run console e2e tests
	PROBO_E2E_BINARY=$(CURDIR)/bin/probod-e2e \
	PROBO_E2E_CONFIG=$(E2E_CONFIG) \
	GOTESTSUM_FORMAT=testname $(GO_TEST) -count=1 ./e2e/console/...

bin/probod-coverage:
	CGO_ENABLED=0 $(GO_BUILD) -tags=e2e -cover -o $@ $(PROBOD_SRC)

.PHONY: test-e2e-coverage
test-e2e-coverage: bin/probod-coverage ## Run e2e tests with coverage
	@$(RM) -rf $(E2E_COVER_DIR) && $(MKDIR) -p $(E2E_COVER_DIR)
	PROBO_E2E_BINARY=$(CURDIR)/bin/probod-coverage \
	PROBO_E2E_COVERDIR=$(E2E_COVER_DIR) \
	PROBO_E2E_CONFIG=$(E2E_CONFIG) \
	CGO_ENABLED=1 $(GO) test -count=1 -v ./e2e/console/...
	$(GO) tool covdata textfmt -i=$(E2E_COVER_DIR) -o=coverage-e2e.out
	$(GO) tool cover -html=coverage-e2e.out -o=coverage-e2e.html

.PHONY: coverage-combined
coverage-combined: coverage-report test-e2e-coverage ## Generate combined coverage report (unit + e2e)
	@$(CAT) coverage.out > coverage-combined.out
	@$(TAIL) -n +2 coverage-e2e.out >> coverage-combined.out
	$(GO) tool cover -html=coverage-combined.out -o=coverage-combined.html

.PHONY: build
build: bin/probod bin/prb bin/probod-bootstrap

.PHONY: sbom-docker
sbom-docker: docker-build
	$(SYFT) docker:$(DOCKER_IMAGE_NAME):$(DOCKER_TAG_NAME) -o cyclonedx-json \
		--source-name "$(DOCKER_IMAGE_NAME)" \
		--source-version "$(DOCKER_TAG_NAME)" \
		> sbom-docker.json

.PHONY: sbom
sbom:
	$(SYFT) dir:. -o cyclonedx-json \
		--source-name "probo" \
		--source-version "$(VERSION)" \
		> sbom.json

.PHONY: scan-sbom
scan-sbom: sbom
	$(GRYPE) sbom:sbom.json --config .grype.yaml --fail-on high

.PHONY: scan-sbom-docker
scan-sbom-docker: sbom-docker
	$(GRYPE) sbom:sbom-docker.json --config .grype.yaml --fail-on high

.PHONY: scan-docker
scan-docker: docker-build
	$(GRYPE) docker:$(DOCKER_IMAGE_NAME):$(DOCKER_TAG_NAME) --config .grype.yaml --fail-on high

.PHONY: scan
scan: scan-sbom scan-sbom-docker scan-docker

.PHONY: scan-license
scan-license: ## Check dependencies licenses compliance
	$(TRIVY) fs --license-full --scanners license --ignorefile .trivyignore.yaml --severity UNKNOWN,HIGH,CRITICAL --exit-code 1 .

.PHONY: docker-build
docker-build:
	$(DOCKER_BUILD) --tag $(DOCKER_IMAGE_NAME):$(DOCKER_TAG_NAME) --file Dockerfile .

.PHONY: bin/probod
bin/probod: pkg/server/api/connect/v1/schema/schema.go \
	pkg/server/api/connect/v1/types/types.go \
	pkg/server/api/console/v1/schema/schema.go \
	pkg/server/api/console/v1/types/types.go \
	pkg/server/api/trust/v1/schema/schema.go \
	pkg/server/api/trust/v1/types/types.go \
	pkg/server/api/mcp/v1/server/server.go \
	pkg/server/api/mcp/v1/types/types.go \
	apps/console/dist/index.html \
	apps/trust/dist/index.html \
	$(PROBOD_BIN_EXTRA_DEPS) \
	@probo/emails
	$(GO_BUILD) -o $(PROBOD_BIN) $(PROBOD_SRC)

# probod built with -tags=e2e. The tag swaps the real vendor assessor for
# a deterministic stub so e2e tests avoid the real LLM/browser pipeline.
# Never ship this binary.
.PHONY: bin/probod-e2e
bin/probod-e2e: pkg/server/api/connect/v1/schema/schema.go \
	pkg/server/api/connect/v1/types/types.go \
	pkg/server/api/console/v1/schema/schema.go \
	pkg/server/api/console/v1/types/types.go \
	pkg/server/api/trust/v1/schema/schema.go \
	pkg/server/api/trust/v1/types/types.go \
	pkg/server/api/mcp/v1/server/server.go \
	pkg/server/api/mcp/v1/types/types.go \
	apps/console/dist/index.html \
	apps/trust/dist/index.html \
	$(PROBOD_BIN_EXTRA_DEPS) \
	@probo/emails
	$(GO_BUILD) -tags=e2e -o $(PROBOD_E2E_BIN) $(PROBOD_SRC)

.PHONY: bin/prb
bin/prb:
	$(GO_BUILD) -o $(PRB_BIN) $(PRB_SRC)

.PHONY: bin/probod-bootstrap
bin/probod-bootstrap:
	$(GO_BUILD) -o $(PROBOD_BOOTSTRAP_BIN) $(PROBOD_BOOTSTRAP_SRC)

.PHONY: @probo/emails
@probo/emails:
	$(NPM) --workspace $@ run build

RELAY_SCHEMAS = \
	pkg/server/api/connect/v1/schema.graphql \
	pkg/server/api/console/v1/schema.graphql \
	pkg/server/api/trust/v1/schema.graphql

.PHONY: relay
relay: $(RELAY_SCHEMAS)
	$(NPX) relay-compiler

MERGE_GRAPHQL = contrib/merge-graphql-schema.sh

CONNECT_GQL = $(wildcard pkg/server/api/connect/v1/graphql/*.graphql)
CONSOLE_GQL = $(wildcard pkg/server/api/console/v1/graphql/*.graphql)
TRUST_GQL   = $(wildcard pkg/server/api/trust/v1/graphql/*.graphql)

pkg/server/api/connect/v1/schema.graphql: pkg/server/api/connect/v1/graphql $(CONNECT_GQL)
	$(MERGE_GRAPHQL) $@ pkg/server/api/connect/v1/graphql

pkg/server/api/console/v1/schema.graphql: pkg/server/api/console/v1/graphql $(CONSOLE_GQL)
	$(MERGE_GRAPHQL) $@ pkg/server/api/console/v1/graphql

pkg/server/api/trust/v1/schema.graphql: pkg/server/api/trust/v1/graphql $(TRUST_GQL)
	$(MERGE_GRAPHQL) $@ pkg/server/api/trust/v1/graphql

.PHONY: @probo/console
@probo/console: NODE_ENV=production
@probo/console: relay
	$(NPM) --workspace $@ run check
	$(NPM) --workspace $@ run build

.PHONY: @probo/trust
@probo/trust: NODE_ENV=production
@probo/trust: relay
	$(NPM) --workspace $@ run check
	$(NPM) --workspace $@ run build

.PHONY: generate
generate: pkg/server/api/connect/v1/schema/schema.go \
	pkg/server/api/connect/v1/types/types.go \
	pkg/server/api/console/v1/schema/schema.go \
	pkg/server/api/console/v1/types/types.go \
	pkg/server/api/trust/v1/schema/schema.go \
	pkg/server/api/trust/v1/types/types.go \
	pkg/server/api/mcp/v1/server/server.go \
	pkg/server/api/mcp/v1/types/types.go \
	relay

pkg/server/api/connect/v1/schema/schema.go \
pkg/server/api/connect/v1/types/types.go: pkg/server/api/connect/v1/gqlgen.yaml pkg/server/api/connect/v1/graphql $(CONNECT_GQL)
	$(GO_GENERATE) ./pkg/server/api/connect/v1

pkg/server/api/console/v1/schema/schema.go \
pkg/server/api/console/v1/types/types.go: pkg/server/api/console/v1/gqlgen.yaml pkg/server/api/console/v1/graphql $(CONSOLE_GQL)
	$(GO_GENERATE) ./pkg/server/api/console/v1

pkg/server/api/trust/v1/schema/schema.go \
pkg/server/api/trust/v1/types/types.go: pkg/server/api/trust/v1/gqlgen.yaml pkg/server/api/trust/v1/graphql $(TRUST_GQL)
	$(GO_GENERATE) ./pkg/server/api/trust/v1

pkg/server/api/mcp/v1/server/server.go \
pkg/server/api/mcp/v1/types/types.go: pkg/server/api/mcp/v1/specification.yaml pkg/server/api/mcp/v1/mcpgen.yaml
	$(GO_GENERATE) ./pkg/server/api/mcp/v1

.PHONY: genmodels
genmodels: ## Refresh LLM model registry from OpenRouter
	$(GO_GENERATE) ./pkg/llm

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: fmt
fmt: fmt-go ## Format Go code

.PHONY: fmt-go
fmt-go: ## Format Go code
	go fmt ./...

.PHONY: clean
clean: ## Clean the project (node_modules and build artifacts)
	$(RM) -rf bin/*
	$(RM) -rf node_modules
	$(RM) -rf apps/{console,trust}/{dist,node_modules}
	$(RM) -rf packages/emails/{dist,node_modules}
	$(RM) -rf sbom-docker.json sbom.json
	$(RM) -rf coverage.out coverage.html coverage-e2e.out coverage-e2e.html coverage-combined.out coverage-combined.html
	$(RM) -rf coverage/
	$(RM) -rf compose/keycloak/certs/cert.pem compose/keycloak/certs/private-key.pem compose/keycloak/probo-realm.json
	$(RM) -f pkg/server/api/connect/v1/schema/schema.go pkg/server/api/connect/v1/types/types.go
	$(RM) -f pkg/server/api/console/v1/schema/schema.go pkg/server/api/console/v1/types/types.go
	$(RM) -f pkg/server/api/trust/v1/schema/schema.go pkg/server/api/trust/v1/types/types.go
	$(RM) -f pkg/server/api/mcp/v1/server/server.go pkg/server/api/mcp/v1/types/types.go
	$(RM) -f $(RELAY_SCHEMAS)
	$(RM) -f pkg/llm/registry_gen.go
	find apps -type d -name __generated__ -exec $(RM) -rf {} +

.PHONY: stack-up
stack-up: compose/pebble/certs/rootCA.pem compose/keycloak/probo-realm.json ## Start the docker stack as a deamon
	$(DOCKER_COMPOSE) up -d

.PHONY: stack-down
stack-down: ## Stop the docker stack
	$(DOCKER_COMPOSE) down

.PHONY: stack-ps
stack-ps: ## List the docker stack containers
	$(DOCKER_COMPOSE) ps

.PHONY: psql
psql: ## Open a psql shell to the postgres container
	$(DOCKER_COMPOSE) exec postgres psql -U probod -d probod

compose/pebble/certs/rootCA.pem:
	@$(MKDIR) compose/pebble/certs
	$(MKCERT) -cert-file compose/pebble/certs/pebble.crt \
		-key-file compose/pebble/certs/pebble.key \
		localhost 127.0.0.1 ::1 pebble
	$(CP) "$$($(MKCERT) -CAROOT)/rootCA.pem" compose/pebble/certs/rootCA.pem
	$(CP) "$$($(MKCERT) -CAROOT)/rootCA-key.pem" compose/pebble/certs/rootCA-key.pem

compose/keycloak/certs/cert.pem:
	$(MKDIR) ./compose/keycloak/certs
	$(OPENSSL) req -x509 -newkey rsa:2048 -keyout compose/keycloak/certs/private-key.pem -out compose/keycloak/certs/cert.pem -days 3650 -nodes -subj "/CN=keycloak-saml-signing"

compose/keycloak/probo-realm.json: compose/keycloak/probo-realm.json.tmpl compose/keycloak/certs/cert.pem
	$(SED) \
	-e "s|CERTIFICATE_PLACEHOLDER|$$(awk 'NR==1 {printf "%s", $$0; next} {printf "\\\\n%s", $$0}' compose/keycloak/certs/cert.pem)|g" \
	-e "s|PRIVATE_KEY_PLACEHOLDER|$$(awk 'NR==1 {printf "%s", $$0; next} {printf "\\\\n%s", $$0}' compose/keycloak/certs/private-key.pem)|g" \
	$@.tmpl > $@

apps/console/dist/index.html apps/trust/dist/index.html:
	$(MKDIR) $(dir $@)
	$(ECHO) dev-server > $@


.PHONY: sandbox-create
sandbox-create: ## Create a Lima sandbox VM for this worktree
	./contrib/lima/sandbox.sh create

.PHONY: sandbox-start
sandbox-start: ## Start the Lima sandbox VM
	./contrib/lima/sandbox.sh start

.PHONY: sandbox-stop
sandbox-stop: ## Stop (hibernate) the Lima sandbox VM
	./contrib/lima/sandbox.sh stop

.PHONY: sandbox-delete
sandbox-delete: ## Delete the Lima sandbox VM
	./contrib/lima/sandbox.sh delete

.PHONY: sandbox-ssh
sandbox-ssh: ## Open a shell in the Lima sandbox VM
	./contrib/lima/sandbox.sh ssh

.PHONY: sandbox-status
sandbox-status: ## Show Lima sandbox VM status and IP
	./contrib/lima/sandbox.sh status

.PHONY: deadcode
deadcode:
	$(GO_TOOL) deadcode ./... | grep -v "With" | grep -v "UnmarshalBigIntScalar" | grep -v "^e2e/"
