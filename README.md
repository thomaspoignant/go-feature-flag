
<p align="center">
  <img width="250" height="238" src="logo.png" alt="go-feature-flag logo" />
</p>

# 🎛️ go-feature-flag [![Tweet](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/intent/tweet?text=I%27ve%20discovered%20go-feature-flag%20a%20great%20solution%20to%20easily%20managed%20feature%20flag%20in%20golang&url=https%3A%2F%2Fgithub.com%2Fthomaspoignant%2Fgo-feature-flag&via=thomaspoignant&hashtags=golang,featureflags,featuretoggle,go)

<p align="center">
    <a href="https://github.com/thomaspoignant/go-feature-flag/actions/workflows/ci.yml"><img src="https://github.com/thomaspoignant/go-feature-flag/actions/workflows/ci.yml/badge.svg" alt="Build Status" /></a>
    <a href="https://coveralls.io/github/thomaspoignant/go-feature-flag"><img src="https://coveralls.io/repos/github/thomaspoignant/go-feature-flag/badge.svg" alt="Coverage Status" /></a>
    <a href="https://sonarcloud.io/dashboard?id=thomaspoignant_go-feature-flag"><img src="https://sonarcloud.io/api/project_badges/measure?project=thomaspoignant_go-feature-flag&metric=alert_status" alt="Sonarcloud Status" /></a>
    <a href="https://github.com/thomaspoignant/go-feature-flag/actions/workflows/codeql-analysis.yml"><img src="https://github.com/thomaspoignant/go-feature-flag/workflows/CodeQL/badge.svg" alt="Build Status" /></a>
    <a href="https://app.fossa.com/projects/git%2Bgithub.com%2Fthomaspoignant%2Fgo-feature-flag"><img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2Fthomaspoignant%2Fgo-feature-flag.svg?type=shield" alt="FOSSA Status"/></a>
    <br/>
    <a href="https://github.com/thomaspoignant/go-feature-flag/releases"><img src="https://img.shields.io/github/v/release/thomaspoignant/go-feature-flag" alt="Release version" /></a>
    <a href="https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag"><img src="https://godoc.org/github.com/thomaspoignant/go-feature-flag?status.svg" alt="GoDoc" /></a>
    <img src="https://img.shields.io/github/go-mod/go-version/thomaspoignant/go-feature-flag?logo=go%20version" alt="Go version"/>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/thomaspoignant/go-feature-flag" alt="License"/></a>
    <a href="https://github.com/avelino/awesome-go/#server-applications"><img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go"></a>
</p>

**Feature flags with no complex system to maintain!**

## Installation
```bash
go get github.com/thomaspoignant/go-feature-flag
```
## What is go-feature-flag?

A simple and complete feature flag solution, without any complex backend system to install, you need only a file as your backend.

No server is needed, just add a file in a central system and all your services will react to the changes of this file.

**go-feature-flags supports:**

- Storing your configuration flags file on various locations (`HTTP`, `S3`, `GitHub`, `file`).
- Configuring your flags in various format (`JSON`, `TOML` and `YAML`).
- Adding complex rules to target your users.
- Use complex rollout strategy for your flags
    - Run A/B testing experimentation.
    - Progressively rollout a feature.
    - Schedule your flag updates.
- Exporting your flags usage data (`S3`, `log` and `file`).
- Getting notified when a flag has changed (`webhook` and `slack`).

