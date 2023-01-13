FROM golang:1.18 AS build

ARG VERSION=0.0.0-SNAPSHOT

WORKDIR /go/src/app
COPY . /go/src/app

RUN make vendor && make build VERSION="${VERSION}"

FROM gcr.io/distroless/base-debian11:latest
COPY --from=build /go/src/app/out/bin/go-feature-flag-relay-proxy /
CMD ["/go-feature-flag-relay-proxy"]
