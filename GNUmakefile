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

PRB_VERSION=             $(shell cat cmd/prb/VERSION)
PROBOD_VERSION=          $(shell cat cmd/probod/VERSION)
PROBOD_BOOTSTRAP_VERSION=$(shell cat cmd/probod-bootstrap/VERSION)
PROBOCTL_VERSION=        $(shell cat cmd/proboctl/VERSION)
PROBO_AGENT_VERSION=     $(shell cat cmd/probo-agent/VERSION)

PRB_LDFLAGS=             -ldflags "-X 'main.version=$(PRB_VERSION)'"
PROBOD_LDFLAGS=          -ldflags "-X 'main.version=$(PROBOD_VERSION)' -X 'main.env=prod'"
PROBOD_BOOTSTRAP_LDFLAGS=-ldflags "-X 'main.version=$(PROBOD_BOOTSTRAP_VERSION)'"
PROBOCTL_LDFLAGS=        -ldflags "-X 'main.version=$(PROBOCTL_VERSION)'"
PROBO_AGENT_LDFLAGS=     -ldflags "-X 'main.version=$(PROBO_AGENT_VERSION)'"

GCFLAGS=	-gcflags="-e"

CGO_ENABLED?=	0
GOOS?=

GO_BASE=	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go
GO_BUILD=	$(GO_BASE) build $(GCFLAGS)
GO_GENERATE=	$(GO_BASE) generate
GO_TEST=	$(GO_BASE) tool gotestsum -- $(TEST_FLAGS)
GO_VET=	$(GO_BASE) vet
GO_TOOL=	$(GO_BASE) tool

TEST_FLAGS?=	-race -cover -coverprofile=coverage.out

E2E_CONFIG ?= $(CURDIR)/e2e/console/testdata/config.yaml
E2E_COVER_DIR ?= $(CURDIR)/coverage/e2e

DOCKER_REGISTRY=	artifact.probo.inc
DOCKER_PROXY=		$(DOCKER_REGISTRY)/dockerhub
DOCKER_BASE_DIGEST=	sha256:c4a8d5503dfb2a3eb8ab5f807da5bc69a85730fb49b5cfca2330194ebcc41c7b
DOCKER_BASE_IMAGE=	ubuntu:24.04@$(DOCKER_BASE_DIGEST)
# Harbor proxy resolves digest refs as library/<name>@sha256:..., not library/<name>:tag@sha256:...
DOCKER_PROXY_BASE_IMAGE=	$(DOCKER_PROXY)/library/ubuntu@$(DOCKER_BASE_DIGEST)
DOCKER_IMAGE_NAME=	$(DOCKER_REGISTRY)/probo/probo
HELM_CHART_OCI=		oci://$(DOCKER_REGISTRY)/probo
DOCKER_TAG_NAME?=	latest

GENERATED= pkg/server/api/connect/v1/schema/schema.go \
	pkg/server/api/connect/v1/types/types.go \
	pkg/server/api/console/v1/schema/schema.go \
	pkg/server/api/console/v1/types/types.go \
	pkg/server/api/complianceportal/v1/schema/schema.go \
	pkg/server/api/complianceportal/v1/types/types.go \
	pkg/server/api/mcp/v1/server/server.go \
	pkg/server/api/mcp/v1/types/types.go

EMBEDDED= apps/console/dist/index.html \
	apps/compliance-portal/dist/index.html \
	@probo/emails

PROBOD_BIN_EXTRA_DEPS=
PROBOD_BIN=	bin/probod
PROBOD_SRC=	cmd/probod/main.go

PRB_BIN=	bin/prb
PRB_SRC=	cmd/prb/main.go

PROBOD_BOOTSTRAP_BIN=	bin/probod-bootstrap
PROBOD_BOOTSTRAP_SRC=	cmd/probod-bootstrap/main.go

PROBOCTL_BIN=	bin/proboctl
PROBOCTL_SRC=	cmd/proboctl/main.go

PROBO_AGENT_BIN=	bin/probo-agent
PROBO_AGENT_SRC=	cmd/probo-agent/main.go

ifdef WITH_APPS
GENERATED += relay
EMBEDDED += \
	@probo/console \
	@probo/compliance-portal
endif

.PHONY: all
all: build

.PHONY: lint
lint: lint-go lint-js

.PHONY: lint-go
lint-go: vet go-fmt go-fix go-lint

.PHONY: lint-js
lint-js:
	$(NPM) run lint

.PHONY: vet
vet: generate embed
	$(GO_VET) ./...

.PHONY: go-fmt
go-fmt:
	@output="$$(gofmt -l apps cmd packages pkg e2e)"; \
	if [ -n "$$output" ]; then \
		echo "error: 'gofmt' found unformatted files:"; \
		echo "$$output"; \
		exit 1; \
	fi

