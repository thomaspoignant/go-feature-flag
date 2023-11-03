FROM golang:1.21

ARG VERSION=127.0.0.1

WORKDIR /go/src/app
COPY . /go/src/app

RUN go build -o /goff-test-configmap /go/src/app/examples/retriever_configmap/main.go
CMD ["/goff-test-configmap"]
