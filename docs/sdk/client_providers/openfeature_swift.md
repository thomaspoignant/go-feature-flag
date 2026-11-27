# Swift SDK

[![version](https://img.shields.io/github/v/release/go-feature-flag/openfeature-swift-provider?label=Swift\&display_name=tag\&style=flat-square\&logo=Swift)](https://github.com/go-feature-flag/openfeature-swift-provider)

In conjunction with the [OpenFeature SDK](https://openfeature.dev/docs/reference/concepts/provider) you will be able to evaluate your feature flags in your **iOS**/**tvOS**/**macOS** applications.

For documentation related to flags management in GO Feature Flag, refer to the [GO Feature Flag documentation website](https://gofeatureflag.org/docs).

### Functionalities:[​](#functionalities "Direct link to Functionalities:")

* Manages the integration of the OpenFeature Swift SDK and GO Feature Flag relay-proxy.
* Prefetch and cache flag evaluations in order to give the flag value in an efficient way.
* Automatic configuration changes polling, to be informed as soon as a flag configuration has changed.
* Automatic data collection about which flags have been accessed by the application

## Dependency Setup[​](#dependency-setup "Direct link to Dependency Setup")

### Swift Package Manager[​](#swift-package-manager "Direct link to Swift Package Manager")

In the dependencies section of Package.swift add:

```
.package(url: "https://github.com/go-feature-flag/openfeature-swift-provider.git", from: "penfeature-swift-provider-v0.7.0")
```

and in the target dependencies section add:

```
.product(name: "GOFeatureFlag", package: "openfeature-swift-provider"),
```

### Xcode Dependencies[​](#xcode-dependencies "Direct link to Xcode Dependencies")

You have two options, both start from File > Add Packages... in the code menu.

First, ensure you have your GitHub account added as an option (`+ > Add Source Control Account...`). You will need to create a [Personal Access Token](https://github.com/settings/tokens) with the permissions defined in the Xcode interface.

1. Add as a remote repository

* Search for `git@github.com:go-feature-flag/openfeature-swift-provider.git` and click "Add Package"

2. Clone the repository locally

* Clone locally using your preferred method
* Use the "Add Local..." button to select the local folder

**Note:** Option 2 is only recommended if you are making changes to the SDK. You will also need to add the relevant OpenFeature SDK dependency manually.

## Getting started[​](#getting-started "Direct link to Getting started")

### Initialize the provider[​](#initialize-the-provider "Direct link to Initialize the provider")

GO Feature Flag provider needs to be created and then set in the global OpenFeatureAPI.

The only required option to create a `GoFeatureFlagProvider` is the URL to your GO Feature Flag relay-proxy instance.

```
import GOFeatureFlag

import OpenFeature



let options = GoFeatureFlagProviderOptions(endpoint: "https://your_domain.io")

let provider = GoFeatureFlagProvider(options: options)



let evaluationContext = MutableContext(targetingKey: "myTargetingKey", structure: MutableStructure())

OpenFeatureAPI.shared.setProvider(provider: provider, initialContext: evaluationContext)
```

The evaluation context is the way for the client to specify contextual data that GO Feature Flag uses to evaluate the feature flags, it allows to define rules on the flag.

The `targetingKey` is mandatory for GO Feature Flag in order to evaluate the feature flag, it could be the id of a user, a session ID or anything you find relevant to use as identifier during the evaluation.

The `setProvider()` function is synchronous and returns immediately, however this does not mean that the provider is ready to be used. An asynchronous network request to the GO Feature Flag backend to fetch all the flags configured for your application must be completed by the provider first. The provider will then emit a `READY` event indicating you can start resolving flags.

If you prefer to wait until the fetch is done you can use the `async/await` compatible API available for waiting the Provider to become ready:

```
await OpenFeatureAPI.shared.setProviderAndWait(provider: provider)
```

### Update the Evaluation Context[​](#update-the-evaluation-context "Direct link to Update the Evaluation Context")

During the usage of your application it may appears that the `EvaluationContext` should be updated. For example if a not logged-in user authenticates, you will probably have to update the evaluation context.

```
let ctx = MutableContext(targetingKey: "myNewTargetingKey", structure: MutableStructure())

OpenFeatureAPI.shared.setEvaluationContext(evaluationContext: ctx)
```

`setEvaluationContext()` is a synchronous function similar to `setProvider()` and will fetch the new version of the feature flags based on this new `EvaluationContext`.

### Limit the flags to evaluate[​](#limit-the-flags-to-evaluate "Direct link to Limit the flags to evaluate")

By default, the provider will fetch all the flags configured in the GO Feature Flag server to be ready to evaluate them. If you know in advance, what are the flags you will evaluate in your application, you can specify the list of flags to evaluate in the context.

You need to add in the evaluation context the restricted key `gofeatureflag.flagList` with the list of flags you want to evaluate.

```
let ctx = MutableContext(targetingKey: "myNewTargetingKey")

ctx.add(

    key: "gofeatureflag",

    value: Value.list([

        Value.string("flag1"),

        Value.string("flag2")

    ])

)

OpenFeatureAPI.shared.setEvaluationContext(evaluationContext: ctx)
```

By setting the `gofeatureflag.flagList` key in the context, the provider will only fetch the flags specified in the list.

warning

When limiting the flags to evaluate, if you try to evaluate a flag not in the list, the provider will return the default value with the error `FLAG_NOT_FOUND`.

### Evaluate a feature flag[​](#evaluate-a-feature-flag "Direct link to Evaluate a feature flag")

The client is used to retrieve values for the current `EvaluationContext`. For example, retrieving a boolean value for the flag **"my-flag"**:

```
let client = OpenFeatureAPI.shared.getClient()

let result = client.getBooleanValue(key: "my-flag", defaultValue: false)
```

GO Feature Flag supports different all OpenFeature supported types of feature flags, it means that you can use all the accessor directly

```
// Bool

client.getBooleanValue(key: "my-flag", defaultValue: false)



// String

client.getStringValue(key: "my-flag", defaultValue: "default")



// Integer

client.getIntegerValue(key: "my-flag", defaultValue: 1)



// Double

client.getDoubleValue(key: "my-flag", defaultValue: 1.1)



// Object

client.getObjectValue(key: "my-flag", defaultValue: Value.structure(["key":Value.integer(1234)]))
```

note

If you add a new flag in GO Feature Flag, expect some delay before having it available for the provider. Refreshing the cache from remote happens when setting a new provider and/or evaluation context in the global OpenFeatureAPI, but also when a configuration change is detected during the polling.

### Handling Provider Events[​](#handling-provider-events "Direct link to Handling Provider Events")

When setting the provider or the context *(via `setEvaluationContext()` or `setProvider()`)* some events can be triggered to know the state of the provider.

To listen to them you can add an event handler via the `OpenFeatureAPI` shared instance:

```
OpenFeatureAPI.shared.observe().sink { event in

    if event == .error {

        // An error has been emitted

    }

}
```

#### Existing type of events are:[​](#existing-type-of-events-are "Direct link to Existing type of events are:")

* `.ready`: Provider is ready.
* `.error`: Provider in error.
* `.configurationChanged`: Configuration has changed in GO Feature Flag.
* `.stale`: Provider has not the latest version of the feature flags.
* `.notReady`: Provider is not ready to evaluate the feature flags.

## Features status[​](#features-status "Direct link to Features status")

| Status | Feature                      | Description                                                                                                                                                                |
| ------ | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **     | Remote Evaluation            | The provider is calling the remote server to evaluate the feature flags.                                                                                                   |
| **     | Tracking Flag Evaluation     | The provider is tracking all the evaluations of your feature flags and you can export them using an exporter.                                                              |
| **     | Configuration Change Updates | The provider is able to update the configuration based on the configuration, it means that the provider is able to react to any feature flag change on your configuration. |
| **     | Provider Events Reactions    | You can add an event handler to the provider to react to the provider events.                                                                                              |
| **     | Tracking Custom Events       | The provider is tracking custom events through the track() function of your SDK. All those events are send to the exporter for you to forward them where you want.         |

**Implemented**In-progress**Not implemented yet

## Contribute to the provider[​](#contribute-to-the-provider "Direct link to Contribute to the provider")

You can find the source of the provider in the [`openfeature-swift-provider`](https://github.com/go-feature-flag/openfeature-swift-provider) repository.