.PHONY: go-fix
go-fix: generate embed
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
test-e2e: $(PROBOD_BIN) ## Run console e2e tests
	PROBO_E2E_BINARY=$(CURDIR)/$(PROBOD_BIN) \
	PROBO_E2E_CONFIG=$(E2E_CONFIG) \
	GOTESTSUM_FORMAT=testname $(GO_TEST) -count=1 ./e2e/console/...

bin/probod-coverage:
	CGO_ENABLED=0 $(GO_BUILD) $(PROBOD_LDFLAGS) -cover -o $@ $(PROBOD_SRC)

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
build: $(PROBOD_BIN) bin/prb bin/probod-bootstrap bin/proboctl $(PROBO_AGENT_BIN)

CFG_DEV_OAUTH2_KEY = cfg/.dev-oauth2-signing-key.pem
DEV_ENV            = .env

.PHONY: dev-config
dev-config: cfg/dev.yaml ## Generate cfg/dev.yaml via probod-bootstrap (picks up edits to .env)

$(CFG_DEV_OAUTH2_KEY):
	@$(MKDIR) $(@D)
	$(OPENSSL) genrsa -out $@ 2048

cfg/dev.yaml: bin/probod-bootstrap $(CFG_DEV_OAUTH2_KEY) compose/step-ca/certs/root_ca.crt $(wildcard $(DEV_ENV))
	@$(MKDIR) $(@D)
	set -a; \
	PROBOD_BASE_URL=http://localhost:8080; \
	PROBOD_API_ADDR=:8080; \
	PROBOD_ENCRYPTION_KEY="AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; \
	PROBOD_AUTH_COOKIE_NAME=SSID; \
	PROBOD_AUTH_COOKIE_DOMAIN=localhost; \
	PROBOD_AUTH_COOKIE_SECRET="this-is-a-secure-secret-for-cookie-signing-at-least-32-bytes"; \
	PROBOD_AUTH_PASSWORD_PEPPER="this-is-a-secure-pepper-for-password-hashing-at-least-32-bytes"; \
	PROBOD_AUTH_COOKIE_SECURE=false; \
	PROBOD_OAUTH2_SERVER_SIGNING_KEY="$$($(CAT) $(CFG_DEV_OAUTH2_KEY))"; \
	PROBOD_API_CORS_ALLOWED_ORIGINS="http://localhost:8080,http://localhost:5173,http://localhost:5174"; \
	PROBOD_PG_ADDR=localhost:5432; \
	PROBOD_PG_USERNAME=postgres; \
	PROBOD_PG_PASSWORD=postgres; \
	PROBOD_PG_DATABASE=probod; \
	PROBOD_TRUST_CENTER_HTTP_ADDR=:10080; \
	PROBOD_TRUST_CENTER_HTTPS_ADDR=:443; \
	PROBOD_TRUST_CENTER_BASE_DOMAIN=probopage.localhost; \
	PROBOD_AWS_REGION=us-east-1; \
	PROBOD_AWS_BUCKET=probod; \
	PROBOD_AWS_ACCESS_KEY_ID=probod; \
	PROBOD_AWS_SECRET_ACCESS_KEY=thisisnotasecret; \
	PROBOD_AWS_ENDPOINT=http://127.0.0.1:8333; \
	PROBOD_SMTP_ADDR=localhost:1025; \
	PROBOD_MAILER_SENDER_EMAIL=no-reply@notification.getprobo.com; \
	PROBOD_MAILER_SENDER_NAME=Probo; \
	PROBOD_OPENAI_API_KEY=thisisnotasecret; \
	PROBOD_AGENT_THIRD_PARTY_VETTER_PROVIDER=openai; \
	PROBOD_AGENT_THIRD_PARTY_VETTER_MODEL_NAME=gpt-4o; \
	PROBOD_CHROME_DP_ADDR=localhost:9222; \
	PROBOD_ACME_DIRECTORY=https://localhost:9000/acme/acme/directory; \
	PROBOD_ACME_EMAIL=admin@probo.com; \
	PROBOD_ACME_ROOT_CA="$$($(CAT) compose/step-ca/certs/root_ca.crt)"; \
	if [ -f $(DEV_ENV) ]; then . $(DEV_ENV); fi; \
	set +a; \
	./bin/probod-bootstrap -output $@

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
		--source-version "$(PROBOD_VERSION)" \
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

.PHONY: $(PROBOD_BIN)
$(PROBOD_BIN): generate embed
	$(GO_BUILD) $(PROBOD_LDFLAGS) -o $(PROBOD_BIN) $(PROBOD_SRC)

