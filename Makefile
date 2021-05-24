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
	# Create cover profile
	$(GOTEST) -cover -covermode=count -coverprofile=coverage.out ./...
ifeq ($(CI), true)
	# Print code coverage details
	GO111MODULE=off go get github.com/mattn/goveralls
	GO111MODULE=off go get golang.org/x/tools/cmd/cover
	goveralls -service=circle-ci -coverprofile=coverage.out -v -package ./... -repotoken=${COVERALLS_TOKEN}
endif

vendor:
	$(GOCMD) mod vendor
