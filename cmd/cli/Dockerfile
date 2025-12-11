FROM gcr.io/distroless/static-debian12:latest
ARG TARGETPLATFORM
ENTRYPOINT ["/go-feature-flag-cli"]
COPY $TARGETPLATFORM/go-feature-flag-cli /
