# GO Feature Flag - OpenFeature Android provider
<p align="center">
   <a href="https://search.maven.org/artifact/org.gofeatureflag.openfeature/gofeatureflag-kotlin-provider"><img src="https://img.shields.io/maven-central/v/org.gofeatureflag.openfeature/gofeatureflag-kotlin-provider?color=blue" alt="Maven Release"></a>
   <img alt="Static Badge" src="https://img.shields.io/badge/android-version%3Fcolor%3Ablue?logo=android&logoColor=blue&logoSize=10px">
   <a href="https://gofeatureflag.org/slack"><img src="https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=green" alt="Join us on slack"></a>
</p>

This repository contains the official Kotlin OpenFeature provider for Android, to access your feature flags with [GO Feature Flag](https://gofeatureflag.org).

In conjuction with the [OpenFeature SDK](https://openfeature.dev/docs/reference/technologies/client/kotlin) you will be able to evaluate your feature flags in your **Android** applications.

For documentation related to flags management in GO Feature Flag, refer to the [GO Feature Flag documentation website](https://gofeatureflag.org/docs).

### Functionalities:
- Manage the integration of the OpenFeature Android SDK and GO Feature Flag relay-proxy.
- Prefetch and cache flag evaluations in order to give the flag value in an efficient way.
- Automatic configuration changes polling, to be informed as soon as a flag configuration has changed.
- Automatic data collection about which flags have been accessed by the application

## Dependency Setup

```kotlin
dependencies {
    api("dev.openfeature:kotlin-sdk:0.7.0") // check on maven central to get the latest version
    implementation("org.gofeatureflag.openfeature:gofeatureflag-kotlin-provider:1.0.0") // check on maven central to get the latest version
}
```

## Getting started

### Initialize the provider

GO Feature Flag provider needs to be created and then set in the global OpenFeatureAPI.

The only required option to create a `GoFeatureFlagProvider` is the URL to your GO Feature Flag relay-proxy instance.

```kotlin
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

The `targetingKey` is mandatory for GO Feature Flag in order to evaluate the feature flag, it could be the id of a user, a session ID or anything you find relevent to use as identifier during the evaluation.

The `setProvider()` function is synchronous and returns immediately however, this does not mean that the provider is ready to be used.
An asynchronous network request to the GO Feature Flag backend to fetch all the flags configured for your application must be completed by the provider first. The provider will then emit a `READY` event indicating you can start resolving flags.

If you prefer to wait until the fetch is done you can use the `suspend` compatible API available for waiting the Provider to become ready:

```kotlin
runBlocking{
  OpenFeatureAPI.shared.setProviderAndWait(
    provider = provider,
    dispatcher = Dispatchers.IO,
    initialContext = evaluationContext
  )
}
```

### Available options

When initializing the provider, you can pass some options to configure the provider and how it behaves with GO Feature Flag.

| Option name               | Type   | Default | Description                                                                                                                                                                                     |
|---------------------------|--------|---------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `endpoint`                | String |         | endpoint is the URL where your GO Feature Flag server is located.                                                                                                                               |
| `timeout`                 | Long   | 10000   | (optional) timeout is the time in millisecond we wait for an answer from the server.                                                                                                            |
| `maxIdleConnections`      | Int    | 1000    | (optional) maxIdleConnections is the maximum number of connexions in the connexion pool.                                                                                                        |
| `keepAliveDuration`       | Long   | 7200000 | (optional) keepAliveDuration is the time in millisecond we keep the connexion open.                                                                                                             |
| `apiKey`                  | String |         | (optional) If GO Feature Flag is configured to authenticate the requests, you should provide an API Key to the provider. Please ask the administrator of the relay proxy to provide an API Key. |
| `pollingIntervalInMillis` | Long   | 300000  | (optional) Polling interval in millisecond to check with GO Feature Flag relay proxy if there is a flag configuration change.                                                                   |
| `flushIntervalMs`         | Long   | 1000    | (optional) Time to wait before calling GO Feature Flag to store all the data about the evaluation in the relay proxy.                                                                           |

### Update the Evaluation Context

During the usage of your application it may appear that the `EvaluationContext` should be updated. For example, if a not logged-in user authentify himself, you will probably have to update the evaluation context.

```kotlin
val newContext: EvaluationContext = ImmutableContext(
    targetingKey = userId,
    attributes = mapOf( "age" to Value.Integer(32), "email" to Value.String("batman@gofeatureflag.org"))
)

OpenFeatureAPI.setEvaluationContext(newEvalCtx)
```

`setEvaluationContext()` is a synchronous function similar to `setProvider()` and will fetch the new version of the feature flags based on this new `EvaluationContext`.

### Evaluate a feature flag
The client is used to retrieve values for the current `EvaluationContext`. For example, retrieving a boolean value for the flag **"my-flag"**:

```kotlin
val client = OpenFeatureAPI.getClient()
val result = client.getBooleanValue("my-flag", false)
```

GO Feature Flag supports different all OpenFeature supported types of feature flags, it means that you can use all the accessor directly
```kotlin
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

> [!NOTE]  
> If you add a new flag in GO Feature Flag, expect some delay before having it available for the provider.
> Refreshing the cache from remote happens when setting a new provider and/or evaluation context in the global OpenFeatureAPI, but also when a configuration change is detected during the polling.

### Handling Provider Events

When setting the provider or the context *(via `setEvaluationContext()` or `setProvider()`)* some events can be triggered to know the state of the provider.

To listen to them, you can add an event handler via the `OpenFeatureAPI` shared instance:

```kotlin
val coroutineScope = CoroutineScope(Dispatchers.IO)
coroutineScope.launch {
    provider.observe<OpenFeatureEvents.ProviderStale>().collect {
        providerStaleEventReceived = true
    }
}
```

#### Existing type of events are:
- `ProviderReady`: Provider is ready.
- `ProviderError`: Provider in error.
- `ProviderStale`: Provider has not the latest version of the feature flags.
- `ProviderNotReady`: Provider is not ready to evaluate the feature flags.

## Features status

| Status | Feature            | Description                                                                                                                                         |
|--------|--------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| ✅      | Flag evaluation    | It is possible to evaluate all the type of flags                                                                                                    |
| ✅      | Cache invalidation | A polling mechanism is in place to refresh the cache in case of configuration change                                                                |
| ❌      | Logging            | Not supported by the SDK                                                                                                                            |
| ✅      | Flag Metadata      | You have access to your flag metadata                                                                                                               |
| ✅      | Event Streaming    | You can register to receive some internal event from the provider                                                                                   |
| ✅      | Unit test          | The test are running one by one, but we still have an [issue open](https://github.com/open-feature/kotlin-sdk/issues/108) to enable fully the tests |

<sub>Implemented: ✅ | In-progress: ⚠️ | Not implemented yet: ❌</sub>

## Contributing
This project welcomes contributions from the community.
If you're interested in contributing, see the [contributors' guide](https://github.com/thomaspoignant/go-feature-flag/blob/main/CONTRIBUTING.md) for some helpful tips.