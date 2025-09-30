FROM golang:1.21

ARG VERSION=127.0.0.1

WORKDIR /go/src/app
COPY . /go/src/app

RUN go build -o /goff-test-k8s-secret /go/src/app/examples/retriever_k8s_secret/main.go
CMD ["/goff-test-k8s-secret"]
