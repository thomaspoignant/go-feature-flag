# 🧩 SDKs (with OpenFeature)

## Overview[​](#overview "Direct link to Overview")

**GO Feature Flag** stands out as a feature flag solution, distinct from others, due to our decision to fully support the [`OpenFeature`](https://openfeature.dev) CNCF project.

The benefit of choosing OpenFeature lies in its framework-agnostic nature. When utilizing OpenFeature SDKs, you minimize the effort required for switching to a different feature flag provider. This flexibility empowers you to make changes without being tied down to a particular vendor.

To show our commitment to this initiative, **GO Feature Flag has opted not to develop any custom SDKs and instead relies entirely on OpenFeature SDKs**. In order to seamlessly integrate with our solution, we offer [`providers`](https://docs.openfeature.dev/docs/reference/concepts/provider) for GO Feature Flag in multiple programming languages.

Rest assured, working with OpenFeature SDKs alongside GO Feature Flag providers is as straightforward as using any other feature flag solution.<br /><!-- -->The added advantage is that you now adhere to a standard approach and avoid any vendor lock-in 🎉!

info

To be able to use GO Feature Flag with OpenFeature SDKs, you need to have the [`relay-proxy`](https://gofeatureflag.org/docs/relay-proxy.md) running in your infrastructure.

## Difference between Server-side SDKs & Client-side SDKs[​](#difference-between-server-side-sdks--client-side-sdks "Direct link to Difference between Server-side SDKs & Client-side SDKs")

The primary difference between these two paradigms has to do with the way they model evaluation context.<br /><!-- -->In server usage, the evaluation context changes frequently, as often as every evaluation, while in client-side usage, the evaluation context changes less frequently, often in response to user actions or UI events.

This is the reason why we have different SDKs for server and client-side usage.

[*More in depth about the differences between server and client-side SDKs*](https://gofeatureflag.org/docs/concepts/sdk-paradigms.md)

## Supported Languages[​](#supported-languages "Direct link to Supported Languages")

GO Feature Flag support both paradigms, and the SDKs are designed to make it easy to use the same feature flags in both server and client applications.

### Client-side Languages[​](#client-side-languages "Direct link to Client-side Languages")

**

##### [Javascript](https://gofeatureflag.org/sdk/client_providers/openfeature_javascript)

![badge](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue\&style=flat-square\&logo=npm)

* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/client_providers/openfeature_javascript)

**

##### [Typescript](https://gofeatureflag.org/sdk/client_providers/openfeature_javascript)

![badge](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue\&style=flat-square\&logo=npm)

* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/client_providers/openfeature_javascript)

**

##### [React](https://gofeatureflag.org/sdk/client_providers/openfeature_react)

![badge](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue\&style=flat-square\&logo=npm)

* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/client_providers/openfeature_react)

**

##### [Angular](https://gofeatureflag.org/sdk/client_providers/openfeature_angular)

![badge](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue\&style=flat-square\&logo=npm)

* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/client_providers/openfeature_angular)

**

##### [Swift](https://gofeatureflag.org/sdk/client_providers/openfeature_swift)

