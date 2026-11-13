# Android / Kotlin

[![Maven Central Version](https://img.shields.io/maven-central/v/org.gofeatureflag.openfeature/gofeatureflag-kotlin-provider?color=blue\&logo=android\&style=flat-square)](https://search.maven.org/artifact/org.gofeatureflag.openfeature/gofeatureflag-kotlin-provider)

This OpenFeature provider has a Kotlin implementation for Android to communicate with the GO Feature Flag Server.

In conjunction with the [OpenFeature SDK](https://openfeature.dev/docs/reference/technologies/client/kotlin) you will be able to evaluate your feature flags in your **Android** applications.

For documentation related to flags management in GO Feature Flag, refer to the [GO Feature Flag documentation website](https://gofeatureflag.org/docs).

### Functionalities:[​](#functionalities "Direct link to Functionalities:")

* Manage the integration of the OpenFeature Android SDK and GO Feature Flag relay-proxy.
* Prefetch and cache flag evaluations in order to give the flag value in an efficient way.
* Automatic configuration changes polling, to be informed as soon as a flag configuration has changed.
* Automatic data collection about which flags have been accessed by the application

## Dependency Setup[​](#dependency-setup "Direct link to Dependency Setup")

```


dependencies {

  api("dev.openfeature:kotlin-sdk:0.4.1")

  implementation("org.gofeatureflag.openfeature:gofeatureflag-kotlin-provider:0.3.0")

}
```

## Getting started[​](#getting-started "Direct link to Getting started")

### Initialize the provider[​](#initialize-the-provider "Direct link to Initialize the provider")

GO Feature Flag provider needs to be created and then set in the global OpenFeatureAPI.

The only required option to create a `GoFeatureFlagProvider` is the URL to your GO Feature Flag relay-proxy instance.

```
import org.gofeatureflag.openfeature.bean.GoFeatureFlagOptions

import org.gofeatureflag.openfeature.GoFeatureFlagProvider

// ...



val evaluationContext: EvaluationContext = ImmutableContext(

        targetingKey = UUID.randomUUID().toString(),

        attributes = mapOf( "age" to Value.Integer(22), "email" to Value.String("contact@gofeatureflag.org"))

    )



OpenFeatureAPI.setProvider(

    GoFeatureFlagProvider(

        options = GoFeatureFlagOptions( endpoint = "http://localhost:1031")

    ),

    evaluationContext

)
```

The evaluation context is the way for the client to specify contextual data that GO Feature Flag uses to evaluate the feature flags, it allows defining rules on the flag.

The `targetingKey` is mandatory for GO Feature Flag in order to evaluate the feature flag, it could be the id of a user, a session ID or anything you find relevant to use as identifier during the evaluation.

The `setProvider()` function is synchronous and returns immediately however, this does not mean that the provider is ready to be used. An asynchronous network request to the GO Feature Flag backend to fetch all the flags configured for your application must be completed by the provider first. The provider will then emit a `READY` event indicating you can start resolving flags.

If you prefer to wait until the fetch is done you can use the `suspend` compatible API available for waiting the Provider to become ready:

```
runBlocking{

  OpenFeatureAPI.shared.setProviderAndWait(

    provider = provider,

    dispatcher = Dispatchers.IO,

    initialContext = evaluationContext

  )

}
```

### Available options[​](#available-options "Direct link to Available options")

When initializing the provider, you can pass some options to configure the provider and how it behaves with GO Feature Flag.

| Option name               | Type               | Default | Description                                                                                                                                                                                     |
| ------------------------- | ------------------ | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `endpoint`                | String             |         | endpoint is the URL where your GO Feature Flag server is located.                                                                                                                               |
| `timeout`                 | Long               | 10000   | (optional) timeout is the time in millisecond we wait for an answer from the server.                                                                                                            |
| `maxIdleConnections`      | Int                | 1000    | (optional) maxIdleConnections is the maximum number of connexions in the connexion pool.                                                                                                        |
| `keepAliveDuration`       | Long               | 7200000 | (optional) keepAliveDuration is the time in millisecond we keep the connexion open.                                                                                                             |
| `apiKey`                  | String             |         | (optional) If GO Feature Flag is configured to authenticate the requests, you should provide an API Key to the provider. Please ask the administrator of the relay proxy to provide an API Key. |
| `pollingIntervalInMillis` | Long               | 300000  | (optional) Polling interval in millisecond to check with GO Feature Flag relay proxy if there is a flag configuration change.                                                                   |
| `flushIntervalMs`         | Long               | 1000    | (optional) Time to wait before calling GO Feature Flag to store all the data about the evaluation in the relay proxy.                                                                           |
| `customHeaders`           | Map\[string]string |         | (optional) Custom HTTP headers that will be added to the requests sent to the relay proxy (for example to set a custom user-agent).                                                             |

### Update the Evaluation Context[​](#update-the-evaluation-context "Direct link to Update the Evaluation Context")

During the usage of your application it may appear that the `EvaluationContext` should be updated. For example, if a not logged-in user authenticates, you will probably have to update the evaluation context.

```
val newContext: EvaluationContext = ImmutableContext(

    targetingKey = userId,

    attributes = mapOf( "age" to Value.Integer(32), "email" to Value.String("batman@gofeatureflag.org"))

)



OpenFeatureAPI.setEvaluationContext(newContext)
```

`setEvaluationContext()` is a synchronous function similar to `setProvider()` and will fetch the new version of the feature flags based on this new `EvaluationContext`.

### Limit the flags to evaluate[​](#limit-the-flags-to-evaluate "Direct link to Limit the flags to evaluate")

By default, the provider will fetch all the flags configured in the GO Feature Flag server to be ready to evaluate them. If you know in advance, what are the flags you will evaluate in your application, you can specify the list of flags to evaluate in the context.

You need to add in the evaluation context the restricted key `gofeatureflag.flagList` with the list of flags you want to evaluate.

```
val newContext: EvaluationContext = ImmutableContext(

            targetingKey = "userId",

            attributes = mapOf(

                "gofeatureflag" to Value.Structure(

                    mapOf(

                        "flagList" to Value.List(

                            listOf(

                                // list of flags to evaluate

                                Value.String("flag1"),

                                Value.String("flag2"),

                                Value.String("flag3")

                            )

                        ),

                    )

                ),

            )

        )



OpenFeatureAPI.setEvaluationContext(newContext)
```

By setting the `gofeatureflag.flagList` key in the context, the provider will only fetch the flags specified in the list.

warning

When limiting the flags to evaluate, if you try to evaluate a flag not in the list, the provider will return the default value with the error `FLAG_NOT_FOUND`.

### Evaluate a feature flag[​](#evaluate-a-feature-flag "Direct link to Evaluate a feature flag")

The client is used to retrieve values for the current `EvaluationContext`. For example, retrieving a boolean value for the flag **"my-flag"**:

```
val client = OpenFeatureAPI.getClient()

val result = client.getBooleanValue("my-flag", false)
```

GO Feature Flag supports different all OpenFeature supported types of feature flags, it means that you can use all the accessor directly

```
// Bool

client.getBooleanValue("my-flag", false)



// String

client.getStringValue("my-flag", "default")



// Integer

client.getIntegerValue("my-flag", 1)



// Double

client.getDoubleValue("my-flag", 1.1)



// Object

client.getObjectValue("my-flag", Value.structure(mapOf("email" to Value.String("contact@gofeatureflag.org"))))
```

> \[!NOTE] If you add a new flag in GO Feature Flag, expect some delay before having it available for the provider. Refreshing the cache from remote happens when setting a new provider and/or evaluation context in the global OpenFeatureAPI, but also when a configuration change is detected during the polling.

### Handling Provider Events[​](#handling-provider-events "Direct link to Handling Provider Events")

When setting the provider or the context *(via `setEvaluationContext()` or `setProvider()`)* some events can be triggered to know the state of the provider.

To listen to them, you can add an event handler via the `OpenFeatureAPI` shared instance:

```
val coroutineScope = CoroutineScope(Dispatchers.IO)

coroutineScope.launch {

    provider.observe<OpenFeatureEvents.ProviderStale>().collect {

        providerStaleEventReceived = true

    }

}
```

#### Existing type of events are:[​](#existing-type-of-events-are "Direct link to Existing type of events are:")

* `ProviderReady`: Provider is ready.
* `ProviderError`: Provider in error.
* `ProviderStale`: Provider has not the latest version of the feature flags.
* `ProviderNotReady`: Provider is not ready to evaluate the feature flags.

## Features status[​](#features-status "Direct link to Features status")

| Status | Feature                      | Description                                                                                                                                                                |
| ------ | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **     | Remote Evaluation            | The provider is calling the remote server to evaluate the feature flags.                                                                                                   |
| **     | Tracking Flag Evaluation     | The provider is tracking all the evaluations of your feature flags and you can export them using an exporter.                                                              |
| **     | Tracking Custom Events       | The provider is tracking custom events through the track() function of your SDK. All those events are send to the exporter for you to forward them where you want.         |
| **     | Configuration Change Updates | The provider is able to update the configuration based on the configuration, it means that the provider is able to react to any feature flag change on your configuration. |
| **     | Provider Events Reactions    | You can add an event handler to the provider to react to the provider events.                                                                                              |

**Implemented**In-progress**Not implemented yet
