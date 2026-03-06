FROM gcr.io/distroless/static-debian13:latest
ARG TARGETPLATFORM
ENTRYPOINT ["/go-feature-flag-cli"]
COPY $TARGETPLATFORM/go-feature-flag-cli /
