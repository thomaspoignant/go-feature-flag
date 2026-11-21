# Ruby

[![gem](https://img.shields.io/gem/v/openfeature-go-feature-flag-provider?color=blue\&style=flat-square\&logo=ruby)](https://github.com/open-feature/ruby-sdk-contrib/tree/main/providers/openfeature-go-feature-flag-provider)

This repository contains the official Ruby OpenFeature provider for accessing your feature flags with [GO Feature Flag](https://gofeatureflag.org).

In conjunction with the [OpenFeature SDK](https://openfeature.dev/docs/reference/concepts/provider) you will be able to evaluate your feature flags in your Ruby applications.

For documentation related to flags management in GO Feature Flag, refer to the [GO Feature Flag documentation website](https://gofeatureflag.org/docs).

### Functionalities:[​](#functionalities "Direct link to Functionalities:")

* Manage the integration of the OpenFeature Ruby SDK and GO Feature Flag relay-proxy.

## Dependency Setup[​](#dependency-setup "Direct link to Dependency Setup")

### Gem Package Manager[​](#gem-package-manager "Direct link to Gem Package Manager")

Add this line to your application's Gemfile:

```
gem 'openfeature-go-feature-flag-provider'
```

And then execute:

```
bundle install
```

Or install it yourself as:

```
gem install openfeature-go-feature-flag-provider
```

## Getting started[​](#getting-started "Direct link to Getting started")

### Initialize the provider[​](#initialize-the-provider "Direct link to Initialize the provider")

The `OpenFeature::GoFeatureFlag::Provider` needs some options to be created and then set in the OpenFeature SDK.

| **Option** | **Description**                                                                                                                    |
| ---------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `endpoint` | **(mandatory)** The URL to access to the relay-proxy.<br />*(example: `https://relay.proxy.gofeatureflag.org/`)*                   |
| `headers`  | A `Hash` object containing the headers to send to the relay-proxy.<br />\*(example to send APIKey: `{"X-API-Key" => "my-api-key"}` |

The only required option to create a `GoFeatureFlagProvider` is the URL *(`endpoint`)* to your GO Feature Flag relay-proxy instance.

```
require "openfeature/go_feature_flag"

require "openfeature/sdk"



# ...



options = OpenFeature::GoFeatureFlag::Options.new(endpoint: "http://localhost:1031")

provider = OpenFeature::GoFeatureFlag::Provider.new(options: options)



evaluation_context = OpenFeature::SDK::EvaluationContext.new(targeting_key: "9b9450f8-ab5c-4dcf-872f-feda3f6ccb16")



OpenFeature::SDK.configure do |config|

   config.set_provider(provider)

end

client = OpenFeature::SDK.build_client()



bool_value = client.fetch_boolean_value(

  flag_key: "my-boolean-flag",

  default_value: false,

  evaluation_context: evaluation_context

)



if bool_value 

  puts "The flag is enabled"

else

  puts "The flag is disabled"

end
```

The evaluation context is the way for the client to specify contextual data that GO Feature Flag uses to evaluate the feature flags, it allows to define rules on the flag.

The `targeting_key` is mandatory for GO Feature Flag to evaluate the feature flag, it could be the id of a user, a session ID or anything you find relevant to use as identifier during the evaluation.

### Evaluate a feature flag[​](#evaluate-a-feature-flag "Direct link to Evaluate a feature flag")

The client is used to retrieve values for the current `EvaluationContext`. For example, retrieving a boolean value for the flag **"my-flag"**:

```
client = OpenFeature::SDK.build_client()



bool_value = client.fetch_boolean_value(

  flag_key: "my-boolean-flag",

  default_value: false,

  evaluation_context: evaluation_context

)
```

GO Feature Flag supports different all OpenFeature supported types of feature flags, it means that you can use all the accessor directly

```
# Bool

client.fetch_boolean_value(flag_key: 'my-flag', default_value: false, evaluation_context: evaluation_context)



# String

client.fetch_string_value(flag_key: 'my-flag', default_value: "default", evaluation_context: evaluation_context)



# Number

client.fetch_number_value(flag_key: 'my-flag', default_value: 0, evaluation_context: evaluation_context)



# Object

client.fetch_object_value(flag_key: 'my-flag', default_value: {"default" => true}, evaluation_context: evaluation_context)
```

## Features status[​](#features-status "Direct link to Features status")

| Status | Feature                      | Description                                                                                                                                                                |
| ------ | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **     | Remote Evaluation            | The provider is calling the remote server to evaluate the feature flags.                                                                                                   |
| **     | Tracking Flag Evaluation     | The provider is tracking all the evaluations of your feature flags and you can export them using an exporter.                                                              |
| **     | In process Evaluation        | The provider is able to evaluate the feature flags in process, it loads the configuration from the relay-proxy and perform evaluation inside the SDK.                      |
| **     | Tracking Custom Events       | The provider is tracking custom events through the track() function of your SDK. All those events are send to the exporter for you to forward them where you want.         |
| **     | Configuration Change Updates | The provider is able to update the configuration based on the configuration, it means that the provider is able to react to any feature flag change on your configuration. |
| **     | Provider Events Reactions    | You can add an event handler to the provider to react to the provider events.                                                                                              |

**Implemented**In-progress**Not implemented yet

## Contributing[​](#contributing "Direct link to Contributing")

This project welcomes contributions from the community. If you're interested in contributing, see the [contributors' guide](https://github.com/thomaspoignant/go-feature-flag/blob/main/CONTRIBUTING.md) for some helpful tips.
