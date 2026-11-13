# Rust

[![crates.io](https://img.shields.io/crates/v/open-feature-ofrep?color=blue\&style=flat-square\&logo=rust)](https://github.com/open-feature/rust-sdk-contrib/tree/main/crates/ofrep)

The OpenFeature OFREP provider connects the OpenFeature Rust SDK to a GO Feature Flag relay-proxy.

It supports a single evaluation mode:

* `REMOTE` delegates every evaluation to the relay-proxy via OFREP: the relay-proxy remains the single source of truth, no local configuration or polling is needed, and flag changes take effect immediately on the server.

This is not a GO Feature Flag specific provider

There is no GO Feature Flag specific provider for Rust. The integration is done through the **generic [OpenFeature OFREP provider](https://github.com/open-feature/rust-sdk-contrib/tree/main/crates/ofrep)** (`open-feature-ofrep` crate), which works with any backend that implements the [OpenFeature Remote Evaluation Protocol (OFREP)](https://gofeatureflag.org/docs/sdk/ofrep.md).

Because the [GO Feature Flag relay-proxy](https://gofeatureflag.org/docs/relay-proxy.md) implements OFREP, you can use this provider to evaluate your GO Feature Flag flags from any Rust application. As a consequence, only the OFREP-compatible features are available — see [Features status](#features-status) below.

## Install dependencies[​](#install-dependencies "Direct link to Install dependencies")

Install the provider and the OpenFeature Rust SDK:

```
cargo add open-feature

cargo add open-feature-ofrep
```

## Initialize your provider[​](#initialize-your-provider "Direct link to Initialize your provider")

Create an `OfrepProvider` pointing at your GO Feature Flag relay-proxy:

```
use std::time::Duration;

use open_feature_ofrep::{OfrepProvider, OfrepOptions};

use reqwest::header::HeaderMap;



let provider = OfrepProvider::new(OfrepOptions {

    base_url: "http://localhost:1031".to_string(),

    headers: HeaderMap::new(),

    connect_timeout: Duration::from_secs(10),

    ..Default::default()

})

.await

.expect("failed to create OFREP provider");
```

## Initialize your OpenFeature client[​](#initialize-your-openfeature-client "Direct link to Initialize your OpenFeature client")

Register the provider in the OpenFeature SDK and then create a client:

```
use std::time::Duration;

use open_feature::OpenFeature;

use open_feature_ofrep::{OfrepProvider, OfrepOptions};

use reqwest::header::HeaderMap;



let provider = OfrepProvider::new(OfrepOptions {

    base_url: "http://localhost:1031".to_string(),

    headers: HeaderMap::new(),

    connect_timeout: Duration::from_secs(10),

    ..Default::default()

})

.await

.expect("failed to create OFREP provider");



let mut api = OpenFeature::singleton_mut().await;

api.set_provider(provider).await;



let client = api.create_client();
```

## Evaluate a flag[​](#evaluate-a-flag "Direct link to Evaluate a flag")

Create an `EvaluationContext` and use the OpenFeature client as usual.

> In this example we evaluate a boolean flag, but the provider also supports string, integer, float, and struct evaluations.
>
> See the [OpenFeature evaluation API documentation](https://openfeature.dev/docs/reference/concepts/evaluation-api#basic-evaluation) for the full API surface.

```
use open_feature::EvaluationContext;



let evaluation_ctx = EvaluationContext::default()

    .with_targeting_key("1d1b9238-2591-4a47-94cf-d2bc080892f1")

    .with_custom_field("firstname", "john")

    .with_custom_field("lastname", "doe")

    .with_custom_field("email", "john.doe@gofeatureflag.org")

    .with_custom_field("admin", true)

    .with_custom_field("anonymous", false);



let admin_flag = client

    .get_bool_value("flag-only-for-admin", Some(&evaluation_ctx), None)

    .await

    .unwrap_or(false);



if admin_flag {

    // flag "flag-only-for-admin" evaluated to true

} else {

    // flag "flag-only-for-admin" evaluated to false

}
```

## Provider options[​](#provider-options "Direct link to Provider options")

`base_url` is required (a default is provided but it almost certainly does not match your relay-proxy). The other options are optional.

| Option            | Type        | Default                 | Description                                                                                     |
| ----------------- | ----------- | ----------------------- | ----------------------------------------------------------------------------------------------- |
| `base_url`        | `String`    | `http://localhost:8016` | Base URL of the GO Feature Flag relay-proxy (e.g. `http://localhost:1031`).                     |
| `headers`         | `HeaderMap` | empty                   | Extra headers added to provider HTTP requests. Use this to send `X-API-Key` or `Authorization`. |
| `connect_timeout` | `Duration`  | `10s`                   | Connection timeout for HTTP requests to the relay-proxy.                                        |

To send an API key:

```
use reqwest::header::{HeaderMap, HeaderValue};



let mut headers = HeaderMap::new();

headers.insert("X-API-Key", HeaderValue::from_static("my-api-key"));
```

Flag sets require `X-API-Key`

If your relay-proxy is configured with [flag sets](https://gofeatureflag.org/docs/relay-proxy/configure-relay-proxy.md#use-multiple-flag-sets), you **must** send your API key in the `X-API-Key` header. The relay-proxy uses the API key to resolve which flag set the request belongs to, so requests without it (or using a different auth header) will not be routed to the right flag set.

```
headers.insert("X-API-Key", HeaderValue::from_static("my-api-key"));
```

If your relay-proxy expects bearer-token auth instead of `X-API-Key`, use `Authorization`:

```
headers.insert(

    "Authorization",

    HeaderValue::from_static("Bearer <token>"),

);
```

## Tracking and event collection[​](#tracking-and-event-collection "Direct link to Tracking and event collection")

### Flag usage events (automatic)[​](#flag-usage-events-automatic "Direct link to Flag usage events (automatic)")

Every flag evaluation goes through the relay-proxy via OFREP, so the relay-proxy itself records each evaluation as part of its built-in [export evaluation data](https://gofeatureflag.org/docs/integrations/export-evaluation-data.md) pipeline. No additional configuration is required on the Rust side — once an exporter is configured on the relay-proxy, every Rust evaluation will be exported alongside evaluations from other languages.

### Custom tracking events (OpenFeature Track API)[​](#custom-tracking-events-openfeature-track-api "Direct link to Custom tracking events (OpenFeature Track API)")

The OFREP protocol does not yet specify a tracking endpoint. As a result, `client.track()` calls made through this provider are **not** forwarded to the relay-proxy. If you need custom tracking events from a server application, use a GO Feature Flag specific provider (Go, Java, Python, .NET, ...).

## Operational notes[​](#operational-notes "Direct link to Operational notes")

* Every evaluation is a network call to the relay-proxy via OFREP. There is no local cache or polling, so flag changes are picked up immediately on the next evaluation.
* The provider does not emit `ProviderConfigChange` / `ProviderStale` events — OFREP has no streaming/push channel from the relay-proxy to the client.
* In-process evaluation is not available for Rust today; if you need lower evaluation latency, host the relay-proxy close to your Rust workload.

## Features status[​](#features-status "Direct link to Features status")

| Status | Feature                      | Description                                                                                                                                                                |
| ------ | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **     | Remote Evaluation            | The provider is calling the remote server to evaluate the feature flags.                                                                                                   |
| **     | Tracking Flag Evaluation     | The provider is tracking all the evaluations of your feature flags and you can export them using an exporter.                                                              |
| **     | In process Evaluation        | The provider is able to evaluate the feature flags in process, it loads the configuration from the relay-proxy and perform evaluation inside the SDK.                      |
| **     | Tracking Custom Events       | The provider is tracking custom events through the track() function of your SDK. All those events are send to the exporter for you to forward them where you want.         |
| **     | Configuration Change Updates | The provider is able to update the configuration based on the configuration, it means that the provider is able to react to any feature flag change on your configuration. |
| **     | Provider Events Reactions    | You can add an event handler to the provider to react to the provider events.                                                                                              |

**Implemented**In-progress**Not implemented yet

## Contribute to the provider[​](#contribute-to-the-provider "Direct link to Contribute to the provider")

You can find the source of the OFREP provider in the [`open-feature/rust-sdk-contrib`](https://github.com/open-feature/rust-sdk-contrib/tree/main/crates/ofrep) repository — please open issues or pull requests there for provider-level changes.
