FROM golang:1.18 AS build

ARG VERSION=127.0.0.1

WORKDIR /go/src/app
COPY . /go/src/app

RUN go build -o /go/src/app/examples/retriever_configmap/goff-test-configmap /go/src/app/examples/retriever_configmap/main.go

FROM gcr.io/distroless/base-debian11:latest
COPY --from=build /go/src/app/examples/retriever_configmap/goff-test-configmap /
CMD ["/goff-test-configmap"]
