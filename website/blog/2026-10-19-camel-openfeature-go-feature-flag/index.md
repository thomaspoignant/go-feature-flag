---
title: 'Feature Flags in Apache Camel Routes with OpenFeature and GO Feature Flag'
description: 'Wire camel-openfeature (Camel 4.23) to the GO Feature Flag Java provider and the relay proxy to toggle routes, roll out new endpoints, and add kill switches.'
authors: [thomaspoignant]
tags: [openfeature, apache-camel, java, relay-proxy, tutorial]
image: ./banner.jpg
image_alt: 'An orange Apache Camel silhouette connected by a dotted line to a green feature-flag toggle'
category: Engineering
---

![Feature flags in Apache Camel with GO Feature Flag](./banner.jpg)

Apache Camel 4.23 shipped [`camel-openfeature`](https://camel.apache.org/components/next/openfeature-component.html): a producer (and a matching language) that evaluates feature flags inside your routes. Out of the box it talks to [flagd](https://flagd.dev/). Because it is built on the [OpenFeature](https://openfeature.dev/) Java SDK, you can plug in any OpenFeature provider — including the official [GO Feature Flag Java provider](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_java).

This post shows how to run a GO Feature Flag relay proxy, register the GOFF provider with Camel, and use flags to branch routes, roll out a new endpoint gradually, and keep a kill switch handy.

<!-- truncate -->

## Why this matters in an integration layer

Camel routes are where messages fork: enrich or skip, call the new HTTP endpoint or the old one, transform with v2 or stay on v1. Those decisions used to mean a redeploy or a config reload. Feature flags move them into runtime configuration.

Typical patterns:

- **Toggle / detour** — evaluate a boolean flag and `choice()` between two branches (enrich vs skip).
- **Gradual rollout** — evaluate a string (variant) flag with a percentage rule so a share of traffic hits a new endpoint or transformation.
- **Kill switch** — a boolean that short-circuits a route when something goes wrong, without rolling back the whole deployment.

OpenFeature keeps the evaluation API vendor-neutral. Camel owns the EIP wiring; GO Feature Flag owns targeting, percentages, and the flag file format you already use elsewhere.

## Prerequisites

- **Apache Camel 4.23.0** or later (the OpenFeature component was added in 4.23).
- A **GO Feature Flag relay proxy** reachable from the JVM (Docker is enough for local work).
- **JDK 17+** (Camel 4.x baseline; this walkthrough was checked against JDK 21).

## 1. Flags and a local relay proxy

Create a small flag file in GO Feature Flag format (YAML). The relay proxy will serve it to the Java provider.

```yaml title="flags.goff.yaml"
# Boolean toggle: enrich an order or skip enrichment
enrichment-enabled:
  variations:
    enabled: true
    disabled: false
  targeting:
    - name: enterprise customers
      query: customer_tier eq "ENTERPRISE"
      variation: enabled
  defaultRule:
    variation: disabled

# Percentage rollout between two route targets (string variants)
routing-algorithm:
  variations:
    legacy: content-based-router
    new: dynamic-router
  defaultRule:
    percentage:
      legacy: 90
      new: 10

# Kill switch for the main processing path
order-processing-kill-switch:
  variations:
    live: true
    stopped: false
  defaultRule:
    variation: live
```

Point the relay proxy at that file with a file retriever:

```yaml title="goff-proxy.yaml"
retrievers:
  - kind: file
    path: /goff/flags.goff.yaml
```

Start the proxy (port `1031` is the default listen port):

```bash
docker pull gofeatureflag/go-feature-flag:latest

docker run --rm \
  -v "$(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml" \
  -v "$(pwd)/flags.goff.yaml:/goff/flags.goff.yaml" \
  -p 1031:1031 \
  gofeatureflag/go-feature-flag
```

You can smoke-test with OFREP:

```bash
curl -s 'http://localhost:1031/ofrep/v1/evaluate/flags/enrichment-enabled' \
  -H 'Content-Type: application/json' \
  -d '{"context":{"targetingKey":"order-1","customer_tier":"ENTERPRISE"}}'
```

More detail: [relay proxy getting started](/docs/relay-proxy/getting_started) and the [file retriever](/docs/integrations/store-flags-configuration/file).

## 2. Maven dependencies

Use the same Camel version for core and the OpenFeature component. Add the GO Feature Flag provider:

```xml
<properties>
  <camel.version>4.23.0</camel.version>
</properties>

<dependencies>
  <dependency>
    <groupId>org.apache.camel</groupId>
    <artifactId>camel-main</artifactId>
    <version>${camel.version}</version>
  </dependency>
  <dependency>
    <groupId>org.apache.camel</groupId>
    <artifactId>camel-direct</artifactId>
    <version>${camel.version}</version>
  </dependency>
  <dependency>
    <groupId>org.apache.camel</groupId>
    <artifactId>camel-openfeature</artifactId>
    <version>${camel.version}</version>
  </dependency>
  <dependency>
    <groupId>dev.openfeature.contrib.providers</groupId>
    <artifactId>go-feature-flag</artifactId>
    <version>2.0.2</version>
  </dependency>
</dependencies>
```

`camel-openfeature` already brings the OpenFeature Java SDK (`dev.openfeature:sdk`, 1.23.0 in Camel 4.23.0). The GOFF provider declares the SDK as `provided` with range `[1.21.0,1.99999)`, so that transitive dependency is enough — you do not need a second SDK declaration unless you want to pin it yourself.

Versions above were checked on Maven Central at the time of writing (`camel-openfeature` 4.23.0, `go-feature-flag` 2.0.2). Prefer the latest patch of each artifact when you copy this into a real project.

## 3. Register the GOFF provider with Camel

Camel resolves the OpenFeature provider in this order (from the [component docs](https://camel.apache.org/components/next/openfeature-component.html)):

1. Explicit `provider` option (bean reference such as `#myProvider`)
2. A registry bean named `flags` of type `FeatureProvider`
3. Built-in flagd (`flags`, `flagsResource`, or `host`)

For GO Feature Flag, build a `GoFeatureFlagProvider` and bind it as `flags` (or pass `provider=#…` on the endpoint). Camel calls `OpenFeatureAPI.setProviderAndWait(domain, provider)` when the endpoint starts, so the provider’s `initialize` runs against your relay proxy.

```java
import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProvider;
import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProviderOptions;
import dev.openfeature.contrib.providers.gofeatureflag.bean.EvaluationType;
import dev.openfeature.sdk.FeatureProvider;
import org.apache.camel.main.Main;

public class CamelGoffApplication {
    public static void main(String[] args) throws Exception {
        FeatureProvider goff = new GoFeatureFlagProvider(
                GoFeatureFlagProviderOptions.builder()
                        .endpoint("http://localhost:1031")
                        // Default is IN_PROCESS: flag config is synced from the
                        // relay proxy and evaluated in the JVM. Use REMOTE to
                        // call the proxy on every evaluation instead.
                        .evaluationType(EvaluationType.IN_PROCESS)
                        .build());

        Main main = new Main();
        main.configure().addRoutesBuilder(new OrderRoutes());
        main.bind("flags", goff);
        main.run(args);
    }
}
```

Useful `GoFeatureFlagProviderOptions` fields (all verified against provider 2.0.2):

| Option | Required | Notes |
| --- | --- | --- |
| `endpoint` | yes | Relay proxy base URL, e.g. `http://localhost:1031` |
| `evaluationType` | no | `IN_PROCESS` (default) or `REMOTE` |
| `apiKey` | no | If the relay proxy requires API key auth |
| `timeout` | no | HTTP timeout in ms (default `10000`) |
| `flagChangePollingIntervalMs` | no | How often in-process mode polls for config changes (default `120000`) |

Full option list: [Java provider docs](/docs/sdk/server_providers/openfeature_java) and the [provider README](https://github.com/open-feature/java-sdk-contrib/tree/main/providers/go-feature-flag).

You can also set the bean at component level:

```properties
camel.component.openfeature.provider = #flags
```

Or per endpoint: `openfeature:flags?flagKey=…&provider=#flags`.

## 4. Route examples (Java DSL)

URI shape (Camel docs): `openfeature:domain[/evaluationType][?options]`.  
`domain` is the OpenFeature client domain (use `flags` when you bound the bean as `flags`). Optional `evaluationType`: `boolean`, `isEnabled`, or `variant`.

Headers you will use most:

- `CamelOpenFeatureTargetingKey` — subject id for consistent bucketing (required by GOFF for meaningful evaluation)
- `CamelOpenFeatureEvaluationContext` — `Map<String, Object>` of extra attributes (e.g. `customer_tier`)
- `resultProperty` — store the flag value in an exchange property and keep the message body

### Boolean branch (toggle / detour)

```java
from("direct:process-order")
    .setHeader("CamelOpenFeatureTargetingKey", simple("${header.orderId}"))
    .to("openfeature:flags?flagKey=enrichment-enabled&resultProperty=enrichEnabled")
    .choice()
        .when(exchangeProperty("enrichEnabled").isEqualTo(true))
            .to("direct:enrich-order")
        .otherwise()
            .to("direct:skip-enrichment")
    .end();
```

For a pure boolean predicate without writing a property, use the OpenFeature language:

```java
from("direct:process-order")
    .setHeader("CamelOpenFeatureTargetingKey", simple("${header.orderId}"))
    .filter().language("openfeature", "enrichment-enabled")
        .to("direct:enrich-order")
    .end();
```

### Kill switch

```java
from("direct:orders")
    .setHeader("CamelOpenFeatureTargetingKey", simple("${header.orderId}"))
    .to("openfeature:flags?flagKey=order-processing-kill-switch&resultProperty=processingLive")
    .choice()
        .when(exchangeProperty("processingLive").isEqualTo(false))
            .log("Kill switch active — dropping order ${header.orderId}")
            .stop()
        .otherwise()
            .to("direct:process-order")
    .end();
```

Flip `order-processing-kill-switch`’s default rule to `stopped` in `flags.goff.yaml` (or via your usual flag delivery) and the route stops accepting work without a redeploy.

### Percentage rollout (string / variant)

GO Feature Flag percentage rules hash the flag key with the targeting key, so the same order always gets the same variant:

```java
from("direct:route-order")
    .setHeader("CamelOpenFeatureTargetingKey", simple("${header.orderId}"))
    .to("openfeature:flags/variant?flagKey=routing-algorithm&resultProperty=algorithm")
    .toD("direct:route-${exchangeProperty.algorithm}");

from("direct:route-content-based-router")
    .log("Legacy router for ${header.orderId}");

from("direct:route-dynamic-router")
    .log("New router for ${header.orderId}");
```

With the sample flag file, about 10% of distinct `orderId`s land on `dynamic-router` and 90% on `content-based-router`. Raise `new` in the YAML when you are ready for more traffic.

### Targeting with evaluation context

The `enrichment-enabled` flag enables enrichment for `customer_tier eq "ENTERPRISE"`. Pass that attribute in the evaluation context header:

```java
from("direct:process-order")
    .setHeader("CamelOpenFeatureTargetingKey", simple("${header.orderId}"))
    .process(exchange -> {
        exchange.getMessage().setHeader(
            "CamelOpenFeatureEvaluationContext",
            Map.of("customer_tier", exchange.getMessage().getHeader("customerTier", String.class)));
    })
    .to("openfeature:flags?flagKey=enrichment-enabled&resultProperty=enrichEnabled")
    .choice()
        .when(exchangeProperty("enrichEnabled").isEqualTo(true))
            .to("direct:enrich-order")
        .otherwise()
            .to("direct:skip-enrichment")
    .end();
```

## 5. Same routes in YAML DSL (optional)

Once the `flags` bean is registered (Java `Main.bind`, Spring `@Bean`, etc.), YAML routes look like the Camel docs examples:

```yaml
- route:
    from:
      uri: direct:process-order
      steps:
        - setHeader:
            name: CamelOpenFeatureTargetingKey
            expression:
              simple:
                expression: "${header.orderId}"
        - to:
            uri: openfeature:flags
            parameters:
              flagKey: enrichment-enabled
              resultProperty: enrichEnabled
        - choice:
            when:
              - expression:
                  simple:
                    expression: "${exchangeProperty.enrichEnabled} == true"
                steps:
                  - to:
                      uri: direct:enrich-order
            otherwise:
              steps:
                - to:
                    uri: direct:skip-enrichment

- route:
    from:
      uri: direct:route-order
      steps:
        - setHeader:
            name: CamelOpenFeatureTargetingKey
            expression:
              simple:
                expression: "${header.orderId}"
        - to:
            uri: openfeature:flags/variant
            parameters:
              flagKey: routing-algorithm
              resultProperty: algorithm
        - toD:
            uri: "direct:route-${exchangeProperty.algorithm}"
```

`GoFeatureFlagProvider` needs constructor options (`endpoint`, …), so prefer creating that bean in Java (or Spring) rather than a no-arg YAML bean definition.

## 6. What you get from GOFF that flagd alone does not

flagd’s JSON definitions work fine for demos. Using GO Feature Flag as the provider means:

- One flag format (`flags.goff.yaml`) shared with your other OpenFeature SDKs
- Retrievers (GitHub, S3, Kubernetes, …) and the same relay proxy your services already use
- Targeting queries, percentage and progressive rollouts, and exporters you configure on the proxy

Camel stays OpenFeature-native either way; you are only swapping the provider bean.

## Wrap-up

1. Run the [relay proxy](/docs/relay-proxy/getting_started) with a file (or any) retriever.
2. Add `camel-openfeature` (4.23+) and `dev.openfeature.contrib.providers:go-feature-flag` (2.0.2+).
3. Bind a `GoFeatureFlagProvider` as the `flags` bean (or set `provider=#…`).
4. Evaluate with `to("openfeature:flags?…")` or `.language("openfeature", "flag-key")`, pass `CamelOpenFeatureTargetingKey` / evaluation context, and branch with `choice` / `toD`.

Further reading:

- [Camel OpenFeature component](https://camel.apache.org/components/next/openfeature-component.html)
- [Camel OpenFeature language](https://camel.apache.org/components/next/languages/openfeature-language.html)
- [Camel 4.23 what’s new](https://camel.apache.org/blog/2026/10/camel423-whatsnew/)
- [GO Feature Flag Java / Kotlin provider](/docs/sdk/server_providers/openfeature_java)
- [Create flags](/docs/configure_flag/create-flags) and [percentage rollout](/docs/configure_flag/rollout-strategies/percentage)
