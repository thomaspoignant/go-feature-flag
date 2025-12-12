GOCMD=go
TINYGOCMD=tinygo
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
ALL_GO_MOD_DIRS := ./ ./modules/evaluation ./modules/core ./cmd/wasm

# In CI we disable workspace mode.
ifeq ($(CI),true)
  GOWORK_ENV := GOWORK=off
  MODFLAG := -mod=vendor
else
  GOWORK_ENV :=
  MODFLAG :=
endif

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test build vendor

all: help
## Build:
build: build-modules build-relayproxy build-lint build-editor-api build-jsonschema-generator build-cli  ## Build all the binaries and put the output in out/bin/

create-out-dir:
	mkdir -p out/bin

build-relayproxy: create-out-dir ## Build the relay proxy in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOWORK_ENV) $(GOCMD) build $(MODFLAG) -o out/bin/relayproxy ./cmd/relayproxy/

build-lint: create-out-dir ## Build the linter in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOWORK_ENV) $(GOCMD) build $(MODFLAG) -o out/bin/lint ./cmd/lint/

build-cli: create-out-dir ## Build the linter in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOWORK_ENV) $(GOCMD) build $(MODFLAG) -o out/bin/cli ./cmd/cli/

build-editor-api: create-out-dir ## Build the linter in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOWORK_ENV) $(GOCMD) build $(MODFLAG) -o out/bin/editor-api ./cmd/editor/

build-jsonschema-generator: create-out-dir ## Build the jsonschema-generator in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOWORK_ENV) $(GOCMD) build $(MODFLAG) -o out/bin/jsonschema-generator ./cmd/jsonschema-generator/

build-wasm: create-out-dir ## Build the wasm evaluation library in out/bin/
	cd cmd/wasm && $(TINYGOCMD) build -o ../../out/bin/gofeatureflag-evaluation.wasm -target wasm -opt=2 -opt=s --no-debug -scheduler=none

build-wasi: create-out-dir ## Build the wasi evaluation library in out/bin/
	cd cmd/wasm && $(TINYGOCMD) build -o ../../out/bin/gofeatureflag-evaluation.wasi -target wasi -opt=2 -opt=s --no-debug -scheduler=none

build-modules:  ## Run build command to build all modules in the workspace
	@echo "Building all modules in the workspace..."
	@$(foreach module, $(ALL_GO_MOD_DIRS), (echo "→ Building $(module)"; CGO_ENABLED=0 GO111MODULE=on $(GOWORK_ENV) $(GOCMD) build $(MODFLAG) ./...);)


build-doc: ## Build the documentation
	cd website; \
	npm i && npm run build

clean: ## Remove build related file
	-rm -fr ./bin ./out ./release
	-rm -f ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.cov yamllint-checkstyle.xml ./coverage-*.cov.tmp
	-rm -rf vendor
	-rm -f go.work.sum go.work

vendor: tidy ## Copy of all packages needed to support builds and tests in the vendor directory
ifneq ($(CI),)
	$(GOWORK_ENV) $(GOCMD) mod vendor
else
	$(GOWORK_ENV) $(GOCMD) work vendor
endif

tidy: ## Run go mod tidy for all modules in the workspace	
ifeq ($(CI),)
	$(GOWORK_ENV) $(GOCMD) work sync
endif
	@$(foreach module, $(ALL_GO_MOD_DIRS), (echo "→ Tidying $(module)"; cd $(module) && $(GOWORK_ENV) $(GOCMD) mod tidy);)

## Dev:
workspace-init:
	go work init
	$(foreach module, $(ALL_GO_MOD_DIRS), go work use $(module);)
	go work sync

watch-relayproxy: ## Launch the relay proxy in watch mode.
	docker run -it --rm -w /go/src/github.com/thomaspoignant/go-feature-flag/ \
		-v $(shell pwd):/go/src/github.com/thomaspoignant/go-feature-flag \
		-v $(shell pwd)/cmd/relayproxy/testdata/config/valid-file.yaml:/goff/goff-proxy.yaml \
		-p 1031:1031 cosmtrek/air \
		--build.cmd "$(GOWORK_ENV) go build $(MODFLAG) -o out/bin/relayproxy ./cmd/relayproxy/" \
		--build.bin "./out/bin/relayproxy"

watch-doc: ## Launch a local server to work on the documentation
	cd website; \
    	npm i && npx docusaurus start

serve-doc: ## Serve the doc build by the build-doc target
	cd website; \
		npm run serve

swagger: ## Build swagger documentation
	$(GOWORK_ENV) $(GOCMD) install github.com/swaggo/swag/cmd/swag@latest
	cd cmd/relayproxy && swag init --parseDependency --parseDepth=1 --parseInternal --markdownFiles docs

generate-helm-docs: ## Generates helm documentation for the project
	$(GOWORK_ENV) $(GOCMD) install github.com/norwoodj/helm-docs/cmd/helm-docs@latest
	helm-docs

bump-helm-chart-version: ## Bump Helm chart version (usage: make bump-helm-chart-version VERSION=v1.2.3)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required$(RESET)"; \
		echo "Usage: make bump-helm-chart-version VERSION=v1.2.3"; \
		echo "       make bump-helm-chart-version VERSION=1.2.3"; \
		exit 1; \
	fi
	.github/ci-scripts/bump-helm-chart.sh $(VERSION)

## Test:
test: ## Run the tests of the project
	@for module in $(ALL_GO_MOD_DIRS); do \
		echo "→ Testing $$module"; \
		cd $$module && $(GOWORK_ENV) $(GOCMD) test -v -race -tags=docker ./... || exit 1; \
		cd - >/dev/null; \
	done

provider-tests: ## Run the integration tests for the Open Feature Providers
	./openfeature/provider_tests/integration_tests.sh

coverage: ## Run the tests of the project and export the coverage for all modules
	@rm -f coverage*.cov coverage*.cov.tmp
	@for module in $(ALL_GO_MOD_DIRS); do \
		echo "Running coverage for $$module..."; \
		export original_path=$(shell pwd); \
		covfile="coverage-$$(basename $$module).cov.tmp"; \
		cd $$module && $(GOWORK_ENV) $(GOTEST) -cover -covermode=count -tags=docker -coverprofile=$${original_path}/$${covfile} ./...; \
		cd $$original_path; \
	done
	@echo "mode: count" > coverage.cov && cat *.cov.tmp | grep -v "mode: count" | grep -v "/examples/" >> coverage.cov
	@rm -f *.cov.tmp

bench: ## Launch the benchmark test
	 $(GOWORK_ENV) $(GOTEST) -tags=bench -bench Benchmark -cpu 2 -run=^$$

## Lint:
lint: ## Use golintci-lint on your project
	mkdir -p ./bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest # Install linters
	./bin/golangci-lint run --timeout=5m ./... # Run linters

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
