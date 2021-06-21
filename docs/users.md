# Targeting users with flags

## Users
Feature flag targeting and rollouts are all determined by the user you pass to your Variation calls.
The SDK defines a [`User`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/ffuser#User) struct and a [`UserBuilder`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/ffuser#UserBuilder) to make this easy.

Here's an example:

```go linenums="1"
// User with only a key
user1 := ffuser.NewUser("user1-key")

// User with a key plus other attributes
user2 = ffuser.NewUserBuilder("user2-key").
 AddCustom("firstname", "John").
 AddCustom("lastname", "Doe").
 AddCustom("email", "john.doe@example.com").
 Build()
```

The most common attribute is the user's key and **this is the only mandatory user attribute.** 
The key should also uniquely identify each user. You can use a primary key, an e-mail address, or a hash, as long as the same user always has the same key.  
**We recommend using a hash if possible.**    
All the other attributes are optional.

!!! info
    Custom attributes are one of the most powerful features.  
    They let you have rules on these attributes and target users according to any data that you want.

## Anonymous users
You can also distinguish logged-in users from anonymous users in the SDK, as follows:

```go linenums="1"
// User with only a key
user1 := ffuser.NewAnonymousUser("user1-key")

// User with a key plus other attributes
user2 = ffuser.NewUserBuilder("user2-key").
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

```go linenums="1"
result, _ := ffclient.BoolVariation("your.feature.key", user, false)

// result is now true or false depending on the setting of
// this boolean feature flag
```
Variation methods take the feature **flag key**, a **user**, and a **default value**.

The default value is return when an error is encountered _(`ffclient` not initialized, variation with wrong type, flag does not exist ...)._

In the example, if the flag `your.feature.key` does not exists, result will be `false`.  
Not that you will always have a usable value in the result. 

## Get all flags for a specific user
If you want to send the information about a specific user to a front-end, you will want a snapshot of all the flags for
this user at a specific time.

The method `ffclient.AllFlagsState` returns a snapshot of flag values and metadata.  
The function is evaluating all available flags for the user and return a `flagstate.AllFlagsState` object containing the
information you need.

```go linenums="1"
user := ffuser.NewUser("example")
// AllFlagsState will give you the value for all the flags available.
allFlagsState := ffclient.AllFlagsState(u)

// If you want to send it to a front-end you can Marshal it by calling MarshalJSON()
forFE, err := allFlagsState.MarshalJSON()
```

The `MarshalJSON()` function will return something like bellow, that can be directly used by your front-end application. 
```json linenums="1"
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

!!! Warning
    There is no tracking done when evaluating all the flag at once.
