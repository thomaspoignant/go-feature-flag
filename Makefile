GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet

.PHONY: all test build vendor

lint:
	mkdir -p ./bin
	# Install linters
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest
	# Run linters
	./bin/golangci-lint run --deadline=65s ./...

test:
	$(GOTEST) -v -race ./...

coverage:
	$(GOTEST) -cover -covermode=count -coverprofile=coverage.cov ./...

vendor:
	$(GOCMD) mod vendor

bench:
	 $(GOTEST) -bench Benchmark -cpu 2 -run=^$$
