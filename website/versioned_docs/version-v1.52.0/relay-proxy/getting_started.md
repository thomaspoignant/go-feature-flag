---
sidebar_position: 10
description: Getting started with the relay proxy.
---

# 🏁 Getting started

## Overview
The **relay proxy** is where the magic happen, it allows to retrieve flags from a remote source and serve them to your applications.

This is the component that will be in charge of serving the flags to the different provider and SDKs of your application.

In this page, we will guide you to set up quickly the relay-proxy

## Installation

We will use the docker image for this first installation:

```shell
docker pull gofeatureflag/go-feature-flag:latest
```

## Configuration

Before starting your **relay proxy** you will need to create a minimal configuration file.  

```yaml title="goff-proxy.yaml"
# this is a minimal config containing only where your flag file is located 
retrievers:
  - kind: github
    repositorySlug: thomaspoignant/go-feature-flag
    path: cmd/relayproxy/testdata/dockerhub-example/flags.goff.yaml
```

## Starting
After that you can launch the **relay proxy** by using this command:
```shell
docker run -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml -p 1031:1031 gofeatureflag/go-feature-flag
```

What is happening?
- The **relay proxy** will read the configuration file and use the github retriever to load the flags from the configuration file.
- The relay-proxy is also starting to poll the file regurarly to check if any changes are happening.
- A set of APIs are available to evaluate the flags

## Evaluate a flag
You can validate that the relay-proxy is working by trying a flag evaluation:

```shell
curl 'http://localhost:1031/ofrep/v1/evaluate/flags/flag-only-for-admin' -X POST \
  -H 'Content-Type: application/json' \
  --data-raw '{"context": {"company": "GO Feature Flag", "firstname": "John", "lastname": "Doe", "targetingKey": "4f433951-4c8c-42b3-9f18-8c9a5ed8e9eb"}}'
```

You should receive a response like this one:
```json
{
    "key": "flag-only-for-admin",
    "value": false,
    "reason": "DEFAULT",
    "variant": "default_var",
    "metadata": {
        "gofeatureflag_cacheable": true
    }
}
```

🎉 Congrats you have a working relay-proxy locally 🎉.

Follow the next pages if you want to deploy it in your infrastructure and to have all the options to configure it.

