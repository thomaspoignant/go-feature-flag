# Scheduled rollout

Scheduling introduces the ability for users to changes their flags for future points in time.  
While this sounds deceptively straightforward, it unlocks the potential for users to create complex release strategies by scheduling the incremental steps in advance.

For example, you may want to turn a feature ON for internal testing tomorrow and then enable it for your ‘beta’ user segment four days later.

## Example
=== "YAML"

    ```yaml linenums="1" hl_lines="6-13"
    scheduled-flag:
      true: "B"
      false: "A"
      default: "Default"
      rollout:
        scheduled:
          steps:
            - date: 2020-04-10T00:00:00.10+02:00
              rule: beta eq "true"
              percentage: 100
            
            - date: 2022-05-12T15:36:00.10+02:00
              rule: beta eq "false"
    ```
=== "JSON"

    ```json linenums="1" hl_lines="6-19"
    {
      "scheduled-flag": {
        "true": "B",
        "false": "A",
        "default": "Default",
        "rollout": {
          "scheduled": {
            "steps": [
              {
                "date": "2020-04-09T22:00:00.100Z",
                "rule": "beta eq \"true\"",
                "percentage": 100
              },
              {
                "date": "2022-05-12T13:36:00.100Z",
                "rule": "beta eq \"false\""
              }
            ]
          }
        }
      }
    }
    ```

=== "TOML"

    ```toml linenums="1" hl_lines="6-17"
    [scheduled-flag]
    true = "B"
    false = "A"
    default = "Default"
    
      [scheduled-flag.rollout]
    
        [scheduled-flag.rollout.scheduled]
    
          [[scheduled-flag.rollout.scheduled.steps]]
          date = 2020-04-09T22:00:00.100Z
          rule = "beta eq \"true\""
          percentage = 100.0
    
          [[scheduled-flag.rollout.scheduled.steps]]
          date = 2022-05-12T13:36:00.100Z
          rule = "beta eq \"false\""
    ```

## Configuration fields

!!! Info
    You can change any fields that are available on your flag.  
    Since your configuration has not been changed manually, it does not trigger any notifier.

| Field | Description |
|---|---|
|**`steps`**| The only mandatory field in a **step** is the `date`.<br>**If no date is provided the step will be skipped.**<br><br>The other attributes of your `step` are what you want to update your flag, so every field available in the [flag format](../../flag_format) can be updated.<br>The new value in a field will override the existing one. |
