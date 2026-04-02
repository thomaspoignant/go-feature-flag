---
title: 'In-Process Evaluation for GO Feature Flag OpenFeature Providers'
description: 'GO, Java, .NET, Python, and JavaScript/TypeScript providers now support in-process evaluation for faster flag resolution with far fewer network calls.'
authors: [thomaspoignant]
tags: [openfeature, performance, in-process, ofrep]
image: https://gofeatureflag.org/assets/images/inprocess-banner-61bbdb9406831475b10959f8702e9872.png
---

![In-Process OpenFeature Providers Banner](./inprocess-banner.png)

# ⚡ In-Process Evaluation for GO Feature Flag OpenFeature Providers

We are excited to introduce **in-process evaluation** for the GO Feature Flag OpenFeature server providers: `Go`, `Java`, `.NET`, `Python`, `JavaScript`, and `TypeScript`.

This capability is designed to improve evaluation performance and drastically reduce network traffic between your applications and the relay-proxy.

## What changes with in-process evaluation?

Previously, when you used the OpenFeature providers we ship, evaluation used **remote evaluation**: each flag evaluation triggered a network call to the relay-proxy.

That model fits many setups, but it is not always the most performant or resilient way to evaluate feature flags on hot paths.

That is why we are launching **in-process** evaluation mode for our most widely used providers.

<!--truncate-->

### How it works

With **in-process evaluation**, the provider:

1. Periodically fetches the latest flag configuration from the relay-proxy.
2. Stores it locally in memory.
3. Evaluates flags directly inside your application process, using a shared evaluation engine so every provider behaves consistently.

Your flag checks no longer depend on a network round trip on the critical path.

## But what about the evaluation data?

Not calling the relay-proxy to evaluate your flag does not mean you lose observability.

The providers collect evaluation data from local evaluations and send it to the relay-proxy, so you can keep observing evaluations as before. For how events are recorded when evaluation runs in the provider, see [flag usage tracking](/docs/tracking/flag-usage-tracking).


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

If your request path evaluates several flags, in-process mode can remove a lot of network calls and reduce infrastructure overhead.

## Which providers support it?

The in-process capability is available in the GO Feature Flag OpenFeature providers for:

- [Go](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_go)
- [Java](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_java)
- [.NET](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_dotnet)
- [Python](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_python)
- [JavaScript/TypeScript](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_javascript)

## How to enable it

Each provider lets you choose an evaluation mode **in-process** (evaluate locally after syncing config) or **remote** (ask the relay-proxy on each evaluation)—through its configuration. The option names differ by language (for example, JavaScript and TypeScript use `evaluationType` with `EvaluationType.InProcess` or `EvaluationType.Remote`; Java and .NET expose `evaluationType` in their provider options with the same idea).

Use the documentation for your runtime above for exact types, defaults, and examples.

## In-process or remote: how to choose?

Use **in-process** when you want:

- the best runtime performance
- very low latency
- fewer network calls in hot code paths

Use **remote** evaluation when you want each evaluation to hit the relay-proxy so decisions are always tied to the latest server-side evaluation. That can be a good fit when network latency to the relay is negligible, for example, in a **sidecar** deployment where the proxy sits beside your app.

## Final thoughts

In-process evaluation gives you a strong reliability and performance boost while keeping the OpenFeature developer experience you already use today.

If your workload is latency-sensitive, this mode is a great default.
