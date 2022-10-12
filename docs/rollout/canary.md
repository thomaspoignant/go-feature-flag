---
sidebar_position: 2
---

# Canary Release

**Canary release** is a technique to reduce the risk of introducing a new software version in production by slowly rolling out the change to a small subset of users before rolling it out to the entire infrastructure and making it available to everybody.

This is the easiest rollout strategy available.  
You just have to select a percentage of your users in your flag, and the `True` behavior will apply to them.

## Example

=== "YAML"

    ``` yaml linenums="1" hl_lines="8"
    canary-flag:
      variations:
        oldBehavior: false
        canary: true
      defaultRule:
        percentage:
          oldBehavior: 99
          canary: 1
    ```

=== "JSON"

    ``` json linenums="1" hl_lines="10"
    {
      "canary-flag": {
        "variations": {
          "oldBehavior": false,
          "canary": true
        },
        "defaultRule": {
          "percentage": {
            "oldBehavior": 99,
            "canary": 1
          }
        }
      }
    }
    ```

=== "TOML"

    ``` toml linenums="1" hl_lines="7"
    [canary-flag.variations]
    oldBehavior = false
    canary = true
    
    [canary-flag.defaultRule.percentage]
    oldBehavior = 99
    canary = 1

    ```
