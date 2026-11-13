# Java / Kotlin

[![Maven Central Version](https://img.shields.io/maven-central/v/dev.openfeature.contrib.providers/go-feature-flag?color=blue\&style=flat-square)](https://search.maven.org/artifact/dev.openfeature.contrib.providers/go-feature-flag)

The Java provider for GO Feature Flag is a server provider that allows you to evaluate feature flags in your Java application. It allows you to:

* Manage the integration of the OpenFeature Java SDK and GO Feature Flag relay-proxy.

* 2 types of evaluations available:

  <!-- -->

  * **In process**: fetch the flag configuration from the GO Feature Flag relay-proxy API and evaluate the flags directly in the provider.
  * **Remote**: Call the GO Feature Flag relay-proxy for each flag evaluation.

* Collect and send evaluation data to the GO Feature Flag relay-proxy for statistics and monitoring purposes.

* Support the OpenFeature [tracking API](https://openfeature.dev/docs/reference/concepts/tracking/) to associate metrics or KPIs with feature flag evaluation contexts.

## Install dependencies[​](#install-dependencies "Direct link to Install dependencies")

The first thing we will do is install the **Open Feature SDK** and the **GO Feature Flag provider**.

* maven
* gradle

```
<dependency>

  <groupId>dev.openfeature</groupId>

  <artifactId>sdk</artifactId>

  <version>1.15.1</version>

</dependency>

<dependency>

  <groupId>dev.openfeature.contrib.providers</groupId>

  <artifactId>go-feature-flag</artifactId>

  <version>0.4.3</version>

</dependency>
```

```
implementation group: 'dev.openfeature', name: 'javasdk', version: '1.15.1'

implementation group: 'dev.openfeature.contrib.providers', name: 'go-feature-flag', version: '0.4.3'
```

## Getting started[​](#getting-started "Direct link to Getting started")

### Initialize the provider[​](#initialize-the-provider "Direct link to Initialize the provider")

GO Feature Flag provider needs to be created and then set in the global OpenFeatureAPI.

The only required option to create a `GoFeatureFlagProvider` is the endpoint to your GO Feature Flag relay-proxy instance.

* java
* Kotlin

```
import dev.openfeature.contrib.providers.gofeatureflag;

//...



FeatureProvider provider = new GoFeatureFlagProvider(

        GoFeatureFlagProviderOptions.builder()

                .endpoint("https://my-instance.gofeatureflag.org")

                .build());



OpenFeatureAPI.getInstance().setProviderAndWait(provider);

// ...

Client client = OpenFeatureAPI.getInstance().getClient("my-goff-provider");



// targetingKey is mandatory for each evaluation

String targetingKey = "ad0c6f75-f5d6-4b17-b8eb-6c923d8d4698";

EvaluationContext evaluationContext = new ImmutableContext(targetingKey);



// Example of a boolean flag evaluation

FlagEvaluationDetails<Boolean> booleanFlagEvaluation = client.getBooleanValue("bool_targeting_match", false, evaluationContext);
```

```
import dev.openfeature.contrib.providers.gofeatureflag.*



//...



val provider = GoFeatureFlagProvider(

    GoFeatureFlagProviderOptions.builder()

        .endpoint("https://my-instance.gofeatureflag.org")

        .build()

)



OpenFeatureAPI.getInstance().setProviderAndWait(provider)

// ...

val client = OpenFeatureAPI.getInstance().getClient("my-goff-provider")



// targetingKey is mandatory for each evaluation

val targetingKey = "ad0c6f75-f5d6-4b17-b8eb-6c923d8d4698"

val evaluationContext = ImmutableContext(targetingKey)



// Example of a boolean flag evaluation

val booleanFlagEvaluation = client.getBooleanValue("bool_targeting_match", false, evaluationContext)
```

```
```

The evaluation context is the way for the client to specify contextual data that GO Feature Flag uses to evaluate the feature flags, it allows to define rules on the flag.

The `targetingKey` is mandatory for GO Feature Flag in order to evaluate the feature flag, it could be the id of a user, a session ID or anything you find relevant to use as identifier during the evaluation.

### Configure the provider[​](#configure-the-provider "Direct link to Configure the provider")

You can configure the provider with several options to customize its behavior. The following options are available:

| name                              | mandatory | Description                                                                                                                                                                                                                                                                                                                                                             |
| --------------------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`endpoint`**                    | `true`    | endpoint contains the DNS of your GO Feature Flag relay proxy *(ex: <https://mydomain.com/gofeatureflagproxy/>)*                                                                                                                                                                                                                                                        |
| **`evaluationType`**              | `false`   | evaluationType is the type of evaluation you want to use.- If you want to have a local evaluation, you should use IN\_PROCESS.<br />- If you want to have an evaluation on the relay-proxy directly, you should use REMOTE.Default: IN\_PROCESS<br />                                                                                                                   |
| **`timeout`**                     | `false`   | timeout in millisecond we are waiting when calling the relay proxy API. *(default: `10000`)*                                                                                                                                                                                                                                                                            |
| **`maxIdleConnections`**          | `false`   | maxIdleConnections is the maximum number of connections in the connection pool. *(default: `1000`)*                                                                                                                                                                                                                                                                     |
| **`keepAliveDuration`**           | `false`   | keepAliveDuration is the time in millisecond we keep the connection open. *(default: `7200000` (2 hours))*                                                                                                                                                                                                                                                              |
| **`apiKey`**                      | `false`   | If the relay proxy is configured to authenticate the requests, you should provide an API Key to the provider. Please ask the administrator of the relay proxy to provide an API Key. (This feature is available only if you are using GO Feature Flag relay proxy v1.7.0 or above). *(default: null)*                                                                   |
| **`flushIntervalMs`**             | `false`   | interval time we publish statistics collection data to the proxy. The parameter is used only if the cache is enabled, otherwise the collection of the data is done directly when calling the evaluation API. default: `1000` ms                                                                                                                                         |
| **`maxPendingEvents`**            | `false`   | max pending events aggregated before publishing for collection data to the proxy. When event is added while events collection is full, event is omitted. *(default: `10000`)*                                                                                                                                                                                           |
| **`disableDataCollection`**       | `false`   | set to true if you don't want to collect the usage of flags retrieved in the cache. *(default: `false`)*                                                                                                                                                                                                                                                                |
| **`exporterMetadata`**            | `false`   | exporterMetadata is the metadata we send to the GO Feature Flag relay proxy when we report the evaluation data usage.                                                                                                                                                                                                                                                   |
| **`evaluationFlagList`**          | `false`   | If you are using in process evaluation, by default we will load in memory all the flags available in the relay proxy. If you want to limit the number of flags loaded in memory, you can use this parameter. By setting this parameter, you will only load the flags available in the list.If null or empty, all the flags available in the relay proxy will be loaded. |
| **`flagChangePollingIntervalMs`** | `false`   | interval time we poll the proxy to check if the configuration has changed. It is used for the in process evaluation to check if we should refresh our internal cache. default: `120000`                                                                                                                                                                                 |

## Evaluate your flag[​](#evaluate-your-flag "Direct link to Evaluate your flag")

This code block explain how you can create an `EvaluationContext` and use it to evaluate your flag.

note

In this example we are evaluating a `boolean` flag, but other types are available.

**Refer to the [Open Feature documentation](https://docs.openfeature.dev/docs/reference/concepts/evaluation-api#basic-evaluation) to know more about it.**

* java
* Kotlin

```
// Context of your flag evaluation.

// With GO Feature Flag you MUST have a targetingKey that is a unique identifier of the user.

EvaluationContext userContext = new MutableContext("1d1b9238-2591-4a47-94cf-d2bc080892f1")

  .add("firstname", "john")

  .add("lastname", "doe")

  .add("email","john.doe@gofeatureflag.org")

  .add("admin", true)

  .add("anonymous", false);



Boolean adminFlag = featureFlagClient.getBooleanValue("flag-only-for-admin", false, userContext);

if (adminFlag) {

  // flag "flag-only-for-admin" is true for the user

} else {

  // flag "flag-only-for-admin" is false for the user

}
```

```
// Context of your flag evaluation.

// With GO Feature Flag you MUST have a targetingKey that is a unique identifier of the user.

val userContext: EvaluationContext = MutableContext("1d1b9238-2591-4a47-94cf-d2bc080892f1")

  .add("firstname", "john")

  .add("lastname", "doe")

  .add("email", "john.doe@gofeatureflag.org")

  .add("admin", true)

  .add("anonymous", false)



val adminFlag = featureFlagClient.getBooleanValue("bool_targeting_match", false, userContext)

if (adminFlag) {

  // flag "flag-only-for-admin" is true for the user

} else {

  // flag "flag-only-for-admin" is false for the user

}
```

## Breaking changes[​](#breaking-changes "Direct link to Breaking changes")

### 1.0.0 - Introduce In Process Evaluation[​](#100---introduce-in-process-evaluation "Direct link to 1.0.0 - Introduce In Process Evaluation")

warning

This version of the provider requires to use GO Feature Flag relay-proxy `v1.45.0` or above. If you have an older version of the relay-proxy, please use the version `0.4.3` of the provider.

The version `1.0.0` of the provider introduces the in process evaluation.<br /><!-- -->This evaluation type is more efficient than the remote evaluation because it doesn't require to call the relay-proxy API for each flag evaluation.

This is a new major version, some features have been added and some have been deprecated.

### 0.4.0 - Cache Implementation Change: Guava to Caffeine[​](#040---cache-implementation-change-guava-to-caffeine "Direct link to 0.4.0 - Cache Implementation Change: Guava to Caffeine")

In this release, we have updated the cache implementation from Guava to Caffeine. This change was made because Caffeine is now the recommended caching solution by the maintainers of Guava due to its performance improvements and enhanced features.

Because of this, the cache configuration on `GoFeatureFlagProviderOptions` that used Guava's `CacheBuilder` is now handled by `Caffeine`.

#### How to migrate[​](#how-to-migrate "Direct link to How to migrate")

Configuration cache with Guava used to be like this:

```
import com.google.common.cache.CacheBuilder;

// ...

CacheBuilder guavaCacheBuilder = CacheBuilder.newBuilder()

  .initialCapacity(100)

  .maximumSize(2000);



FeatureProvider provider = new GoFeatureFlagProvider(

  GoFeatureFlagProviderOptions

  .builder()

  .endpoint("https://my-gofeatureflag-instance.org")

  .cacheBuilder(guavaCacheBuilder)

  .build());



OpenFeatureAPI.getInstance().setProviderAndWait(provider);



// ...
```

Now with Caffeine it should be like this:

```
import com.github.benmanes.caffeine.cache.Caffeine;

// ...

Caffeine caffeineCacheConfig = Caffeine.newBuilder()

  .initialCapacity(100)

  .maximumSize(2000);



FeatureProvider provider = new GoFeatureFlagProvider(

  GoFeatureFlagProviderOptions

  .builder()

  .endpoint("https://my-gofeatureflag-instance.org")

  .cacheConfig(caffeineCacheConfig)

  .build());



OpenFeatureAPI.getInstance().setProviderAndWait(provider);



// ...
```

For a complete list of customizations options available in Caffeine, please refer to the [Caffeine documentation](https://github.com/ben-manes/caffeine/wiki) for more details.

## Features status[​](#features-status "Direct link to Features status")

| Status | Feature                      | Description                                                                                                                                                                |
| ------ | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **     | In process Evaluation        | The provider is able to evaluate the feature flags in process, it loads the configuration from the relay-proxy and perform evaluation inside the SDK.                      |
| **     | Remote Evaluation            | The provider is calling the remote server to evaluate the feature flags.                                                                                                   |
| **     | Tracking Flag Evaluation     | The provider is tracking all the evaluations of your feature flags and you can export them using an exporter.                                                              |
| **     | Tracking Custom Events       | The provider is tracking custom events through the track() function of your SDK. All those events are send to the exporter for you to forward them where you want.         |
| **     | Configuration Change Updates | The provider is able to update the configuration based on the configuration, it means that the provider is able to react to any feature flag change on your configuration. |
| **     | Provider Events Reactions    | You can add an event handler to the provider to react to the provider events.                                                                                              |

**Implemented**In-progress**Not implemented yet

## Contribute to the provider[​](#contribute-to-the-provider "Direct link to Contribute to the provider")

You can find the source of the provider in the [`open-feature/java-sdk-contrib`](https://github.com/open-feature/java-sdk-contrib/tree/main/providers/go-feature-flag) repository.
