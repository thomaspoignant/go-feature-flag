---
sidebar_position: 30
description: Scheduled rollout introduces the ability for users to changes their flags for future points in time.
---
# ⏱️ Scheduled rollout

## Overview
Scheduled rollout offer a structured, multi-stage approach to flag deployment, automating the rollout to specific environments and audiences.

At each stage happening on a specific date, you can modify your feature flags as you want, allowing to change the target audience of your flag.    
While this sounds deceptively straightforward, it unlocks the potential for users to create complex release strategies by scheduling the incremental steps in advance.

_For example, you may want to turn a feature ON for internal testing tomorrow and then enable it for your ‘beta’ user segment four days later._

## Define a scheduled rollout
To use a scheduled rollout, you need to set the `scheduledRollout` field in your flag configuration.

This field is an array of **steps**, each step is a map with a `date` field and the other fields are the fields you want to update in your flag.

## Format

:::info
You can change any fields that are available on your flag.
Since your configuration has not been changed manually, it does not trigger any notifier.
:::

| Field       | Description                                                                                                                                                                                                                                                                                                                 |
|-------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **`steps`** | The only mandatory field in a **step** is the `date`.<br/>**If no date is provided the step will be skipped.**<br/><br/>The other attributes of your `step` are what you want to update your flag, so every field available in the flag format can be updated.<br/>The new value in a field will override the existing one. |

## Example
In this example, the flag will be update multiple times:
1. We will add a targeting rule to target beta users at `2020-04-10T00:00:00.1+02:00`.
2. We will update the query of this rule later to target non-beta users at `2022-05-12T15:36:00.1+02:00`.

```yaml
scheduled-flag:
  variations:
    variationA: A
    variationB: B
  defaultRule:
    name: defaultRule
    percentage:
      variationA: 100
      variationB: 0
# highlight-start
  scheduledRollout:
    - date: 2020-04-10T00:00:00.1+02:00
      targeting:
        - name: rule1
          query: beta eq "true"
          percentage:
            variationA: 0
            variationB: 100

    - date: 2022-05-12T15:36:00.1+02:00
      targeting:
        - name: rule1
          query: beta eq "false"
# highlight-end
```