.PHONY: bin/prb
bin/prb:
	$(GO_BUILD) $(PRB_LDFLAGS) -o $(PRB_BIN) $(PRB_SRC)

.PHONY: $(PROBOD_BOOTSTRAP_BIN)
$(PROBOD_BOOTSTRAP_BIN):
	$(GO_BUILD) $(PROBOD_BOOTSTRAP_LDFLAGS) -o $(PROBOD_BOOTSTRAP_BIN) $(PROBOD_BOOTSTRAP_SRC)

.PHONY: bin/proboctl
bin/proboctl:
	$(GO_BUILD) $(PROBOCTL_LDFLAGS) -o $(PROBOCTL_BIN) $(PROBOCTL_SRC)

.PHONY: $(PROBO_AGENT_BIN)
$(PROBO_AGENT_BIN):
	$(GO_BUILD) $(PROBO_AGENT_LDFLAGS) -o $(PROBO_AGENT_BIN) $(PROBO_AGENT_SRC)

.PHONY: @probo/emails
@probo/emails:
	$(NPM) --workspace $@ run build

RELAY_SCHEMAS = \
	pkg/server/api/connect/v1/schema.graphql \
	pkg/server/api/console/v1/schema.graphql \
	pkg/server/api/complianceportal/v1/schema.graphql

.PHONY: relay
relay: $(RELAY_SCHEMAS)
	$(NPM) run relay

MERGE_GRAPHQL = contrib/merge-graphql-schema.sh

