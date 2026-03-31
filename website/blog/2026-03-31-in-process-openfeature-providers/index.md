---
title: '⚡ In-Process Evaluation for GO Feature Flag OpenFeature Providers'
description: 'GO, Java, .NET, Python, and JavaScript/TypeScript providers now support in-process evaluation for faster flag resolution with far fewer network calls.'
authors: [thomaspoignant]
tags: [openfeature, performance, in-process, ofrep]
---

# ⚡ In-Process Evaluation for GO Feature Flag OpenFeature Providers

We are happy to introduce **in-process evaluation** for the GO Feature Flag OpenFeature server providers:

- Go
- Java
- .NET
- Python
- JavaScript/TypeScript

This new capability is designed to improve evaluation performance and drastically reduce network traffic between your applications and the relay-proxy.

<!--truncate-->

## What changes with in-process evaluation?

With **remote evaluation**, each flag resolution triggers a network call to the relay-proxy.

With **in-process evaluation**, the provider:
1. Periodically fetches the latest flag configuration from the relay-proxy.
2. Stores it locally in memory.
3. Evaluates flags directly inside your application process.

That means your flag checks no longer depend on a network round trip on the critical path.

## Performance gains in practice

Moving evaluation in-process gives you immediate performance benefits:

- **Lower latency**: evaluations are local, so there is no request/response overhead per flag check.
- **Higher throughput**: your app can evaluate more flags with less pressure on the relay-proxy.
- **Better resilience**: short network hiccups have less impact on runtime flag evaluation.

## Fewer network calls

This is one of the biggest wins.

| Mode | Network behavior |
| --- | --- |
| Remote evaluation | 1 network call for each flag evaluation |
| In-process evaluation | No network call for each flag evaluation (only periodic config refresh + event export) |

So if your request path evaluates several flags, in-process mode can remove a lot of network calls and reduce infrastructure overhead.

## Which providers support it?

The in-process capability is available in the GO Feature Flag OpenFeature providers for:

- [Go](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_go)
- [Java](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_java)
- [.NET](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_dotnet)
- [Python](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_python)
- [JavaScript/TypeScript](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_javascript)

## In-process or remote: how to choose?

Use **in-process** when you want:
- the best runtime performance
- very low latency
- fewer network calls in hot code paths

You can always use **remote evaluation with the OpenFeature Remote Evaluation Protocol (OFREP)** if you prefer to rely on the latest information from the relay-proxy at evaluation time.

Learn more about OFREP here:  
[OpenFeature Remote Evaluation Protocol (OFREP)](https://gofeatureflag.org/docs/sdk/ofrep)

## Final thoughts

In-process evaluation gives you a strong performance boost while keeping the OpenFeature developer experience you already use today.

If your workload is latency-sensitive, this mode is a great default. And when your architecture needs centralized, always-fresh remote decisions, OFREP remains available.
