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
	GO111MODULE=off go get -u github.com/jstemmer/go-junit-report
ifeq ($(CI), true)
	$(GOTEST) -v -race ./... | tee /dev/tty | go-junit-report -set-exit-code > /tmp/test-results/junit-report.xml
else
	$(GOTEST) -v -race ./...
endif


coverage:
	# Create cover profile
	$(GOTEST) -cover -covermode=count -coverprofile=coverage.out ./...
	# Print code coverage details
	GO111MODULE=off go get github.com/mattn/goveralls
	GO111MODULE=off go get golang.org/x/tools/cmd/cover
	goveralls -service=circle-ci -coverprofile=coverage.out -v -package ./... -repotoken=${COVERALLS_TOKEN}

vendor:
	$(GOCMD) mod vendor
