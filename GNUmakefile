NPM?=	npm
NPX?=	npx
PRETTIER?=	$(NPX) prettier
GO?=	go
DOCKER?=	docker
SYFT ?= syft
GRYPE ?= grype

DOCKER_BUILD_FLAGS?=
DOCKER_BUILD=	DOCKER_BUILDKIT=1 $(DOCKER) build $(DOCKER_BUILD_FLAGS)

DOCKER_COMPOSE=	$(DOCKER) compose -f compose.yaml $(DOCKER_COMPOSE_FLAGS)

VERSION=	0.50.1
LDFLAGS=	-ldflags "-X 'main.version=$(VERSION)' -X 'main.env=prod'"
GCFLAGS=	-gcflags="-e"

CGO_ENABLED?=	0
GOOS?=

GO_BASE=	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go
GO_BUILD=	$(GO_BASE) build $(LDFLAGS) $(GCFLAGS)
GO_GENERATE=	$(GO_BASE) generate
GO_TEST=	$(GO_BASE) test $(TEST_FLAGS)
GO_VET=	$(GO_BASE) vet

TEST_FLAGS?=	-race -cover -coverprofile=coverage.out

DOCKER_IMAGE_NAME=	ghcr.io/getprobo/probo
DOCKER_TAG_NAME?=	latest

PROBOD_BIN=	bin/probod
PROBOD_SRC=	cmd/probod/main.go

.PHONY: all
all: build

.PHONY: lint
lint: vet # npm-lint

.PHONY: vet
vet:
	$(GO_VET) ./...

.PHONY: npm-lint
npm-lint:
	$(NPM) run lint

.PHONY: test
test: CGO_ENABLED=1
test: ## Run tests with race detection and coverage (usage: make test [MODULE=./pkg/some/module])
	$(GO_TEST) $(if $(MODULE),$(MODULE),./...)

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

.PHONY: build
build: @probo/console bin/probod

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
	$(GRYPE) sbom:sbom.json --fail-on high

.PHONY: scan-sbom-docker
scan-sbom-docker: sbom-docker
	$(GRYPE) sbom:sbom-docker.json --fail-on high

.PHONY: scan-docker
scan-docker: docker-build
	$(GRYPE) docker:$(DOCKER_IMAGE_NAME):$(DOCKER_TAG_NAME) --fail-on high

.PHONY: scan
scan: scan-sbom scan-sbom-docker scan-docker

.PHONY: docker-build
docker-build:
	$(DOCKER_BUILD) --tag $(DOCKER_IMAGE_NAME):$(DOCKER_TAG_NAME) --file Dockerfile .

.PHONY: bin/probod
bin/probod: pkg/server/api/console/v1/schema/schema.go pkg/server/api/console/v1/types/types.go pkg/server/api/console/v1/v1_resolver.go vet
	$(GO_BUILD) -o $(PROBOD_BIN) $(PROBOD_SRC)

.PHONY: @probo/console
@probo/console: NODE_ENV=production
@probo/console:
	$(NPM) --workspace $@ run check
	$(NPM) --workspace $@ run build

pkg/server/api/console/v1/schema/schema.go \
pkg/server/api/console/v1/types/types.go \
pkg/server/api/console/v1/v1_resolver.go: pkg/server/api/console/v1/gqlgen.yaml pkg/server/api/console/v1/schema.graphql
	$(GO_GENERATE) ./pkg/server/api/console/v1

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
	$(RM) -rf apps/console/{dist,node_modules}
	$(RM) -rf sbom-docker.json sbom.json
	$(RM) -rf coverage.out coverage.html

.PHONY: stack-up
stack-up: ## Start the docker stack as a deamon
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
