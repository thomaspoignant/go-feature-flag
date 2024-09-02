---
title: Changes to golang OpenFeature Integration
description: We have released 2 new golang providers for OpenFeature, let's explain all the changes.
authors: [thomaspoignant]
tags: [openfeature,golang]
image: https://gofeatureflag.org/assets/images/blog_cover-551c6a50d204cc3316f7a43ae59625e3.png
---
![blog_cover.png](blog_cover.png)
We're excited to announce significant improvements to the integration between GO Feature Flag and OpenFeature.

To better serve our users' diverse needs, we've decided to split the previous Go provider into two distinct providers:
- [`go-feature-flag`](https://github.com/open-feature/go-sdk-contrib/tree/main/providers/go-feature-flag): For use with the GO Feature Flag relay proxy.
- [`go-feature-flag-in-process`](https://github.com/open-feature/go-sdk-contrib/tree/main/providers/go-feature-flag-in-process): For direct integration of GO Feature Flag into your application.

<!--truncate-->

## Why the Split?

The decision to create separate providers was driven by several factors:
- **Distinct Use Cases:** It's uncommon for users to require both the relay proxy and in-process capabilities simultaneously.
- **Dependency Management:** The in-process provider introduces additional dependencies, which can complicate integration.
- **Maintainability:** Managing two distinct functionalities within a single provider proved challenging.

By separating these functionalities, we've streamlined the integration process and enhanced overall maintainability.

We also use this split to rewrite from scratch the provider to increase the quality of the OpenFeature integration.

## Choosing the Right Provider

To select the appropriate provider for your application, consider the following:
- [`go-feature-flag`](https://github.com/open-feature/go-sdk-contrib/tree/main/providers/go-feature-flag): Ideal for multi-language environments using the GO Feature Flag relay proxy. This provider offers a lightweight integration with minimal dependencies, ensuring consistency with other language-specific providers.
- [`go-feature-flag-in-process`](https://github.com/open-feature/go-sdk-contrib/tree/main/providers/go-feature-flag-in-process): Best suited for Go-only applications that prefer a direct integration without the overhead of a separate service.

## Breaking Changes
As you may imagine, this split comes with some breaking changes, but we've done everything possible to limit it as much as possible.
Please be aware of the following breaking changes:

- **Existing go-feature-flag users:** If you were previously using the `go-feature-flag` provider with the Go module, you'll need to transition to the `go-feature-flag-in-process` provider.
- **Configuration Changes:** The `HTTPClient` interface has been removed from the configuration in favor of using the standard `http.Client` _(This should have no impact on the way to pass the http client in the options)_. 
- **Cache Compatibility:** If you are using the in-memory cache of the provider, ensure your relay proxy is updated to version `1.32.0` or later, if not the cache will not work as expected and you may have an increase on the number of connection to the relay-proxy.

We understand that these changes may require adjustments to your existing setup. We recommend carefully reviewing the updated documentation for both providers to ensure a smooth transition.

## Next Steps

We encourage you to explore the new providers and leverage the enhanced features they offer. If you encounter any issues or have feedback, please don't hesitate to [open an issue](https://github.com/thomaspoignant/go-feature-flag/issues/new/choose).

We're committed to providing you with the best possible feature flagging experience.