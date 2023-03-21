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
build: build-migrationcli build-relayproxy build-lint ## Build all the binaries and put the output in out/bin/

build-migrationcli: ## Build the migration cli in out/bin/
	mkdir -p out/bin
	GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/migrationcli ./cmd/migrationcli/

build-relayproxy: ## Build the relay proxy in out/bin/
	mkdir -p out/bin
	GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/relayproxy ./cmd/relayproxy/

build-lint: ## Build the linter in out/bin/
	mkdir -p out/bin
	GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/lint ./cmd/lint/

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
	cd cmd/relayproxy && swag init --markdownFiles docs

## Test:
test: ## Run the tests of the project
	$(GOTEST) -v -race ./...

provider-tests:
	./openfeature/provider_tests/integration_tests.sh

coverage: ## Run the tests of the project and export the coverage
	$(GOTEST) -cover -covermode=count -coverprofile=coverage.cov.tmp ./... \
	&& cat coverage.cov.tmp | grep -v "/examples/" > coverage.cov

bench: ## Launch the benchmark test
	 $(GOTEST) -bench Benchmark -cpu 2 -run=^$$

## Lint:
lint: ## Use golintci-lint on your project
	mkdir -p ./bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest # Install linters
	./bin/golangci-lint run --deadline=5m --timeout=5m ./... # Run linters

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
