---
sidebar_position: 10
description: How GO Feature Flag is working with OpenFeature.
---

# Concepts

## What is OpenFeature?
:::note
OpenFeature is an open standard for feature flag management, created to support a robust feature flag ecosystem using cloud native technologies. OpenFeature provides a unified API and SDK, and a developer-first, cloud-native implementation, with extensibility for open source and commercial offerings.

https://docs.openfeature.dev/docs/category/concepts
:::

OpenFeature offer a framework-agnostic way of using feature flags, it means that if you use OpenFeature SDKs you will have minimum changes to do if you want to choose another provider.

To support this initiative, **GO Feature Flag** does not have any SDKs and rely 100% on OpenFeature SDKs.  
To be compatible with our solution, we offer [`providers`](https://docs.openfeature.dev/docs/reference/concepts/provider) for GO Feature Flag in several languages. 


## How OpenFeature and GO Feature Flag are working together?

To use the OpenFeature SDKs you need what we call a provider.  
A **provider** is responsible for performing flag evaluations. It provides an abstraction between **GO Feature Flag** and the OpenFeature SDK.

A provider need a backend service to perform the flag evaluation. This is why we have introduced the [**relay proxy**](../category/use-the-relay-proxy).  
This component retrieve your feature flag configuration file using the GO module and expose APIs to get your flags variations.

![](/docs/openfeature/concepts.jpg)

With this simple architecture you have a central component _(the relay proxy)_ that is in charge of the flag evaluation, while the SDKs and providers are responsible to communicate with the relay proxy.

This supports different languages the same way and makes you able to use GO Feature Flag with all your favorite languages.
