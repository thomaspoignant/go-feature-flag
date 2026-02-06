#!/bin/bash

# exit when any command fails
set -e

# wait_relay_proxy is a function waiting for the relay proxy to be started
wait_relay_proxy () {
  PORT="$1"
  NB_ITERATION=10
  while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' "localhost:${PORT}/health")" != "200" ]]; do
    sleep 1
    NB_ITERATION=$((NB_ITERATION - 1))
    if [ ${NB_ITERATION} == "0" ]; then echo "ERROR: relay-proxy ${PORT} is not ready" && exit 123; fi
  done
}

# build and launch relay proxies
# we are launching 2 relay proxies because one is authenticated and the other is not.
make vendor
make build-relayproxy
./out/bin/relayproxy --config $(pwd)/openfeature/provider_tests/goff-proxy.yaml &
RELAY_PROXY_PID=$!
./out/bin/relayproxy --config $(pwd)/openfeature/provider_tests/goff-proxy-authenticated.yaml &
RELAY_PROXY_PID_AUTHENTICATED=$!

# Waiting for the relay proxies to be ready
wait_relay_proxy 1031
wait_relay_proxy 1032


# Launch java integration tests
echo "------------------------------------------------------------------------------------------------"
echo "----------- JAVA PROVIDER TESTS ----------------------------------------------------------------"
echo "------------------------------------------------------------------------------------------------"
mvn -f $(pwd)/openfeature/provider_tests/java-integration-tests/pom.xml test

# Launch js integration tests
echo "------------------------------------------------------------------------------------------------"
echo "--------- JAVASCRIPT PROVIDER TESTS ------------------------------------------------------------"
echo "------------------------------------------------------------------------------------------------"
npm install --prefix $(pwd)/openfeature/provider_tests/js-integration-tests/
npm run test --prefix $(pwd)/openfeature/provider_tests/js-integration-tests/

# Launch GO integration test
echo "------------------------------------------------------------------------------------------------"
echo "------------- GO PROVIDER TESTS ----------------------------------------------------------------"
echo "------------------------------------------------------------------------------------------------"
CURRENT_FOLDER=$(pwd)
cd openfeature/provider_tests/go-integration-tests
GOWORK=off go mod vendor
GOWORK=off go mod tidy
GOWORK=off go test . -tags=integration
cd "${CURRENT_FOLDER}"

# Launch .NET integration test
echo "------------------------------------------------------------------------------------------------"
echo "------------- .NET PROVIDER TESTS --------------------------------------------------------------"
echo "------------------------------------------------------------------------------------------------"
dotnet test openfeature/provider_tests/dotnet-integration-tests

# Stop the relay proxy PID
kill ${RELAY_PROXY_PID}
kill ${RELAY_PROXY_PID_AUTHENTICATED}
