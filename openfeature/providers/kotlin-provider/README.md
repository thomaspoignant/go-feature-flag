# GO Feature Flag Kotlin OpenFeature Provider for Android

![Static Badge](https://img.shields.io/badge/status-experimental-red)

This OpenFeature provider is a Kotlin implementation for Android to communicate with the GO Feature
Flag Server.

The OpenFeature Kotlin is experimental, and the provider is also experimental.  
We don't recommend using this in production yet.

## About this provider

[GO Feature Flag](https://gofeatureflag.org) provider allows you to connect to your GO Feature Flag
instance with the OpenFeature Kotlin SDK.

This is a client provider made for Android, we do not recommend using it in a server environment.  
If you want to use it in a server environment, you should use
the [`Java` provider](https://gofeatureflag.org/docs/openfeature_sdk/server_providers/openfeature_java).

## What is GO Feature Flag?

GO Feature Flag is a simple, complete and lightweight self-hosted feature flag solution 100% Open
Source.  
Our focus is to avoid any complex infrastructure work to use GO Feature Flag.

This is a complete feature flagging solution with the possibility to target only a group of users,
use any types of flags, store your configuration in various location and advanced rollout
functionality. You can also collect usage data of your flags and be notified of configuration
changes.

## Install the provider

TODO

## How to use the provider?

```kotlin
val evaluationContext = ImmutableContext(
    targetingKey = "0a23d9a5-0a8f-42c9-9f5f-4de3afd6cf99",
    attributes = mutableMapOf(
        "region" to Value.String("us-east-1"),
        "email" to Value.String("john.doe@gofeatureflag.org")
    )
)

OpenFeatureAPI.setProvider(
    GoFeatureFlagProvider(
        options = GoFeatureFlagOptions(
            endpoint = "http://localhost:1031"
        )
    ), evaluationContext
)

val client = OpenFeatureAPI.getClient("my-client")
if (client.getBooleanValue("my-flag", false)) {
    println("my-flag is enabled")
}
OpenFeatureAPI.shutdown()
```

### Available options

| Option name        | Type   | Default | Description                                                                                                                                                                                     |
|--------------------|--------|---------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| endpoint           | String |         | endpoint is the URL where your GO Feature Flag server is located.                                                                                                                               |
| timeout            | Long   | 10000   | (optional) timeout is the time in millisecond we wait for an answer from the server.                                                                                                            |
| maxIdleConnections | Int    | 1000    | (optional) maxIdleConnections is the maximum number of connexions in the connexion pool.                                                                                                        |
| keepAliveDuration  | Long   | 7200000 | (optional) keepAliveDuration is the time in millisecond we keep the connexion open.                                                                                                             |
| apiKey             | String |         | (optional) If GO Feature Flag is configured to authenticate the requests, you should provide an API Key to the provider. Please ask the administrator of the relay proxy to provide an API Key. |
| retryDelay         | Long   | 300     | (optional) delay in millisecond to wait before retrying to connect the websocket                                                                                                                |

### Reconnection

If the connection to the GO Feature Flag instance fails, the provider will attempt to reconnect.

### Event streaming

Event streaming is not implemented yet in the GO Feature Flag provider.

## Features status

| Status | Feature            | Description                                                                          |
|--------|--------------------|--------------------------------------------------------------------------------------|
| ✅      | Flag evaluation    | It is possible to evaluate all the type of flags                                     |
| ✅      | Cache invalidation | Websocket mechanism is in place to refresh the cache in case of configuration change |
| ❌      | Logging            | Not supported by the SDK                                                             |
| ❌      | Flag Metadata      | Not supported by the SDK                                                             |
| ❌      | Event Streaming    | Not implemented                                                                      |
| ❌      | Unit test          | Not implemented                                                                      |

<sub>Implemented: ✅ | In-progress: ⚠️ | Not implemented yet: ❌</sub>


