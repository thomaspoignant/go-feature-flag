#!/usr/bin/env bash
# Without `set -e` a failing build or verification would not stop the script:
# the mv commands below would still run and the workflow would publish the
# assets with a zero exit code.
set -euo pipefail

VERSION=$1

make build-wasm
make build-wasi
# Fail the release rather than publish a binary whose shadow stack silently
# regressed to the 64KB wasm-ld default (see WASM_STACK_SIZE in the Makefile).
make verify-wasm-stack

mkdir -p "./out/release-wasm/"
mv "./out/bin/gofeatureflag-evaluation.wasi" "./out/release-wasm/gofeatureflag-evaluation_${VERSION}.wasi"
mv "./out/bin/gofeatureflag-evaluation.wasm" "./out/release-wasm/gofeatureflag-evaluation_${VERSION}.wasm"
 