---
description: Use GO Feature Flag with Openfeature SDKs
title: OpenFeature Remote Evaluation Protocol (OFREP)
---
# OpenFeature Remote Evaluation Protocol (OFREP)

## Overview
In addition of the GO Feature Flag specific providers, we also support the OpenFeature Remote Evaluation Protocol (OFREP) which is a standard protocol for feature flag evaluation.
This protocol is designed to be used with any feature flag provider that supports it.

## What is OFREP?
OFREP is an API specification that allows you to connect your applications to any feature flag management system that supports the protocol using generic providers.
This means more flexibility and choice when it comes to managing your features.

For the full technical details, you can head over to the Open Feature Protocol documentation: Open Feature Protocol: https://github.com/open-feature/protocol/.

## When to use OFREP instead of the official GO Feature Flag providers?
- When you want to be able to switch between different feature flag vendors without changing your code.
- When what you need is only the evaluation and you don't need the full feature flag management capabilities _(collect of flag evaluations, provider specific function for GO Feature Flag, etc ...)_.
- When you want to avoid vendor lock-in as much as possible.
- When you have a software that wants to integrate with different vendors, without importing all the providers of those vendors.
- When you want to use a SDK that has support of OFREP but not the official GO Feature Flag providers.

## How to use OFREP with GO Feature Flag?
To use OpenFeature Remote Evaluation Protocol (OFREP) with GO Feature Flag, you need to have the [`relay-proxy`](../relay-proxy) running in your infrastructure.

The `relay-proxy` has all the endpoints needed to communicate with any of the feature flag providers that support OFREP.

If you want to look at those endpoints, you can look at them in the [API doc](https://gofeatureflag.org/API_relayproxy#tag/OpenFeature-Remote-Evaluation-Protocol-(OFREP)).

## Any questions about OFREP?
Ask them on [Slack](/slack) or open an issue on the [GitHub repository](https://github.com/thomaspoignant/go-feature-flag/issues).