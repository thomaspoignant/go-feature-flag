---
sidebar_position: 4
---

# Progressive rollout

A **progressive rollout** allows you to increase the percentage of your flag over time.

You can select a **release ramp** where the percentage of your flag will increase progressively between the start date
and the end date.

## Example

=== "YAML"

``` yaml linenums="1" hl_lines="6-14"
progressive-flag:
  variations:
    variationA: A
    variationB: B
  defaultRule:
    progressiveRollout:
      initial:
        variation: variationB
        percentage: 0
        date: 2021-03-20T00:00:00.1-05:00
      end:
        variation: variationB
        percentage: 100
        date: 2021-03-21T00:00:00.1-05:00
```

=== "JSON"

``` json linenums="1" hl_lines="8-18"
{
  "progressive-flag": {
    "variations": {
      "variationA": "A",
      "variationB": "B"
    },
    "defaultRule": {
      "progressiveRollout": {
        "initial": {
          "variation": "variationB",
          "percentage": 0,
          "date": "2021-03-20T05:00:00.100Z"
        },
        "end": {
          "variation": "variationB",
          "percentage": 100,
          "date": "2021-03-21T05:00:00.100Z"
        }
      }
    }
  }
}
```

=== "TOML"

``` toml linenums="1" hl_lines="5-13"
[progressive-flag.variations]
variationA = "A"
variationB = "B"

[progressive-flag.defaultRule.progressiveRollout.initial]
variation = "variationB"
percentage = 0
date = 2021-03-20T05:00:00.100Z

[progressive-flag.defaultRule.progressiveRollout.end]
variation = "variationB"
percentage = 100
date = 2021-03-21T05:00:00.100Z
```

## Configuration fields

!!! Info
    The dates are in the format supported natively by your flag file format.

| Field             | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **`releaseRamp`** | It contains the time slot where we will progressively increase the percentage of the flag.<ul><li>**Before** the `start` date we will serve the `percentage.initial` percentage of the flag.</li><li>**Between** `start` and `end` we will serve a percentage of the flag corresponding of the actual time.</li><li>**After** the `end` date we will serve the `percentage.end` percentage of the flag.</li></ul><p>If you have no date in your `releaseRamp` we will not do any progressive rollout and use the top level percentage you have configured *(0% in our example)*.</p> |
| **`percentage`**  | *(optional)*<br/>It represents the ramp of progress, at which level the flag starts (`initial`) and at which level it ends (`end`).<br/>**Default: `initial` = `0` and `end` = `100`**                                                                                                                                                                                                                                                                                                                                                                                               |
