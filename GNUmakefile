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

VERSION=	0.127.1
LDFLAGS=	-ldflags "-X 'main.version=$(VERSION)' -X 'main.env=prod'"
GCFLAGS=	-gcflags="-e"

CGO_ENABLED?=	0
GOOS?=

GO_BASE=	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go
GO_BUILD=	$(GO_BASE) build $(LDFLAGS) $(GCFLAGS)
GO_GENERATE=	$(GO_BASE) generate
GO_TEST=	$(GO_BASE) tool gotestsum -- $(TEST_FLAGS)
GO_VET=	$(GO_BASE) vet

TEST_FLAGS?=	-race -cover -coverprofile=coverage.out

E2E_CONFIG ?= $(CURDIR)/e2e/console/testdata/config.yaml
E2E_COVER_DIR ?= $(CURDIR)/coverage/e2e

DOCKER_IMAGE_NAME=	ghcr.io/getprobo/probo
DOCKER_TAG_NAME?=	latest

PROBOD_BIN_EXTRA_DEPS=
PROBOD_BIN=	bin/probod
PROBOD_SRC=	cmd/probod/main.go

ifndef SKIP_APPS
PROBOD_BIN_EXTRA_DEPS += \
	@probo/console \
	@probo/trust
endif

.PHONY: all
all: build

.PHONY: lint
lint: vet go-lint npm-lint

.PHONY: vet
vet: @probo/emails
	$(GO_VET) ./...

.PHONY: npm-lint
npm-lint:
	$(NPM) run lint

.PHONY: go-lint
go-lint:
	$(GOLINTCMD) run ./...

.PHONY: test
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
test-e2e: bin/probod ## Run console e2e tests
	PROBO_E2E_BINARY=$(CURDIR)/bin/probod \
	PROBO_E2E_CONFIG=$(E2E_CONFIG) \
	GOTESTSUM_FORMAT=testname $(GO_TEST) -count=1 ./e2e/console/...

bin/probod-coverage:
	CGO_ENABLED=0 $(GO_BUILD) -cover -o $@ $(PROBOD_SRC)

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
build: bin/probod

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
	pkg/server/api/connect/v1/v1_resolver.go \
	pkg/server/api/console/v1/schema/schema.go \
	pkg/server/api/console/v1/types/types.go \
	pkg/server/api/console/v1/v1_resolver.go \
	pkg/server/api/trust/v1/schema/schema.go \
	pkg/server/api/trust/v1/types/types.go \
	pkg/server/api/trust/v1/v1_resolver.go \
	pkg/server/api/mcp/v1/server/server.go \
	pkg/server/api/mcp/v1/types/types.go \
	apps/console/dist/index.html \
	apps/trust/dist/index.html \
	$(PROBOD_BIN_EXTRA_DEPS) \
	@probo/emails
	$(GO_BUILD) -o $(PROBOD_BIN) $(PROBOD_SRC)

.PHONY: @probo/emails
@probo/emails:
	$(NPM) --workspace $@ run build

.PHONY: @probo/console
@probo/console: NODE_ENV=production
@probo/console:
	$(NPM) --workspace $@ run check
	$(NPM) --workspace $@ run build

.PHONY: @probo/trust
@probo/trust: NODE_ENV=production
@probo/trust:
	$(NPM) --workspace $@ run check
	$(NPM) --workspace $@ run build

pkg/server/api/connect/v1/schema/schema.go \
pkg/server/api/connect/v1/types/types.go \
pkg/server/api/connect/v1/v1_resolver.go: pkg/server/api/connect/v1/gqlgen.yaml pkg/server/api/connect/v1/schema.graphql
	$(GO_GENERATE) ./pkg/server/api/connect/v1

pkg/server/api/console/v1/schema/schema.go \
pkg/server/api/console/v1/types/types.go \
pkg/server/api/console/v1/v1_resolver.go: pkg/server/api/console/v1/gqlgen.yaml pkg/server/api/console/v1/schema.graphql
	$(GO_GENERATE) ./pkg/server/api/console/v1

pkg/server/api/trust/v1/schema/schema.go \
pkg/server/api/trust/v1/types/types.go \
pkg/server/api/trust/v1/v1_resolver.go: pkg/server/api/trust/v1/gqlgen.yaml pkg/server/api/trust/v1/schema.graphql
	$(GO_GENERATE) ./pkg/server/api/trust/v1

pkg/server/api/mcp/v1/server/server.go \
pkg/server/api/mcp/v1/types/types.go: pkg/server/api/mcp/v1/specification.yaml pkg/server/api/mcp/v1/mcpgen.yaml
	$(GO_GENERATE) ./pkg/server/api/mcp/v1

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY:dev
dev: ## Start the development server
	parallel -j 2 --line-buffer ::: "gow -r=false run cmd/probod/main.go" "cd apps/console && npm run dev"

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

.PHONY: goreleaser-snapshot
goreleaser-snapshot: ## Build a snapshot release with goreleaser
	goreleaser release --snapshot --clean --config .goreleaser.yaml

.PHONY: goreleaser-check
goreleaser-check: ## Check goreleaser configuration
	goreleaser check

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
