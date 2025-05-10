#!/usr/bin/env bash

VERSION=$1

make build-wasm
make build-wasi

mkdir -p "./out/release-wasm/"
mv "./out/bin/gofeatureflag-evaluation.wasi" "./out/release-wasm/gofeatureflag-evaluation_${VERSION}.wasi"
mv "./out/bin/gofeatureflag-evaluation.wasm" "./out/release-wasm/gofeatureflag-evaluation_${VERSION}.wasm"
 