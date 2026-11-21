# 🏗️ Architecture

## Overview[​](#overview "Direct link to Overview")

GO Feature Flag is a feature flagging system that allows you to manage feature flags in your application.<br /><!-- -->The architecture is really simple and efficient with the goal to be easy to use and let you experiment with feature flags in your application as fast as possible.

In most cases you have 1 microservice running the `relay-proxy` and multiple applications running the SDKs, nothing else.

note

GO Feature Flag can also be run without the `relay-proxy` if you only use the GO Module.

The relay-proxy is an API layer on top of the GO Module, so it has the same logic inside.

## Architecture concepts[​](#architecture-concepts "Direct link to Architecture concepts")

![GO Feature Flag architecture](/assets/images/architecture-709ce924cb02eb19ccce8a0daea14c67.svg)

### 🧩 OpenFeature Providers[​](#-openfeature-providers "Direct link to 🧩 OpenFeature Providers")

GO Feature Flag is working in collaboration with OpenFeature to be integrated with all the languages supported by GO Feature Flag.<br /><!-- -->To achieve this goal, we have created the OpenFeature providers that combines with OpenFeature SDKs to allow you to use GO Feature Flag in your application.

> **OpenFeature Providers**<br /><!-- -->An SDK-compliant implementation which resolves flag values from a particular flag management system, allowing the use of the Evaluation API as an abstraction for the system in question.

This is what developers will use to interact with GO Feature Flag in their application.

### ↔️ Relay-Proxy[​](#️-relay-proxy "Direct link to ↔️ Relay-Proxy")

The relay-proxy is the only component you have to run in your infrastructure to use GO Feature Flag with the different SDKs.<br /><!-- -->This is a standalone GO binary that you can run in your infrastructure to expose the GO Feature Flag API to your applications.

info

Want to know more about the relay-proxy? [Check the relay-proxy documentation](https://gofeatureflag.org/docs/relay-proxy.md).

The relay-proxy is designed to be simple, lightweight, and stateless. It operates without the need for any databases or complex systems, as it loads feature flag configuration files from a specified location and stores them in memory.

### ⬇️ Internal concepts[​](#️-internal-concepts "Direct link to ⬇️ Internal concepts")

**You have 3 main concepts in GO Feature Flag architecture:**

* [retrievers](https://gofeatureflag.org/docs/concepts/retriever.md): This is the component in charge of fetching the feature flag configuration.
* [notifiers](https://gofeatureflag.org/docs/concepts/notifier.md): This is the component in charge of notifying external systems that a flag configuration has changed.
* [exporters](https://gofeatureflag.org/docs/concepts/exporter.md): This is the component in charge of exporting the flag evaluation results for analytics or monitoring.

Those 3 concepts are the core of GO Feature Flag and allows you to integrate GO Feature Flag in your current ecosystem with a minimum of effort.<br /><!-- -->Every component is designed to be as lightweight as possible and to be easy to configure.

## 🧑‍💻 Technologies[​](#-technologies "Direct link to 🧑‍💻 Technologies")

* GO Feature Flag is written in `golang`.
* The `retrievers`, `notifiers`, and `exporters` are using official libraries to interact with external systems.
* Providers are written in the language of the system they are interacting with and we are limiting the dependencies to the minimum to be as lightweight as possible.
