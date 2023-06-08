<p align="center">
  <img width="250" height="238" src="logo.png" alt="go-feature-flag logo" />
</p>

# 🎛️ GO Feature Flag [![Tweet](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/intent/tweet?text=I%27ve%20discovered%20go-feature-flag%20a%20great%20solution%20to%20easily%20managed%20feature%20flag%20in%20golang&url=https%3A%2F%2Fgithub.com%2Fthomaspoignant%2Fgo-feature-flag&via=gofeatureflag&hashtags=golang,featureflags,featuretoggle,go)

<p align="center">
    <a href="https://github.com/thomaspoignant/go-feature-flag/actions/workflows/ci.yml"><img src="https://github.com/thomaspoignant/go-feature-flag/actions/workflows/ci.yml/badge.svg" alt="Build Status" /></a>
    <a href="https://codecov.io/gh/thomaspoignant/go-feature-flag"><img src="https://codecov.io/gh/thomaspoignant/go-feature-flag/branch/main/graph/badge.svg?token=T59ACQWW1H"/></a>
    <a href="https://sonarcloud.io/dashboard?id=thomaspoignant_go-feature-flag"><img src="https://sonarcloud.io/api/project_badges/measure?project=thomaspoignant_go-feature-flag&metric=alert_status" alt="Sonarcloud Status" /></a>
    <a href="https://github.com/thomaspoignant/go-feature-flag/actions/workflows/codeql-analysis.yml"><img src="https://github.com/thomaspoignant/go-feature-flag/workflows/CodeQL/badge.svg" alt="Build Status" /></a>
    <br/>
    <a href="https://github.com/thomaspoignant/go-feature-flag/releases"><img src="https://img.shields.io/github/v/release/thomaspoignant/go-feature-flag" alt="Release version" /></a>
    <a href="https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag"><img src="https://godoc.org/github.com/thomaspoignant/go-feature-flag?status.svg" alt="GoDoc" /></a>
    <img src="https://img.shields.io/github/go-mod/go-version/thomaspoignant/go-feature-flag?logo=go%20version" alt="Go version"/>
    <a href="https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE"><img src="https://img.shields.io/github/license/thomaspoignant/go-feature-flag" alt="License"/></a>
    <a href="https://github.com/avelino/awesome-go/#server-applications"><img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go"></a>
    <a href="https://gophers.slack.com/messages/go-feature-flag"><img src="https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=green" alt="Join us on slack"></a>
</p>

:pray: If you are using **GO Feature Flag** we've launched the [GO Feature Flag Usage Survey](https://forms.gle/cD4Rj7gJSMStg1Hc9), it would be awesome and super helpful if you can fill it.

## Table of Contents

