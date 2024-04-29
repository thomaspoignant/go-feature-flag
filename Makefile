GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test build vendor



all: help
## Build:
build: build-migrationcli build-relayproxy build-lint build-editor-api build-jsonschema-generator ## Build all the binaries and put the output in out/bin/

create-out-dir:
	mkdir -p out/bin

build-migrationcli: create-out-dir ## Build the migration cli in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/migrationcli ./cmd/migrationcli/

build-relayproxy: create-out-dir ## Build the relay proxy in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/relayproxy ./cmd/relayproxy/

build-lint: create-out-dir ## Build the linter in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/lint ./cmd/lint/

build-editor-api: create-out-dir ## Build the linter in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/editor-api ./cmd/editor/

build-jsonschema-generator: create-out-dir ## Build the jsonschema-generator in out/bin/
	CGO_ENABLED=0 GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/jsonschema-generator ./cmd/jsonschema-generator/

build-doc: ## Build the documentation
	cd website; \
	npm i && npm run build

clean: ## Remove build related file
	-rm -fr ./bin ./out ./release
	-rm -f ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.cov yamllint-checkstyle.xml

vendor: ## Copy of all packages needed to support builds and tests in the vendor directory
	$(GOCMD) mod tidy
	$(GOCMD) mod vendor

## Dev:
watch-relayproxy: ## Launch the relay proxy in watch mode.
	docker run -it --rm -w /go/src/github.com/thomaspoignant/go-feature-flag/ \
		-v $(shell pwd):/go/src/github.com/thomaspoignant/go-feature-flag \
		-v $(shell pwd)/cmd/relayproxy/testdata/config/valid-file.yaml:/goff/goff-proxy.yaml \
		-p 1031:1031 cosmtrek/air \
		--build.cmd "go build -mod vendor -o out/bin/relayproxy ./cmd/relayproxy/" \
		--build.bin "./out/bin/relayproxy"

watch-doc: ## Launch a local server to work on the documentation
	cd website; \
    	npm i && npx docusaurus start

serve-doc: ## Serve the doc build by the build-doc target
	cd website; \
		npm run serve

swagger: ## Build swagger documentation
	$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest
	cd cmd/relayproxy && swag init --parseDependency --parseDepth=1 --parseInternal --markdownFiles docs

generate-helm-docs: ## Generates helm documentation for the project
	$(GOCMD) install github.com/norwoodj/helm-docs/cmd/helm-docs@latest
	helm-docs

## Test:
test: ## Run the tests of the project
	$(GOTEST) -v -race ./... -tags=docker

provider-tests: ## Run the integration tests for the Open Feature Providers
	./openfeature/provider_tests/integration_tests.sh

coverage: ## Run the tests of the project and export the coverage
	$(GOTEST) -cover -covermode=count -tags=docker -coverprofile=coverage.cov.tmp ./... \
	&& cat coverage.cov.tmp | grep -v "/examples/" > coverage.cov

bench: ## Launch the benchmark test
	 $(GOTEST) -bench Benchmark -cpu 2 -run=^$$

## Lint:
lint: ## Use golintci-lint on your project
	mkdir -p ./bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest # Install linters
	./bin/golangci-lint run --timeout=5m --timeout=5m ./... # Run linters

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
