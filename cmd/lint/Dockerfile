FROM gcr.io/distroless/static-debian13:latest
ARG TARGETPLATFORM
ENTRYPOINT ["/go-feature-flag-lint"]
COPY $TARGETPLATFORM/go-feature-flag-lint /
