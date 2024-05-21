# OpenFeature Remote Evaluation Protocol (OFREP)

![Experimental](https://img.shields.io/badge/Status-Experimental-red.svg)  
⚠️ Note that this a work in progress and the protocol is subject to change. ⚠️

OpenFeature Remote Flag Evaluation Protocol is an API specification for feature flagging that allows the use of generic
providers to connect to any feature flag management systems that supports the protocol.

Currently, the protocol is in the early stages of development and is not yet ready for production use, but GO Feature Flag
is supporting the protocol and is the first implementation of the protocol.
We are part of the leading team in the protocol, and we try to follow the specification during the early stages of the protocol
to allow people to try it and be able to develop the providers.

## How to test it?

The OFREP implementation is part of the GO Feature Flag Relay Proxy.
We have a new API endpoints `/ofrep/v1/evaluate/flags/{key}` and `/ofrep/v1/evaluate/flags` that you can use to test the protocol.

You just have to start the GO Feature Flag Relay Proxy (starting from version `v1.24.0`) and use the API to evaluate your flags.
For this, follow the instruction on how to use the relay-proxy [here](../relay_proxy/getting_started.md). 

### Want to start even faster?
```shell
curl https://gist.githubusercontent.com/thomaspoignant/181a067291a04bd1fbb55468629625d2/raw/eacfc2ae1036c1cfef669b41ec7b54c119639c0c/goff-proxy.yaml -o goff-proxy.yaml
docker run -p 1031:1031 -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml thomaspoignant/go-feature-flag:latest
```

This will launch a GO Feature Flag Relay Proxy with a configuration file that will retrieve the flags from the test server.

Swagger is enabled, so you can directly go to http://localhost:1031/swagger/index.html to test the OFREP endpoints (the API Key to use is `apikey1`).
