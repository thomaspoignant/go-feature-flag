# 🎛️ go-feature-flag
[![Release version](https://img.shields.io/github/v/release/thomaspoignant/go-feature-flag "version")](https://github.com/thomaspoignant/go-feature-flag/releases)
[![Build Status](https://travis-ci.com/thomaspoignant/go-feature-flag.svg?branch=main)](https://travis-ci.com/thomaspoignant/go-feature-flag)
[![Coverage Status](https://coveralls.io/repos/github/thomaspoignant/go-feature-flag/badge.svg)](https://coveralls.io/github/thomaspoignant/go-feature-flag)
[![Sonarcloud Status](https://sonarcloud.io/api/project_badges/measure?project=thomaspoignant_go-feature-flag&metric=alert_status)](https://sonarcloud.io/dashboard?id=thomaspoignant_go-feature-flag)
[![GitHub](https://img.shields.io/github/license/thomaspoignant/go-feature-flag)](LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fthomaspoignant%2Fgo-feature-flag.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fthomaspoignant%2Fgo-feature-flag?ref=badge_shield)
![Go version](https://img.shields.io/github/go-mod/go-version/thomaspoignant/go-feature-flag?logo=go%20version "Go version")

A feature flag solution, with YAML file in the backend (S3, HTTP, local file ...).  
No server to install, just add a file in a central system *(HTTP, S3, ...)* and all your services will react to the changes of this file.

## Installation
```bash
go get github.com/thomaspoignant/go-feature-flag
```

## Quickstart
First, you need to initialize the `ffclient` with the location of your backend file.
```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    HTTPRetriever: &ffClient.HTTPRetriever{
        URL:    "http://example.com/test.yaml",
    },
})
defer ffclient.Close()
```
*This example will load a file from an HTTP endpoint and will refresh the flags every 3 seconds.*

Now you can evalute your flags anywhere in your code.
```go
user := ffuser.NewUser("user-unique-key")
hasFlag, _ := ffclient.BoolVariation("test-flag", user, false)
if hasFlag {
    // flag "test-flag" is true for the user
} else {
    // flag "test-flag" is false for the user
}
```

## Flags file format
`go-feature-flag` is to avoid to have to host a backend to manage your feature flags and to keep them centralized by using a file a source.  
Your file should be a YAML file with a list of flags *([see example](testdata/test.yaml))*.

A flag configuration looks like:
```yaml
test-flag:
  percentage: 100
  rule: key eq "random-key"
  true: true
  false: false
  default: false
  disable: false
```

|   |   |   |
|---|---|---|
|`test-flag`   |![mandatory](https://img.shields.io/badge/-mandatory-red)   |  Name of the flag. It should be unique.  |
|`percentage`   |![optional](https://img.shields.io/badge/-optional-green)   |  Percentage of the users affect by the flag.<br>**Default value is 0**  |
|`rule`   |![optional](https://img.shields.io/badge/-optional-green)   |  This is the query use to select on which user the flag should apply.<br>Rule format is describe in the [rule format section](#rule-format).<br>**If no rule set, the flag apply to all users *(percentage still apply)*.** |
|`true`   |![mandatory](https://img.shields.io/badge/-mandatory-red)   |  The value return by the flag if apply to the user *(rule is evaluated to true)* and user is in the active percentage. |
|`false`   |![mandatory](https://img.shields.io/badge/-mandatory-red)   |  The value return by the flag if apply to the user *(rule is evaluated to true)* and user is **not** in the active percentage. |
|`disable`   |![optional](https://img.shields.io/badge/-optional-green)   |  True if the flag is disabled. |

## Rule format
The rule format is based on the [`nikunjy/rules`](https://github.com/nikunjy/rules) library.

All the operations can be written capitalized or lowercase (ex: eq or EQ can be used).  
Logical Operations supported are `AND` `OR`.

Compare Expression and their definitions (a|b means you can use either one of the two a or b):

```
eq|==: equals to 
ne|!=: not equals to
lt|<: less than 
gt|>: greater than
le|<=: less than equal to
ge|>=: greater than equal to 
co: contains 
sw: starts with 
ew: ends with
in: in a list
pr: present
not: not of a logical expression
```

### Examples

- Select a specific user: `key eq "example@example.com"`
- Select all anonymous user: `anonymous eq true`
- Select a user with a custom property: `userId eq "12345"`