![badge](https://img.shields.io/github/v/release/go-feature-flag/openfeature-swift-provider?label=Swift\&amp;display_name=tag\&style=flat-square\&logo=Swift)

* **Remote Evaluation
* **Tracking Flag Evaluation
* **Configuration Change Updates
* **Provider Events Reactions
* **Tracking Custom Events

[More details**](https://gofeatureflag.org/sdk/client_providers/openfeature_swift)

**

##### [Kotlin / Android](https://gofeatureflag.org/sdk/client_providers/openfeature_android)

![badge](https://img.shields.io/maven-central/v/org.gofeatureflag.openfeature/gofeatureflag-kotlin-provider?color=blue\&style=flat-square\&logo=android)

* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/client_providers/openfeature_android)

### Server-side Languages[​](#server-side-languages "Direct link to Server-side Languages")

**

##### [Golang](https://gofeatureflag.org/sdk/server_providers/openfeature_go)

![badge](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fproxy.golang.org%2Fgithub.com%2Fopen-feature%2Fgo-sdk-contrib%2Fproviders%2Fgo-feature-flag%2F%40latest\&query=%24.Version\&label=GO\&color=blue\&style=flat-square\&log=go)

* **In process Evaluation
* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_go)

**

##### [Java](https://gofeatureflag.org/sdk/server_providers/openfeature_java)

![badge](https://img.shields.io/maven-central/v/dev.openfeature.contrib.providers/go-feature-flag?color=blue\&style=flat-square\&logo=java)

* **In process Evaluation
* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_java)

**

##### [Kotlin](https://gofeatureflag.org/sdk/server_providers/openfeature_java)

![badge](https://img.shields.io/maven-central/v/dev.openfeature.contrib.providers/go-feature-flag?color=blue\&style=flat-square\&logo=java)

* **In process Evaluation
* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_java)

**

##### [.NET](https://gofeatureflag.org/sdk/server_providers/openfeature_dotnet)

![badge](https://img.shields.io/nuget/v/OpenFeature.Providers.GOFeatureFlag?color=blue\&style=flat-square\&logo=nuget)

* **In process Evaluation
* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_dotnet)

**

##### [Python](https://gofeatureflag.org/sdk/server_providers/openfeature_python)

![badge](https://img.shields.io/pypi/v/gofeatureflag-python-provider?color=blue\&style=flat-square\&logo=pypi)

* **In process Evaluation
* **Remote Evaluation
* **Tracking Flag Evaluation
* **Configuration Change Updates
* **Provider Events Reactions
* **Tracking Custom Events

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_python)

**

##### [Node.JS](https://gofeatureflag.org/sdk/server_providers/openfeature_javascript)

![badge](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-provider?color=blue\&style=flat-square\&logo=npm)

* **In process Evaluation
* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_javascript)

**

##### [PHP](https://gofeatureflag.org/sdk/server_providers/openfeature_php)

![badge](https://img.shields.io/packagist/v/open-feature/go-feature-flag-provider?logo=php\&color=blue\&style=flat-square)

* **Remote Evaluation
* **Tracking Flag Evaluation
* **In process Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_php)

**

##### [Ruby](https://gofeatureflag.org/sdk/server_providers/openfeature_ruby)

![badge](https://img.shields.io/gem/v/openfeature-go-feature-flag-provider?color=blue\&style=flat-square\&logo=ruby)

* **Remote Evaluation
* **Tracking Flag Evaluation
* **In process Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_ruby)

**

##### [Rust](https://gofeatureflag.org/sdk/server_providers/openfeature_rust)

![badge](https://img.shields.io/crates/v/open-feature-ofrep?color=blue\&style=flat-square\&logo=rust)

* **Remote Evaluation
* **Tracking Flag Evaluation
* **In process Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_rust)

**

##### [NestJS](https://gofeatureflag.org/sdk/server_providers/openfeature_nestjs)

![badge](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-provider?color=blue\&style=flat-square\&logo=npm)

* **In process Evaluation
* **Remote Evaluation
* **Tracking Flag Evaluation
* **Tracking Custom Events
* **Configuration Change Updates
* **Provider Events Reactions

[More details**](https://gofeatureflag.org/sdk/server_providers/openfeature_nestjs)

### OpenFeature Remote Evaluation Protocol (OFREP)[​](#openfeature-remote-evaluation-protocol-ofrep "Direct link to OpenFeature Remote Evaluation Protocol (OFREP)")

In addition to the GO Feature Flag specific providers, we also support the OpenFeature Remote Evaluation Protocol (OFREP) which is a standard protocol for feature flag evaluation.

![logo](/assets/images/logo-538d39dbc6512181cc954cebe1a7579b.png)

##### [OFREP](https://gofeatureflag.org/sdk/ofrep)

* **OpenFeature (OFREP) API

[More details**](https://gofeatureflag.org/sdk/ofrep)
