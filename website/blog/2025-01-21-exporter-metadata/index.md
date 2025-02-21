---
title: Introducing Exporter Metadata in the GO Feature Flag Provider for OpenFeature.
description: Exporter metadata in the GO Feature Flag provider for OpenFeature allows enriching evaluation events with static context data like environment or app version for improved observability and analysis.
authors: [thomaspoignant]
tags: [openfeature,exporter]
---

# Introducing Exporter Metadata in the GO Feature Flag Provider for OpenFeature

This blog post dives into a new functionality introduced in the GO Feature Flag provider for OpenFeature: exporter metadata. We'll explore how it works and how you can leverage it to enrich your evaluation events with valuable static information.

## What is Exporter Metadata?
Exporter metadata is an object you can configure when initializing your GO Feature Flag provider for OpenFeature.

It allows you to specify a set of static information that you want to consistently include with all your evaluation events.
A good example of information you might include in exporter metadata is the environment in which your application is running,
the version of the application or anything that helps you when you analyze your evaluation data.

This data is then forwarded to the exporter and incorporated into your feature events as a new field called `metadata`.
<!--truncate-->
## Why Use Exporter Metadata?

Including exporter metadata in your evaluation events offers several advantages:
- **Enhanced Context**: By adding static information to your events, you can provide richer context for analysis. This can include details about the environment (e.g., development, staging, production), the application version, or any other relevant data that sheds light on feature usage patterns.
- **Improved Observability**: With exporter metadata, you gain a more comprehensive view of feature evaluation across different contexts. This can be particularly useful for debugging purposes or identifying trends in feature adoption for specific environments or application versions.
- **Simplified Data Enrichment**: Exporter metadata eliminates the need to manually inject context information into every evaluation event. This streamlines your code and ensures consistency in the data you collect.

## How to Use Exporter Metadata

In all providers that support exporter metadata, you can pass a map of key-value pairs to the `ExporterMetadata` field in the provider's configuration.
Value can be a `string`, a `bool`, an `int` or a `float64`.

This map will be included in all evaluation events sent to the exporter and added to the `metadata` field by the exporter.

Here's an example of how you can configure exporter metadata in the Kotlin provider:
```kotlin
val options =
    GoFeatureFlagOptions(
        endpoint = "https://mygoffserver.com/",
        exporterMetadata = mapOf("device" to "Pixel 4", "appVersion" to "1.0.0")
    )

val provider = GoFeatureFlagProvider(options)
```
In this example, we're passing some information that can help to understand the context of the evaluation events such as the device type and the appVersion.

With this configuration, all evaluation events sent to the exporter will include the `metadata` field with the specified key-value pairs.
And it will look like this:

```json
{"kind":"feature","contextKind":"user","userKey":"4f433951-4c8c-42b3-9f18-8c9a5ed8e9eb","creationDate":1737465936,"key":"REDIRECTION","variation":"on","value":true,"default":false,"version":"","source":"PROVIDER_CACHE", "metadata": {"device": "Pixel 4", "appVersion":"1.0.0", "openfeature": true, "provider": "android"}}
```

## Which Providers Support Exporter Metadata?
To be able to use exporter metadata, you need to use the relay-proxy in version `v1.41.1` or above and as minimum the following version of the providers:
- **Kotlin**: `v0.3.0`
- **Web**: `v0.2.4`
- **Swift**: `v0.3.0`
- **Java**: `v0.4.2`
- **NodeJS**: `v0.7.5`
- **PHP**: `1.1.0`
- **Ruby**: `0.1.3`
- **Python**: `v0.4.1`
- **.NET**: `0.2.1`
- **GO**: `v0.2.3`

## Conclusion
Exporter metadata provides a powerful mechanism to enrich your evaluation events with static information in the GO Feature Flag provider for OpenFeature.
By incorporating this functionality, you can gain deeper insights into feature usage patterns and streamline your data collection process.

This blog post has provided a brief introduction to exporter metadata.
We encourage you to experiment with this feature and explore its potential to enhance your feature flag evaluation practices.
