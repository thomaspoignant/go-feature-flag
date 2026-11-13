---
sidebar_position: 20
description: A progressive rollout refers to an automated gradual release of a new feature.
---

# 📈 Progressive rollout

## Overview
A **progressive rollout** refers to a gradual release of a new feature to a subset of users. Instead of making the feature available to everyone at once, it's rolled out incrementally automatically, often starting with a small percentage and gradually increasing it over time.

## Define a progressive rollout
In GO Feature Flag, you can define a progressive rollout by setting the condition of the release ramp for the progressive rollout.

The condition to switch from one variation to another are based on dates and will roll out from `0%` to `100%` within the specified time frame.

To achieve this you will start by configuring the initial state of the flag and the end state of the flag.
- Before the `initial` date, the flag will return the initial variation
- Between the `initial` and `end` date, the flag will gradually shift from one variation to the other.
- After the `end` date, the flag will return the end state variation.

## Format
:::info
The dates are in the format supported natively by your flag file format.
:::

| Field                             | Description                                                                                                                                          |
|-----------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| **`initial`**                     | The initial state of this flag.<br/>`variation` is the variation you intend to switch from.<br/>`date` is the date to start the rollout.<br/>        |
| **`end`**                         | The end state of this flag.<br/>`variation` is the variation you intend to switch to.<br/>`date` is the date when rollout is completed.<br/>         |
| **`percentage`**<br/>*(optional)* | It represents the ramp of progress, at which level the flag starts (`initial`) and ends (`end`).<br/>**Default: `initial` = `0` and `end` = `100`**. |

## Example
Between the `2024-01-01T05:00:00.100Z` and `2024-01-05T05:00:00.100Z`, the flag will gradually shift from `variationA` to `variationB`.

```yaml
progressive-flag:
  variations:
    variationA: A
    variationB: B
  defaultRule:
# highlight-start
    progressiveRollout:
      initial:
        variation: variationA
        date: 2024-01-01T05:00:00.100Z
      end:
        variation: variationB
        date: 2024-01-05T05:00:00.100Z
# highlight-end
```

## Advanced: Using the percentage field
If you intend to partially rollout or keep both features, you can use the `percentage` field to define the ramp of progress.

```yaml
    progressiveRollout:
      initial:
        variation: variationA
        # highlight-start
        percentage: 20   
        # highlight-end
        date: 2024-01-01T05:00:00.100Z
      end:
        variation: variationB
        # highlight-start
        percentage: 80
        # highlight-end
        date: 2024-01-05T05:00:00.100Z
```

- Before the `initial` date, the flag will return **variationA for all users**.
- At the `initial` date, the configured split starts at **20% variationB** and **80% variationA**.
- Between the `initial` and `end` dates, the share of users receiving variationB gradually increases from **20% to 80%**.
- At and after the `end` date, the configured split remains at **80% variationB** and **20% variationA**.
