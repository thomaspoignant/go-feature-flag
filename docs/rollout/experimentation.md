# Experimentation rollout / A/B Testing
An **experimentation rollout** is when your flag is configured to be served only for a determined time.

1. It means that before the rollout start date, the `default` value is served to all users.
2. Between the dates the flag is evaluated.
3. After the end date the `default` value is served to all users.

## Example

=== "YAML"

    ``` yaml linenums="1" hl_lines="6-10"
    experimentation-flag:
      percentage: 50
      true: "B"
      false: "A"
      default: "Default"
      rollout:
        experimentation:
          start: 2021-03-20T00:00:00.10-05:00
          end: 2021-03-21T00:00:00.10-05:00
    ```

=== "JSON"

    ``` json linenums="1" hl_lines="7-12"
    {
      "experimentation-flag": {
        "percentage": 50,
        "true": "B",
        "false": "A",
        "default": "Default",
        "rollout": {
          "experimentation": {
            "start": "2021-03-20 00:00:00 -0500",
            "end": "2021-03-21 00:00:00 -0500"
          }
        }
      }
    }
    ```

=== "TOML"

    ``` toml linenums="1" hl_lines="7-11" 
    [experimentation-flag]
    percentage = 50.0
    true = "B"
    false = "A"
    default = "Default"
    
    [experimentation-flag.rollout]
    
        [experimentation-flag.rollout.experimentation]
        start = 2021-03-20T05:00:00.100Z
        end = 2021-03-21T05:00:00.100Z
    ```
 
Check this [example](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples/rollout_scheduled) to see how it works. 

## Configuration fields

!!! Info
    The dates are in the format supported natively by your flag file format.

| Field | Description |
|---|---|
|**`start`**| The date the flag will be started to be served.|
|**`end`**| The date the flag will be stopped to be served.|

## A/B testing

!!! Info
    A/B test is the shorthand for a simple controlled experiment.  
    As the name implies, two versions (A and B) of a single variable are compared, which are identical except for one variation that might affect a user's behavior.  
    A/B tests are widely considered the simplest form of controlled experiment.
    
    _**(source wikipedia)**_

To have a proper A/B testing solution with the module you should use the experimentation rollout combined with the [export of your data](../../data_collection/).

This combination will allow to have your experimentation running for a dedicated time, and you will have the data to knows exactly which user was on which version of the flag.

To setup the duration of your A/B test you can use a tool [ab-test-duration-calculator](https://vwo.com/tools/ab-test-duration-calculator/) from vwo, that will help you to set up the test duration correctly.