- [What is GO Feature Flag?](#what-is-go-feature-flag)
- [What can I do with GO Feature Flag?](#what-can-i-do-with-go-feature-flag)
- [Getting started](#getting-started)
    - [Using the GO Module](#using-go-module)
    - [Using Open Feature SDKs](#using-open-feature)
- [Can I use GO Feature Flag with any language?](#can-i-use-go-feature-flag-with-any-language)
- [Where do I store my flags file?](#where-do-i-store-my-flags-file)
- [Flags file format](#flags-file-format)
- [Rule format](#rule-format)
- [Users](#users)
- [Variations](#variations)
- [Get all flags for a specific user](#get-all-flags-for-a-specific-user)
- [Rollout](#rollout)
- [Notifiers](#notifiers)
- [Export data](#export-data)
- [Linter](#linter)
- [How can I contribute?](#how-can-i-contribute?)
- [Contributors](#contributors)

## What is GO Feature Flag?

GO Feature Flag is a lightweight and open-source solution that provides a simple and complete feature flag implementation.

The solution has been built to facilitate the usage of feature flags in your code without having to contact any vendor.

Originally, **GO Feature Flag** was designed as a solution exclusively for the `GO` language. With the new standardization of feature flags by the [Openfeature](https://openfeature.dev/) project, the solution is now available for multiple languages _(`JAVA`, `typescript`, `javascript`, ...)_ through a simple API server called the relay proxy, which can be hosted.

> ℹ️ Info  
If you are not familiar with feature flags, I've written an [article](https://medium.com/better-programming/feature-flags-and-how-to-iterate-quickly-7e3371b9986) which explains why feature flags can fasten your iteration cycle.

## What can I do with GO Feature Flag?

- Storing your configuration flags file on various locations (`HTTP`, `S3`, `Kubernetes`, [_see full list_](https://gofeatureflag.org/docs/configure_flag/store_your_flags)).
- Configuring your flags in various [format](https://gofeatureflag.org/docs/configure_flag/flag_format) (`JSON`, `TOML` and `YAML`).
- Adding complex [rules](https://gofeatureflag.org/docs/configure_flag/flag_format#rule-format) to target your users.
- Use a complex rollout strategy for your flags :
    - [Run A/B testing experimentation](https://gofeatureflag.org/docs/configure_flag/rollout/experimentation).
    - [Progressively rollout a feature](https://gofeatureflag.org/docs/configure_flag/rollout/progressive).
    - [Schedule your flag updates](https://gofeatureflag.org/docs/configure_flag/rollout/scheduled).
- Exporting your flags usage data to various destinations such as _(`S3`, `Google cloud storage`, `file`, see the [_full list_](https://gofeatureflag.org/docs/configure_flag/export_flags_usage))_.
- Getting notified when a flag has been changed _(`webhook` and `slack`)_.
- Use **GO Feature Flag** in several languages with **Open Feature SDKs**.


https://user-images.githubusercontent.com/17908063/211581747-f6354a9d-8be6-4e52-aa53-7f0d6a40827e.mp4

_The code of this demo is available in [`examples/demo`](examples/demo) repository_.

## Getting started

Before starting to use **GO Feature Flag** you should decide if you want to use the GO Module directly or if you want to install the relay proxy.

The GO module is ideal for using GO Feature Flag exclusively in GO projects. If your project involves multiple languages, we recommend using the Open Feature SDKs.

<a id="using-go-module"></a>
<details>
<summary><b>Using the GO Module</b></summary>

### Installation
```bash
go get github.com/thomaspoignant/go-feature-flag
```

### Create a feature flag configuration

Create a new `YAML` file containing your first flag configuration.

```yaml title="flag-config.yaml"
# 20% of the users will use the variation "my-new-feature"
test-flag:
  variations:
    my-new-feature: true
    my-old-feature: false
  defaultRule:
    percentage:
      my-new-feature: 20
      my-old-feature: 80
```

This flag split the usage of this flag, 20% will use the variation `my-new-feature` and 80% the variation `my-old-feature`.

### SDK Initialisation
First, you need to initialize the `ffclient` with the location of your backend file.
```go linenums="1"
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever:      &fileretriever.Retriever{
        Path: "flag-config.yaml",
    },
})
defer ffclient.Close()
```
*This example will load a file from your local computer and will refresh the flags every 3 seconds (if you omit the
PollingInterval, the default value is 60 seconds).*

> ℹ info  
This is a basic configuration to test locally, in production it is better to use a remote place to store your feature flag configuration file.  
Look at the list of available options in the [**Store your feature flag file** page](https://gofeatureflag.org/docs/go_module/store_file/).

### Evaluate your flags
Now you can evaluate your flags anywhere in your code.

```go linenums="1"
user := ffuser.NewUser("user-unique-key")
hasFlag, _ := ffclient.BoolVariation("test-flag", user, false)
if hasFlag {
    // flag "test-flag" is true for the user
} else {
    // flag "test-flag" is false for the user
}
```
The full documentation is available on https://docs.gofeatureflag.org  
You can find more examples in the [examples/](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples) directory.

</details>


<a id="using-open-feature"></a>
<details>
<summary><b>Using Open Feature SDKs</b></summary>

### Create a feature flag configuration

Create a new `YAML` file containing your first flag configuration.

```yaml title="flag-config.yaml"
# 20% of the users will use the variation "my-new-feature"
test-flag:
  variations:
    my-new-feature: true
    my-old-feature: false
  defaultRule:
    percentage:
      my-new-feature: 20
      my-old-feature: 80
```

This flag split the usage of this flag, 20% will use the variation `my-new-feature` and 80% the variation `my-old-feature`.

### Create a relay proxy configuration file

Create a new `YAML` file containing the configuration of your relay proxy.

```yaml title="goff-proxy.yaml"
listen: 1031
pollingInterval: 1000
startWithRetrieverError: false
retriever:
  kind: file
  path: /goff/flag-config.yaml
exporter:
  kind: log
```

### Install the relay proxy

And we will run the **relay proxy** locally to make the API available.  
The default port will be `1031`.

```shell
# Launch the container
docker run \
  -p 1031:1031 \
  -v $(pwd)/flag-config.yaml:/goff/flag-config.yaml \
  -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml \
  thomaspoignant/go-feature-flag-relay-proxy:latest

```

_If you don't want to use docker to install the **relay proxy** you can go to [docker hub](https://hub.docker.com/r/thomaspoignant/go-feature-flag-relay-proxy)_.

### Use Open Feature SDK

_In this example, we are using the javascript SDK, but it is still relevant for all the languages_.

#### Install dependencies

```shell
npm i @openfeature/js-sdk @openfeature/go-feature-flag-provider
```

#### Init your Open Feature client

In your app initialization you have to create a client using the Open Feature SDK and initialize it.

```javascript
const {OpenFeature} = require("@openfeature/js-sdk");
const {GoFeatureFlagProvider} = require("@openfeature/go-feature-flag-provider");


// init Open Feature SDK with GO Feature Flag provider
const goFeatureFlagProvider = new GoFeatureFlagProvider({
  endpoint: 'http://localhost:1031/' // DNS of your instance of relay proxy
});
OpenFeature.setProvider(goFeatureFlagProvider);
const featureFlagClient = OpenFeature.getClient('my-app')
```

#### Evaluate your flag

Now you can evaluate your flags anywhere in your code using this client.

```javascript
// Context of your flag evaluation.
// With GO Feature Flag you MUST have a targetingKey that is a unique identifier of the user.
const userContext = {
  targetingKey: '1d1b9238-2591-4a47-94cf-d2bc080892f1', // user unique identifier (mandatory)
  firstname: 'john',
  lastname: 'doe',
  email: 'john.doe@gofeatureflag.org',
  admin: true, // this field is used in the targeting rule of the flag "flag-only-for-admin"
  // ...
};

const adminFlag = await featureFlagClient.getBooleanValue('flag-only-for-admin', false, userContext);
if (adminFlag) {
   // flag "flag-only-for-admin" is true for the user
  console.log("new feature");
} else {
  // flag "flag-only-for-admin" is false for the user
}
```

</details>

## Can I use GO Feature Flag with any language?

Originally GO Feature Flag was built to be a GOlang only library, but it limits the ecosystem too much.  
To be compatible with more languages we have implemented the [GO Feature Flag Relay Proxy](cmd/relayproxy/).
It is a service you can host that provides an API to evaluate your flags, you can call it using HTTP to get your variation.

Since we believe in standardization we are also implementing [OpenFeature](https://github.com/open-feature) providers to interact with this API in the language of your choice.  
_(OpenFeature is still at an early stage, so not all languages are supported and expect some changes in the future)_

For now, we have providers for `python`, `java`, `typescript`, `javascript`, `GO` and `.NET`.


## Where do I store my flags file?

The module supports different ways of retrieving the flag file.  
The available retrievers are:
- **GitHub**
- **HTTP endpoint**
- **AWS S3**
- **Local file**
- **Google Cloud Storage**
- **Kubernetes ConfigMaps**

_[See the full list and more information.](https://gofeatureflag.org/docs/configure_flag/store_your_flags)_

## Flags file format
**GO Feature Flag** core feature is to centralize all your feature flags in a single file and to avoid hosting and maintaining a backend server to manage them.

Your file should be a `YAML`, `JSON` or `TOML` file with a list of flags *(examples: [`YAML`](testdata/flag-config.yaml), [`JSON`](testdata/flag-config.json), [`TOML`](testdata/flag-config.toml))*.

The easiest way to create your configuration file is to use **GO Feature Flag Editor** available at https://editor.gofeatureflag.org.  
If you prefer to do it manually please follow the instruction bellow.

**A flag configuration looks like this:**

<details open>
<summary>YAML</summary>

```yaml
# This is your configuration for your first flag
first-flag:
  variations: # All possible return value for your feature flag
    A: false
    B: true
  targeting: # If you want to target a subset of your users in particular
    - query: key eq "random-key"
      percentage:
        A: 0
        B: 100
  defaultRule: # When no targeting match we use the defaultRule
    variation: A

# A second example of a flag configuration
second-flag:
  variations:
    A: "valueA"
    B: "valueB"
    defaultValue: "a default value"
  targeting:
    - name: notkey_rule
      query: key eq "not-a-key"
      percentage:
        A: 10
        B: 90
  defaultRule:
    variation: defaultValue
  version: "12"
  experimentation: 
    start: 2021-03-20T00:00:00.1-05:00
    end: 2021-03-21T00:00:00.1-05:00
```
</details>
<details>
<summary>JSON</summary>

```json
{
  "first-flag": {
    "variations": {
      "A": false,
      "B": true
    },
    "targeting": [
      {
        "query": "key eq \"random-key\"",
        "percentage": {
          "A": 0,
          "B": 100
        }
      }
    ],
    "defaultRule": {
      "variation": "A"
    }
  },
  
  "second-flag": {
    "variations": {
      "A": "valueA",
      "B": "valueB",
      "defaultValue": "a default value"
    },
    "targeting": [
      {
        "name": "notkey_rule",
        "query": "key eq \"not-a-key\"",
        "percentage": {
          "A": 10,
          "B": 90
        }
      }
    ],
    "defaultRule": {
      "variation": "defaultValue"
    },
    "version": "12",
    "experimentation": {
      "start": "2021-03-20T05:00:00.100Z",
      "end": "2021-03-21T05:00:00.100Z"
    }
  }
}
```

</details>

<details>
<summary>TOML</summary>

```toml
[first-flag.variations]
A = false
B = true

[[first-flag.targeting]]
query = 'key eq "random-key"'

  [first-flag.targeting.percentage]
  A = 0
  B = 100

[first-flag.defaultRule]
variation = "A"

[second-flag]
version = "12"

  [second-flag.variations]
  A = "valueA"
  B = "valueB"
  defaultValue = "a default value"

  [[second-flag.targeting]]
  name = "notkey_rule"
  query = 'key eq "not-a-key"'

    [second-flag.targeting.percentage]
    A = 10
    B = 90

  [second-flag.defaultRule]
  variation = "defaultValue"

  [second-flag.experimentation]
  start = 2021-03-20T05:00:00.100Z
  end = 2021-03-21T05:00:00.100Z
```

</details>

For detailed information on the fields required to create a flag, please refer to the [documentation](https://gofeatureflag.org/docs/configure_flag/flag_format). 

## Rule format

The query format is based on the [`nikunjy/rules`](https://github.com/nikunjy/rules) library.

All the operations can be written in capitalized or lowercase (ex: `eq` or `EQ` can be used).  
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
Feature flag targeting and rollouts are all determined by the user you pass to your evaluation calls.

The only required field for a user is his unique `key`, it is used by the internals of GO Feature Flag to do a hash to define
if the flag can apply to this user or not.
You can use a primary key, an e-mail address, or a hash, as long as the same user always has the same key.    
**We recommend using a hash if possible.**   
All the other attributes are optional.

Since it is useful to make complex queries on your flag, you can add as many information fields you want to your user.
It will be used when testing the targeting rules.

You can also distinguish logged-in users from anonymous users in the SDK ([check documentation about anonymous users](https://gofeatureflag.org/docs/go_module/target_user#anonymous-users)).

## Variations

The Variation methods determine whether a flag is enabled or not for a specific user.

GO Feature Flag can manage more than just `boolean` values; the value of your flag can be any of the following types:
- `bool`
- `int`
- `float`
- `string`
- `json array`
- `json object`

### Example
```go linenums="1"
result, _ := ffclient.BoolVariation("your.feature.key", user, false)

// result is now true or false depending on the setting of
// this boolean feature flag
```
Variation methods take the feature **flag key**, a **user**, and a **default value**.

The default value is returned when an error is encountered _(`ffclient` not initialized, variation with wrong type, flag does not exist ...)._

In the example, if the flag `your.feature.key` does not exist, the result will be `false`.  
Note that the result will always provide a usable value.

## Get all flags for a specific user

If you want to send the information about a specific user to a front-end, you will want a snapshot of all the flags for
this user at a specific time.

The method `ffclient.AllFlagsState` returns a snapshot of flag values and metadata.  
The function is evaluating all available flags for the user and returns a `flagstate.AllFlagsState` object containing the
information you need.

The `MarshalJSON()` function will return a JSON Object, that can be directly used by your front-end application.  
[More details in the documentation.](https://gofeatureflag.org/docs/go_module/target_user#get-all-flags-for-a-specific-user)

## Rollout

A critical part of every new feature release is orchestrating the actual launch schedule between the Product, Engineering, and Marketing teams.

Delivering powerful user experiences typically requires software teams to manage complex releases and make manual updates at inconvenient times.

But it does not have to, having a complex **rollout** strategy allows you to have a lifecycle for your flags.

### Complex rollout strategy available

- [Canary releases](https://gofeatureflag.org/docs/configure_flag/rollout/canary) - impact randomly a subset of your users.
- [Progressive rollout](https://gofeatureflag.org/docs/configure_flag/rollout/progressive) - increase the percentage of your flag over time.
- [Scheduled rollout](https://gofeatureflag.org/docs/configure_flag/rollout/scheduled/) - update your flag over time.
- [Experimentation rollout](https://gofeatureflag.org/docs/configure_flag/rollout/experimentation) - serve your feature only for a determined time *(perfect for A/B testing)*.

## Notifiers
If you want to be informed when a flag has changed, you can configure a [**notifier**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#NotifierConfig).

A notifier will send one notification to the targeted system to inform them that a new flag configuration has been loaded.

ℹ️ **GO Feature Flag** can handle more than one notifier at a time.

Available notifiers are:
- **Slack**
- **Webhook**

## Export data
**GO Feature Flag** allows you to export data about the usage of your flags.    
It collects all variation events and can save these events in several locations:

- **Local file** *- create local files with the variation usages.*
- **Log** *- use your logger to write the variation usages.*
- **AWS S3** *- export your variation usages to S3.*
- **Google Cloud Storage** *- export your variation usages to Google Cloud Storage.*
- **Webhook** *- export your variation usages by calling a webhook.*

Currently, we are supporting only feature events.  
It represents individual flag evaluations and is considered "full fidelity" events.

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
The format of the data is [described in the documentation](https://gofeatureflag.org/docs/).
Events are collected and sent in bulk to avoid spamming your exporter.

## Linter
A command line tool is available to help you lint your configuration file: [go-feature-flag-lint](cmd/lint/README.md).

# How can I contribute?

This project welcomes contributions from the community. If you're interested in contributing, see the [contributors' guide](CONTRIBUTING.md) for some helpful tips.

## Contributors

Thanks so much to our contributors.

<a href="https://github.com/thomaspoignant/go-feature-flag/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=thomaspoignant/go-feature-flag" />
</a>