CONNECT_GQL = $(wildcard pkg/server/api/connect/v1/graphql/*.graphql)
CONSOLE_GQL = $(wildcard pkg/server/api/console/v1/graphql/*.graphql)
COMPLIANCEPORTAL_GQL = $(wildcard pkg/server/api/complianceportal/v1/graphql/*.graphql)

pkg/server/api/connect/v1/schema.graphql: pkg/server/api/connect/v1/graphql $(CONNECT_GQL)
	$(MERGE_GRAPHQL) $@ pkg/server/api/connect/v1/graphql

pkg/server/api/console/v1/schema.graphql: pkg/server/api/console/v1/graphql $(CONSOLE_GQL)
	$(MERGE_GRAPHQL) $@ pkg/server/api/console/v1/graphql

pkg/server/api/complianceportal/v1/schema.graphql: pkg/server/api/complianceportal/v1/graphql $(COMPLIANCEPORTAL_GQL)
	$(MERGE_GRAPHQL) $@ pkg/server/api/complianceportal/v1/graphql

.PHONY: @probo/console
@probo/console: NODE_ENV=production
@probo/console: relay
	$(NPM) --workspace $@ run check
	$(NPM) --workspace $@ run build

.PHONY: @probo/compliance-portal
@probo/compliance-portal: NODE_ENV=production
@probo/compliance-portal: relay
	$(NPM) --workspace $@ run check
	$(NPM) --workspace $@ run build

.PHONY: generate
generate: $(GENERATED)

.PHONY: embed
embed: $(EMBEDDED)

pkg/server/api/connect/v1/schema/schema.go \
pkg/server/api/connect/v1/types/types.go: pkg/server/api/connect/v1/gqlgen.yaml pkg/server/api/connect/v1/graphql $(CONNECT_GQL)
	$(GO_GENERATE) ./pkg/server/api/connect/v1

# gqlgen instances must run sequentially: parallel runs race on the Go build
# cache and cause gqlgen's Rewriter.getSource() to panic with empty source.
pkg/server/api/console/v1/schema/schema.go \
pkg/server/api/console/v1/types/types.go: pkg/server/api/console/v1/gqlgen.yaml pkg/server/api/console/v1/graphql $(CONSOLE_GQL) | pkg/server/api/connect/v1/types/types.go apps/console/dist/index.html
	$(GO_GENERATE) ./pkg/server/api/console/v1

pkg/server/api/complianceportal/v1/schema/schema.go \
pkg/server/api/complianceportal/v1/types/types.go: pkg/server/api/complianceportal/v1/gqlgen.yaml pkg/server/api/complianceportal/v1/graphql $(COMPLIANCEPORTAL_GQL) | pkg/server/api/console/v1/types/types.go pkg/server/api/mcp/v1/types/types.go apps/compliance-portal/dist/index.html
	$(GO_GENERATE) ./pkg/server/api/complianceportal/v1

pkg/server/api/mcp/v1/server/server.go \
pkg/server/api/mcp/v1/types/types.go: pkg/server/api/mcp/v1/specification.yaml pkg/server/api/mcp/v1/mcpgen.yaml
	$(GO_GENERATE) ./pkg/server/api/mcp/v1

.PHONY: genmodels
genmodels: ## Refresh LLM model registry from OpenRouter
	$(GO_GENERATE) ./pkg/llm

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: fix
fix: fix-go ## Auto-fix Go code

.PHONY: fix-go
fix-go: generate embed ## Auto-fix Go code (format, go fix, lint --fix)
	gofmt -w apps cmd packages pkg e2e
	$(GO_BASE) fix -omitzero=false ./apps/... ./cmd/... ./packages/... ./pkg/... ./e2e/...
	$(GOLINTCMD) run --fix ./...

.PHONY: fmt
fmt: fmt-go ## Format Go code

.PHONY: fmt-go
fmt-go: ## Format Go code
	go fmt ./...

.PHONY: clean
clean: ## Clean the project (node_modules and build artifacts)
	$(RM) -rf bin/*
	$(RM) -rf node_modules
	$(RM) -rf apps/{console,compliance-portal}/{dist,node_modules}
	$(RM) -rf packages/emails/{dist,node_modules}
	$(RM) -rf sbom-docker.json sbom.json
	$(RM) -rf coverage.out coverage.html coverage-e2e.out coverage-e2e.html coverage-combined.out coverage-combined.html
	$(RM) -rf coverage/
	$(RM) -rf compose/keycloak/certs/cert.pem compose/keycloak/certs/private-key.pem compose/keycloak/probo-realm.json
	$(RM) -f pkg/server/api/connect/v1/schema/schema.go pkg/server/api/connect/v1/types/types.go
	$(RM) -f pkg/server/api/console/v1/schema/schema.go pkg/server/api/console/v1/types/types.go
	$(RM) -f pkg/server/api/complianceportal/v1/schema/schema.go pkg/server/api/complianceportal/v1/types/types.go
	$(RM) -f pkg/server/api/mcp/v1/server/server.go pkg/server/api/mcp/v1/types/types.go
	$(RM) -f $(RELAY_SCHEMAS)
	$(RM) -f pkg/llm/registry_gen.go
	find apps -type d -name __generated__ -exec $(RM) -rf {} +

.PHONY: stack-up
stack-up: compose/step-ca/certs/root_ca.crt compose/keycloak/probo-realm.json ## Start the docker stack as a deamon
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

compose/step-ca/certs/root_ca.crt:
	# step-ca runs as UID 1000 and creates 0700 dirs. On Linux CI the host
	# user cannot traverse them, so `-f` never sees root_ca.crt and a host
	# chmod cannot fix ownership. Open perms from inside the container.
	@$(MKDIR) compose/step-ca
	@chmod a+rwx compose/step-ca
	$(DOCKER_COMPOSE) up -d acme-http-01-proxy step-ca
	@i=0; \
	while [ ! -f $@ ] && [ $$i -lt 60 ]; do \
		$(DOCKER_COMPOSE) exec -T -u 0 step-ca chmod -R a+rX /home/step 2>/dev/null || true; \
		sleep 1; \
		i=$$((i + 1)); \
	done
	@test -f $@ || ($(DOCKER_COMPOSE) logs step-ca >&2; echo "step-ca root CA not ready" >&2; exit 1)
	@$(DOCKER_COMPOSE) exec -T -u 0 step-ca chmod -R a+rX /home/step

compose/keycloak/certs/cert.pem:
	$(MKDIR) ./compose/keycloak/certs
	$(OPENSSL) req -x509 -newkey rsa:2048 -keyout compose/keycloak/certs/private-key.pem -out compose/keycloak/certs/cert.pem -days 3650 -nodes -subj "/CN=keycloak-saml-signing"

compose/keycloak/probo-realm.json: compose/keycloak/probo-realm.json.tmpl compose/keycloak/certs/cert.pem
	$(SED) \
	-e "s|CERTIFICATE_PLACEHOLDER|$$(awk 'NR==1 {printf "%s", $$0; next} {printf "\\\\n%s", $$0}' compose/keycloak/certs/cert.pem)|g" \
	-e "s|PRIVATE_KEY_PLACEHOLDER|$$(awk 'NR==1 {printf "%s", $$0; next} {printf "\\\\n%s", $$0}' compose/keycloak/certs/private-key.pem)|g" \
	$@.tmpl > $@

apps/console/dist/index.html apps/compliance-portal/dist/index.html:
	$(MKDIR) $(dir $@)
	$(ECHO) dev-server > $@


.PHONY: sandbox-create
sandbox-create: ## Create a Lima sandbox VM for this worktree
	./contrib/lima/sandbox.sh create

.PHONY: sandbox-start
sandbox-start: ## Start the Lima sandbox VM
	./contrib/lima/sandbox.sh start

.PHONY: sandbox-boot-logs
sandbox-boot-logs: ## Show the Lima sandbox VM boot logs
	./contrib/lima/sandbox.sh boot-logs

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
