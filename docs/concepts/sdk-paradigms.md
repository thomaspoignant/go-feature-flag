# 🧩 SDK paradigms

## Overview[​](#overview "Direct link to Overview")

GO Feature Flag in association with OpenFeature can be used in 2 "flavors"; those intended for use in server applications and those intended for use in client applications.

The primary difference between these two paradigms has to do with the way they model evaluation context.

## Dynamic Context Paradigm (Server-side SDKs)[​](#dynamic-context-paradigm-server-side-sdks "Direct link to Dynamic Context Paradigm (Server-side SDKs)")

In server usage, the evaluation context changes frequently, as often as every evaluation. Though some context data may remain static for the entire application lifecycle (such as the `hostname` or `application version`) much of it may be dynamic.

For instance, on a web-server, the context might contain the **remote IP** of the connected client or the **session-id** or **email address** of the logged-in user.<br /><!-- -->For this reason, the evaluation API defines a parameter for the evaluation context which can be provided at each evaluation.

## Static Context Paradigms (Client-side SDKs)[​](#static-context-paradigms-client-side-sdks "Direct link to Static Context Paradigms (Client-side SDKs)")

In client-side usage, the evaluation context changes less frequently, often in response to user actions or UI events.

In this case, the context frequently corresponds to a single user, and the context data represents facts about them *(such as their name or subscription level)* or the application state (if the user has items in their cart).

When the context data changes, it's often necessary for the provider to update it's cache of evaluated flags; we refer to this as context reconciliation.

## GO Feature Flag support[​](#go-feature-flag-support "Direct link to GO Feature Flag support")

GO Feature Flag supports both paradigms, and the SDKs are designed to make it easy to use the same feature flags in both server and client applications.

The goal of GO Feature Flag is to bring a top level support of all the languages supported.

#### Client-side Languages supported:[​](#client-side-languages-supported "Direct link to Client-side Languages supported:")

* Javascript
* Typescript
* React
* Angular
* Swift
* Kotlin / Android

#### Server-side Languages supported:[​](#server-side-languages-supported "Direct link to Server-side Languages supported:")

* Golang
* Java
* Kotlin
* .NET
* Python
* Node.JS
* PHP
* Ruby
* Rust
* NestJS

## Want to know more?[​](#want-to-know-more "Direct link to Want to know more?")

A good read about client / server differences can be found in the [OpenFeature Blog](https://openfeature.dev/blog/catering-to-the-client-side).