If you are not familiar with feature flags also called feature Toggles you can read this [article of Martin Fowler](https://www.martinfowler.com/articles/feature-toggles.html)
that explains why this is a great pattern.

I've also written an [article](https://medium.com/better-programming/feature-flags-and-how-to-iterate-quickly-7e3371b9986) that explains why feature flags can help you to iterate quickly.

## Getting started
First, you need to initialize the `ffclient` with the location of your backend file.
```go
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &ffclient.HTTPRetriever{
        URL:    "http://example.com/flag-config.yaml",
    },
})
defer ffclient.Close()
```
*This example will load a file from an HTTP endpoint and will refresh the flags every 3 seconds (if you omit the
PollingInterval, the default value is 60 seconds).*

Now you can evaluate your flags anywhere in your code.

```go
user := ffuser.NewUser("user-unique-key")
hasFlag, _ := ffclient.BoolVariation("test-flag", user, false)
if hasFlag {
    // flag "test-flag" is true for the user
} else {
    // flag "test-flag" is false for the user
}
```
The full documentation is available on https://thomaspoignant.github.io/go-feature-flag/  
You can find more examples programs in the [examples/](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples) directory.

## Configuration

`go-feature-flag` needs to be initialized to be used.  
During the initialization you must give a [`ffclient.Config{}`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Config) configuration object.

[`ffclient.Config{}`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Config) is the only location where you can put the configuration.

### Example
```go
ffclient.Init(ffclient.Config{ 
    PollingInterval:   3 * time.Second,
    Logger:         log.New(file, "/tmp/log", 0),
    Context:        context.Background(),
    Retriever:      &ffclient.FileRetriever{Path: "testdata/flag-config.yaml"},
    FileFormat:     "yaml",
    Notifiers: []ffclient.NotifierConfig{
        &ffclient.WebhookConfig{
            EndpointURL: " https://example.com/hook",
            Secret:     "Secret",
            Meta: map[string]string{
                "app.name": "my app",
            },
        },
    },
    DataExporter: ffclient.DataExporter{
        FlushInterval:   10 * time.Second,
        MaxEventInMemory: 1000,
        Exporter: &ffexporter.File{
            OutputDir: "/output-data/",
        },
    },
    StartWithRetrieverError: false,
})
```
### Configuration fields

| Field | Description |
|---|---|
|`Retriever`  | The configuration retriever you want to use to get your flag file<br> *see [Store your flag file](https://thomaspoignant.github.io/go-feature-flag/flag_file/) for the configuration details*.|
|`Context`  | *(optional)*<br>The context used by the retriever.<br />Default: `context.Background()`|
|`DataExporter` | *(optional)*<br>DataExporter defines how to export data on how your flags are used.<br> *see [export data section](https://thomaspoignant.github.io/go-feature-flag/data_collection/) for more details*.|
|`FileFormat`| *(optional)*<br>Format of your configuration file. Available formats are `yaml`, `toml` and `json`, if you omit the field it will try to unmarshal the file as a `yaml` file.<br>Default: `YAML`|
|`Logger`   | *(optional)*<br>Logger used to log what `go-feature-flag` is doing.<br />If no logger is provided the module will not log anything.<br>Default: No log|
|`Notifiers` | *(optional)*<br>List of notifiers to call when your flag file has changed.<br> *see [notifiers section](https://thomaspoignant.github.io/go-feature-flag/notifier/) for more details*.|
|`PollingInterval`   | *(optional)* Duration to wait before refreshing the flags.<br>The minimum polling interval is 1 second<br>Default: 60 * time.Second|
|`StartWithRetrieverError` | *(optional)*<br>If **true**, the SDK will start even if we did not get any flags from the retriever. It will serve only default values until the retriever returns the flags.<br>The init method will not return any error if the flag file is unreachable.<br>Default: **false**|

### Multiple configuration flag files
`go-feature-flag` comes ready to use out of the box by calling the `Init` function and, it will be available everywhere.  
Since most applications will want to use a single central flag configuration, the package provides this. It is similar to a singleton.

In all the examples above, they demonstrate using `go-feature-flag` in its singleton style approach.  
You can also create many `go-feature-flag` clients to use in your application.  
[See the documentation for more details.](https://thomaspoignant.github.io/go-feature-flag/configuration/#multiple-configuration-flag-files)


## Where do I store my flags file?
The module supports different ways of retrieving the flag file.  
Available retriever are:

- [From GitHub](https://thomaspoignant.github.io/go-feature-flag/flag_file/github/)
- [From an HTTP endpoint](https://thomaspoignant.github.io/go-feature-flag/flag_file/http/)
- [From a S3 Bucket](https://thomaspoignant.github.io/go-feature-flag/flag_file/s3/)
- [From a file](https://thomaspoignant.github.io/go-feature-flag/flag_file/file/)

## Flags file format
`go-feature-flag` is to avoid to have to host a backend to manage your feature flags and to keep them centralized by using a file a source.  
Your file should be a `YAML`, `JSON` or `TOML` file with a list of flags *(examples: [`YAML`](testdata/flag-config.yaml), [`JSON`](testdata/flag-config.json), [`TOML`](testdata/flag-config.toml))*.

**A flag configuration looks like:**

<details open>
<summary>YAML</summary>

```yaml
test-flag:
  percentage: 100
  rule: key eq "random-key"
  true: true
  false: false
  default: false
  disable: false
  trackEvents: true
  version: 1
  rollout:
    experimentation:
      start: 2021-03-20T00:00:00.10-05:00
      end: 2021-03-21T00:00:00.10-05:00

  test-flag2:
    rule: key eq "not-a-key"
    percentage: 100
    true: true
    false: false
    default: false
    version: 12
```
</details>
<details>
<summary>JSON</summary>

```json
{
  "test-flag": {
    "percentage": 100,
    "rule": "key eq \"random-key\"",
    "true": true,
    "false": false,
    "default": false,
    "disable": false,
    "trackEvents": true,
    "version": 1,
    "rollout": {
      "experimentation": {
        "start": "2021-03-20T05:00:00.100Z",
        "end": "2021-03-21T05:00:00.100Z"
      }
    }
  },
  "test-flag2": {
    "rule": "key eq \"not-a-key\"",
    "percentage": 100,
    "true": true,
    "false": false,
    "default": false,
    "version": 12
  }
}
```

</details>

<details>
<summary>TOML</summary>

```toml
[test-flag]
percentage = 100.0
rule = "key eq \"random-key\""
true = true
false = false
default = false
disable = false
trackEvents = true
version = 1.0

[test-flag.rollout]

  [test-flag.rollout.experimentation]
  start = 2021-03-20T05:00:00.100Z
  end = 2021-03-21T05:00:00.100Z

[test-flag2]
rule = "key eq \"not-a-key\""
percentage = 100.0
true = true
false = false
default = false
version = 12.0
```

</details>


| Field | Description |
|:---:|---|
| **flag-key** | The `flag-key` is the name of your flag.<br> It must be unique.<br>*On the example the flag keys are **`test-flag`** and **`test-flag2`**.*|
| `true` | The value return by the flag if apply to the user *(rule is evaluated to true)* and user is in the active percentage.|
| `false`| The value return by the flag if apply to the user *(rule is evaluated to true)* and user is **not** in the active percentage.|
| `default` |The value return by the flag if not apply to the user *(rule is evaluated to false).*|
| `percentage` |*(optional)*<br>Percentage of the users affect by the flag.<br>**Default: 0**<br><br>The percentage is compute by doing a hash of the user key *(100000 variations)*, it means that you can have 3 numbers after the comma.|
| `rule` |*(optional)*<br>This is the query use to select on which user the flag should apply.<br>Rule format is describe in the <a href="#rule-format">rule format section</a>.<br>**If no rule set, the flag apply to all users *(percentage still apply)*.**|
| `disable` |*(optional)*<br>True if the flag is disabled.<br>**Default: `false`**|
| `trackEvents` |*(optional)*<br>False if you don't want to export the data in your data exporter.<br>**Default: `true`**|
| `version` |*(optional)*<br>The version is the version of your flag.<br>This number is used to display the information in the notifiers and data collection, you have to update it your self.<br>**Default: 0**|
| `rollout` |*(optional)*<br><code>rollout</code> contains a specific rollout strategy you want to use.<br>**See [rollout section](https://thomaspoignant.github.io/go-feature-flag/rollout/) for more details.**|

## Rule format
The rule format is based on the [`nikunjy/rules`](https://github.com/nikunjy/rules) library.

All the operations can be written capitalized or lowercase (ex: `eq` or `EQ` can be used).  
Logical Operations supported are `AND` `OR`.

Compare Expression and their definitions (`a|b` means you can use either one of the two `a` or `b`):

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
- Select all identified users: `anonymous ne true`
- Select a user with a custom property: `userId eq "12345"`

## Users
Feature flag targeting and rollouts are all determined by the user you pass to your Variation calls.
The SDK defines a [`User`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/ffuser#User) struct and a [`UserBuilder`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/ffuser#UserBuilder) to make this easy.

Here's an example:

```go
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

ℹ️ Custom attributes are one of the most powerful features. They let you have rules on these attributes and target users according to any data that you want.

You can also distinguish logged-in users from anonymous users in the SDK ([check documentation about anonymous users](https://thomaspoignant.github.io/go-feature-flag/users/#anonymous-users)).

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

The `MarshalJSON()` function will return a JSON Object, that can be directly used by your front-end application.  
[More details in the documentation.](https://thomaspoignant.github.io/go-feature-flag/users/#get-all-flags-for-a-specific-user)

## Rollout
A critical part of every new feature release is orchestrating the actual launch schedule between Product, Engineering, and Marketing teams.

Delivering powerful user experiences typically requires software teams to manage complex releases and make manual updates at inconvenient times.

But it doesn’t have to, having a complex **rollout** strategy allows you to have lifecycle for your flags.

### Complex rollout strategy available

- [Progressive rollout](https://thomaspoignant.github.io/go-feature-flag/rollout/progressive/) - increase the percentage of your flag over time.
- [Scheduled rollout](https://thomaspoignant.github.io/go-feature-flag/rollout/scheduled/) - update your flag over time.
- [Experimentation rollout](https://thomaspoignant.github.io/go-feature-flag/rollout/experimentation/) - serve your feature only for a determined time *(perfect for A/B testing)*.

## Notifiers
If you want to be informed when a flag has changed, you can configure a [**notifier**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#NotifierConfig).

A notifier will send one notification to the targeted system to inform them that a new flag configuration has been loaded.

ℹ️ `go-feature-flag` can handle more than one notifier at a time.

Available notifiers are:

- [Slack](https://thomaspoignant.github.io/go-feature-flag/notifiers/slack/) - Get a slack message with the changes.
- [Webhook](https://thomaspoignant.github.io/go-feature-flag/notifiers/webhook/) - Call an API with the changes.

## Export data
If you want to export data about how your flag are used, you can use the **`DataExporter`**.  
It collects all the variations events and can save these events on several locations:

- [File](https://thomaspoignant.github.io/go-feature-flag/data_collection/file/) *- create local files with the variation usages.*
- [Log](https://thomaspoignant.github.io/go-feature-flag/data_collection/log/) *- use your logger to write the variation usages.*
- [S3](https://thomaspoignant.github.io/go-feature-flag/data_collection/s3/) *- export your variation usages to S3.*
- [Webhook](https://thomaspoignant.github.io/go-feature-flag/data_collection/webhook/) *- export your variation usages by calling a webhook.*

Currently, we are supporting only feature events.  
It represents individual flag evaluations and are considered "full fidelity" events.

**An example feature event below:**
```json
{
    "kind": "feature",
    "contextKind": "anonymousUser",
    "userKey": "ABCD",
    "creationDate": 1618228297,
    "key": "test-flag",
    "variation": "Default",
    "value": false,
    "default": false
}
```
The format of the data is [described in the documentation](https://thomaspoignant.github.io/go-feature-flag/data_collection/#data-format).

Events are collected and send in bulk to avoid spamming your exporter *(see details in [how to configure data export](#how-to-configure-data-export)*).

### How to configure data export?
In your `ffclient.Config` add the `DataExporter` field and configure your export location.

To avoid spamming your location everytime you have a variation called, `go-feature-flag` is storing in memory all the events and send them in bulk to the exporter.  
You can decide the threshold on when to send the data with the properties `FlushInterval` and `MaxEventInMemory`. The first threshold hit will export the data.

If there are some flags you don't want to export, you can use `trackEvents` fields on these specific flags to disable the data export *(see [flag file format](https://thomaspoignant.github.io/go-feature-flag/flag_format/))*.

### Example
```go  linenums="1"
ffclient.Config{ 
    // ...
   DataExporter: ffclient.DataExporter{
        FlushInterval:   10 * time.Second,
        MaxEventInMemory: 1000,
        Exporter: &ffexporter.File{
            OutputDir: "/output-data/",
        },
    },
    // ...
}
```
The full configuration is [described in the documentation](https://thomaspoignant.github.io/go-feature-flag/data_collection/#how-to-configure-data-export).

# How can I contribute?
This project is open for contribution, see the [contributor's guide](CONTRIBUTING.md) for some helpful tips.
