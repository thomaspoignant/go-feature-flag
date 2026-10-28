---
sidebar_position: 30
description: How to select who should have the flag activated.
---
# 🎯 Performing flag evaluations

## Users
Feature flag targeting and rollouts are all determined by the user you pass to your **Variation** calls.
The SDK defines a [`User`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/ffuser#User) struct and a [`UserBuilder`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/ffuser#UserBuilder) to make this easy.

Here's an example:

```go showLineNumbers
// User with only a key
user1 := ffcontext.NewEvaluationContext("user1-key")

// User with a key plus other attributes
user2 = ffcontext.NewEvaluationContextBuilder("user2-key").
 AddCustom("firstname", "John").
 AddCustom("lastname", "Doe").
 AddCustom("email", "john.doe@example.com").
 Build()
```

The most common attribute is the user's key and **this is the only mandatory user attribute.**  
The key should also uniquely identify each user. You can use a primary key, an e-mail address, or a hash, as long as the same user always has the same key.  
**We recommend using a hash if possible.**  
All the other attributes are optional.

:::info
Custom attributes are one of the most powerful features.  
They let you have rules on these attributes and target users according to any data that you want.
:::

## Anonymous users
You can also distinguish logged-in users from anonymous users in the SDK, as follows:

```go showLineNumbers
// User with only a key
user1 := ffcontext.NewAnonymousEvaluationContext("user1-key")

// User with a key plus other attributes
user2 = ffcontext.NewEvaluationContextBuilder("user2-key").
  Anonymous(true).
  AddCustom("firstname", "John").
  AddCustom("lastname", "Doe").
  AddCustom("email", "john.doe@example.com").
  Build()
```
You will still need to generate a unique key for anonymous users. Session IDs or UUIDs work best for this.

Anonymous users work just like regular users, this information just helps you to add a rule to target a specific population.

## Variation
The Variation methods determine whether a flag is enabled or not for a specific user.
There is a Variation method for each type:   
[`BoolVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#BoolVariation) , [`IntVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#IntVariation)
, [`Float64Variation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Float64Variation)
, [`StringVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#StringVariation)
, [`JSONArrayVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#JSONArrayVariation)
, [`JSONVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#JSONVariation)

```go showLineNumbers
result, _ := ffclient.BoolVariation("your.feature.key", user, false)

// result is now true or false depending on the setting of
// this boolean feature flag
```
Variation methods take the feature **flag key**, a **user**, and a **default value**.

The default value is returned when an error is encountered _(`ffclient` not initialized, variation with wrong type, flag does not exist ...)._

In the example, if the flag `your.feature.key` does not exist, result will be `false`.  
Not that you will always have a usable value in the result. 

## Variation details
If you want more information about your flag evaluation, you can use the variation details functions.
There is a Variation method for each type:   
[`BoolVariationDetails`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#BoolVariationDetails) 
, [`IntVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#IntVariationDetails)
, [`Float64VariationDetails`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Float64VariationDetails)
, [`StringVariationDetails`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#StringVariationDetails)
, [`JSONArrayVariationDetails`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#JSONArrayVariationDetails)
, [`JSONVariationDetails`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#JSONVariationDetails)

You can use these functions the same way as the other variation functions BUT it will return a generic object `model.VariationResult[<type>]` containing your result.  
This object will contain these fields:

| field           | type                    | description                                                                    |
|-----------------|-------------------------|--------------------------------------------------------------------------------|
| `TrackEvents`   | `bool`                  | `true` if this evaluation was tracked.                                         |
| `VariationType` | `string`                | The name of the variation used to get this value.                              |
| `Failed`        | `bool`                  | `true` if an error occurred during the evaluation.                             |
| `Version`       | `string`                | The **version** of the flag used to do the evaluation.                         |
| `Reason`        | `flag.ResolutionReason` | The reason used for this evaluation.                                           |
| `ErrorCode`     | `flag.ErrorCode`        | Error code in case we have an error.                                           |
| `ErrorDetails`  | `string`                | A string providing more detail about the error.                                |
| `Value`         | `<type T>`              | Value of the flag in the expected type.                                        |
| `Cacheable`     | `bool`                  | `true` if it can be cached (by user or for everyone depending on the reason).  |


### Reason
GO Feature Flag can furnish you with diverse reasons in the variation details, giving you insight into the evaluation process of your feature flag.
Here is the full list of reason available:

| Reason                  | description                                                                                                                                                                                           |
|-------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `TARGETING_MATCH`       | The resolved value was the result of a dynamic evaluation, such as a rule or specific user-targeting. _(ex: serve variation A if username is Thomas)_                                                 |
| `TARGETING_MATCH_SPLIT` | The resolved value was the result of a dynamic evaluation, that is serving a percentage. _(ex: serve variation A to 10% of users with the username Thomas)_                                           |
| `SPLIT`                 | The resolved value was the result of pseudorandom assignment. _(ex: serve variation A to 10% of all the users.)_                                                                                      |
| `DISABLED`              | Indicates that the feature flag is disabled                                                                                                                                                           |
| `DEFAULT`               | The resolved value was the result of the flag being disabled in the management system.                                                                                                                |
| `STATIC`                | Indicates that the feature flag evaluated to a static value, for example, the default value for the flag. _(Note: Typically means that no dynamic evaluation has been executed for the feature flag)_ |
| `UNKNOWN`               | Indicates that an unknown issue occurred during evaluation                                                                                                                                                 |
| `ERROR`                 | Indicates that an error occurred during evaluation *(Note: The `errorCode` field contains the details of this error)*                                                                                 |
| `OFFLINE`               | Indicates that GO Feature Flag is currently evaluating in offline mode.                                                                                                                               |


## Get all flags for a specific user
If you want to send the information about a specific user to the front-end, you will need a snapshot of all the flags of this user at a specific time.

The method `ffclient.AllFlagsState` returns a snapshot of flag values and metadata.  
The function is evaluating all available flags for the user and returns a `flagstate.AllFlagsState` object containing the information you need.

```go showLineNumbers
user := ffcontext.NewEvaluationContext("example")
// AllFlagsState will give you the value for all the flags available.
allFlagsState := ffclient.AllFlagsState(u)

// If you want to send it to the front-end you can Marshal it by calling MarshalJSON()
forFE, err := allFlagsState.MarshalJSON()
```

The `MarshalJSON()` function will return something like below, that can be directly used by your front-end application. 
```json showLineNumbers
{
    "flags": {
        "test-flag0": {
            "value": true,
            "timestamp": 1622209328,
            "variationType": "True",
            "trackEvents": true
        },
        "test-flag1": {
            "value": "true",
            "timestamp": 1622209328,
            "variationType": "True",
            "trackEvents": true
        },
        "test-flag2": {
            "value": 1,
            "timestamp": 1622209328,
            "variationType": "True",
            "trackEvents": true
        },
        "test-flag3": {
            "value": [
                "yo",
                "ya"
            ],
            "timestamp": 1622209328,
            "variationType": "True",
            "trackEvents": true
        },
        "test-flag4": {
            "value": {
                "test": "yo"
            },
            "timestamp": 1622209328,
            "variationType": "True",
            "trackEvents": true
        },
        "test-flag5": {
            "value": 1.1,
            "timestamp": 1622209328,
            "variationType": "True",
            "trackEvents": false
        }
    },
    "valid": true
}
```

:::caution
There is no tracking done when evaluating all the flag at once.
:::
