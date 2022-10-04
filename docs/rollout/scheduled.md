---
sidebar_position: 5
---

# Scheduled rollout

Scheduling introduces the ability for users to changes their flags for future points in time.  
While this sounds deceptively straightforward, it unlocks the potential for users to create complex release strategies by scheduling the incremental steps in advance.

For example, you may want to turn a feature ON for internal testing tomorrow and then enable it for your ‘beta’ user segment four days later.

## Example

=== "YAML"

```yaml linenums="1" hl_lines="10-22"
scheduled-flag:
  variations:
    variationA: A
    variationB: B
  defaultRule:
    name: legacyDefaultRule
    percentage:
      variationA: 100
      variationB: 0
  scheduledRollout:
    - date: 2020-04-10T00:00:00.1+02:00
      targeting:
        - name: legacyRuleV0
          query: beta eq "true"
          percentage:
            variationA: 0
            variationB: 100
      
    - date: 2022-05-12T15:36:00.1+02:00
      targeting:
        - name: legacyRuleV0
          query: beta eq "false"
```

=== "JSON"

```json linenums="1" hl_lines="14-39"
{
  "scheduled-flag": {
    "variations": {
      "variationA": "A",
      "variationB": "B"
    },
    "defaultRule": {
      "name": "legacyDefaultRule",
      "percentage": {
        "variationA": 100,
        "variationB": 0
      }
    },
    "scheduledRollout": [
      {
        "date": "2020-04-09T22:00:00.100Z",
        "targeting": [
          {
            "name": "legacyRuleV0",
            "query": "beta eq \"true\"",
            "percentage": {
              "variationA": 0,
              "variationB": 100
            }
          }
        ]
      },
      {
        "date": "2022-05-12T13:36:00.100Z",
        "targeting": [
          {
            "name": "legacyRuleV0",
            "query": "beta eq \"false\""
          }
        ]
      }
    ]
  }
}
```

=== "TOML"

```toml linenums="1" hl_lines="12-28"
[scheduled-flag.variations]
variationA = "A"
variationB = "B"

[scheduled-flag.defaultRule]
name = "legacyDefaultRule"

  [scheduled-flag.defaultRule.percentage]
  variationA = 100
  variationB = 0

[[scheduled-flag.scheduledRollout]]
date = 2020-04-09T22:00:00.100Z

  [[scheduled-flag.scheduledRollout.targeting]]
  name = "legacyRuleV0"
  query = 'beta eq "true"'

    [scheduled-flag.scheduledRollout.targeting.percentage]
    variationA = 0
    variationB = 100

[[scheduled-flag.scheduledRollout]]
date = 2022-05-12T13:36:00.100Z

  [[scheduled-flag.scheduledRollout.targeting]]
  name = "legacyRuleV0"
  query = 'beta eq "false"'
```

## Configuration fields

!!! Info
    You can change any fields that are available on your flag.  
    Since your configuration has not been changed manually, it does not trigger any notifier.

| Field       | Description                                                                                                                                                                                                                                                                                                                                  |
|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **`steps`** | The only mandatory field in a **step** is the `date`.<br/>**If no date is provided the step will be skipped.**<br/><br/>The other attributes of your `step` are what you want to update your flag, so every field available in the [flag format](../../flag_format) can be updated.<br/>The new value in a field will override the existing one. |
