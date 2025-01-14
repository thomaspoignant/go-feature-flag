---
sidebar_position: 10
description: A percentage rollout refers to a gradual release of a new feature to a subset of users.
---
# 💯 Percentage rollout

## Overview

A percentage rollout refers to a gradual release of a new feature to a subset of users.
Instead of making the feature available to everyone at once, you want to have only a subset of your customers to see the new feature.

## Define a percentage rollout

In GO Feature Flag, you can define a percentage rollout by setting the percentage of users that will see the new feature.
You can define the percentage in the flag configuration, and it works for the default rule or any other targeting rule you define.

The expected format is a map where the **key is the variation name** and the **value is the percentage** of users that will see this variation.

You can have as many variations as you want in the percentages, and you can define the percentage of users that will see each variation.

:::note
If the sum of your percentages is not equal to 100 this is not a problem, we are computing the affectation as fractional numbers.

_**example**: if you have 2 variations and you set `variation1: 10` and `variation2: 50` it will be equivalent to `variation1: 10/60` and `variation2: 50/60`._
:::


## Examples
### Define in the default rule
In this example `99%` of the users will see the `disabled` variation and `1%` will see the `enabled` variation.

```yaml title="flag-config.goff.yaml"
percentage-flag-1:
  variations:
    disabled: false
    enabled: true
  defaultRule:
  # highlight-start
    percentage:
      disabled: 99
      enabled: 1
  # highlight-end
```

### Define in a targeting rule
In this second example, `50%` of the users that are `admin` will see the `disabled` variation and `50%` will see the `enabled` variation.
Non admin users will see the `disabled` variation.
```yaml title="flag-config.goff.yaml"
percentage-flag-1:
  variations:
    disabled: false
    enabled: true
  targeting:
    - query: admin eq true
      # highlight-start
      percentage:
        disabled: 50
        enabled: 50
      # highlight-end
  defaultRule:
    variation: disabled

```

## 🐥 Canary releases
Percentage rollout are also known as **canary releases**, where you release a new feature to a small subset of users before rolling it out to everyone.
