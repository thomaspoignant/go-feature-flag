# Scanner image for https://github.com/anthropics/oss-scanner (projects/go-feature-flag).
#
# The audit runs offline, so this build fetches every module, builds the library, the relay proxy, the CLI, the editor
# API and the jsonschema generator, and precompiles the tests. threat_model.md, next to this file, says how to run them.
FROM golang:1.26-bookworm
ENV DEBIAN_FRONTEND=noninteractive
# curl to talk to a running relay proxy.
RUN apt-get update && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home --uid 10001 scanner \
    && git config --system --add safe.directory /src

# scanner contract: the checkout lives inside the image, at /src. The scanner needs the whole checkout (the build
# context is the public git repository, nothing else), so the recursive copy is intended.
COPY --chown=scanner:scanner . /src
WORKDIR /src
USER scanner

# GOWORK=off mirrors CI (see Makefile): the root module resolves modules/core from the module proxy.
# GOCACHE lives under the world-writable /go so the build cache is reused whichever user runs go later.
ENV GOWORK=off \
    GOCACHE=/go/cache

RUN go mod download && (cd modules/core && go mod download)

# A step that fails fails the build, and primary_contact gets the log.
RUN go build ./... \
    && (cd modules/core && go build ./... && go test -run '^$' ./...) \
    && go test -run '^$' ./... \
    && mkdir -p out/bin \
    && go build -o out/bin/relayproxy ./cmd/relayproxy \
    && go build -o out/bin/cli ./cmd/cli \
    && go build -o out/bin/editor-api ./cmd/editor \
    && go build -o out/bin/jsonschema-generator ./cmd/jsonschema-generator

# Run the short tests so the log shows the image works, without letting a failure abort the build. Tests tagged
# `docker` (testcontainers) are left out: there is no Docker daemon in the scanner.
RUN (timeout 1800 go test -short ./... || echo "WARNING: go test reported failures") \
    && (cd modules/core && (timeout 600 go test -short ./... || echo "WARNING: modules/core tests reported failures"))

# Everything is fetched. From here on the go command must not try to reach a proxy or download a toolchain.
ENV GOPROXY=off \
    GOTOOLCHAIN=local
