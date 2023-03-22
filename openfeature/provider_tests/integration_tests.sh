#!/bin/bash

# exit when any command fails
set -e

# build and launch relay proxy
make vendor
make build-relayproxy
./out/bin/relayproxy --config $(pwd)/openfeature/provider_tests/goff-proxy.yaml &

# Waiting for the relay proxy to be ready
NB_ITERATION=10
while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' localhost:1031/health)" != "200" ]]; do
  sleep 1
  NB_ITERATION=$((NB_ITERATION - 1))
  if [ ${NB_ITERATION} == "0" ]; then echo "ERROR: relay-proxy is not ready" && exit 123; fi
done

# Launch java integration tests
mvn -f $(pwd)/openfeature/provider_tests/java-integration-tests/pom.xml test

# Launch js integration tests
npm run test --prefix $(pwd)/openfeature/provider_tests/js-integration-tests/

# Kill all process launched by the script (here the relay-proxy)
kill -KILL %1