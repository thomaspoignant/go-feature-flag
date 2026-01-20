---
sidebar_position: 30
description: An experimentation rollout is when your flag is configured to be served only for a determined time.
---

# 📊 Experimentation rollout

## Overview
In GO Feature Flag, an experimentation rollout is a way to test different versions of a feature within a specific timeframe.

It allows you to test a feature with a subset of users for a limited time before deciding to roll it out to everyone.

## Define an experimentation rollout
To define an experimentation rollout, you need to set the start and end dates of the rollout in the field `experimentation` of your flag configuration.

The flag will be served only between these dates, outside of this timeframe the default value will be served and the flag will be considered as not active.

## Format

:::info
The dates are in the format supported natively by your flag file format.
:::

| Field       | Description                                     |
|-------------|-------------------------------------------------|
| **`start`** | The date the flag will be started to be served. |
| **`end`**   | The date the flag will be stopped to be served. |

## Example

```yaml
experimentation-flag:
  variations:
    enabled: true
    disabled: B
  defaultRule:
    percentage:
      enabled: 50
      disabled: 50
  # highlight-start
  experimentation:
    start: 2021-03-20T00:00:00.1-05:00
    end: 2021-03-21T00:00:00.1-05:00
  # highlight-end
```

Check this [example](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples/rollout_experimentation) to see how it works.

## A/B testing

:::info
A/B test is the shorthand for a simple controlled experiment.
As the name implies, two versions (A and B) of a single variable are compared, which are identical except for one variation that might affect a user's behavior.
A/B tests are widely considered the simplest form of controlled experiment.

_**(source wikipedia)**_
:::

To have a proper A/B testing solution with the module you should use the experimentation rollout combined with the [export of your data](../../go_module/data_collection/).

This combination will allow to have your experimentation running for a dedicated time, and you will have the data to know exactly which user was on which version of the flag.

To setup the duration of your A/B test you can use a tool [ab-test-duration-calculator](https://vwo.com/tools/ab-test-duration-calculator/) from vwo, that will help you to set up the test duration correctly.
