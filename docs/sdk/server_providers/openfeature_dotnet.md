# .NET

[![NuGet Version](https://img.shields.io/nuget/v/OpenFeature.Providers.GOFeatureFlag?color=blue\&style=flat-square)](https://www.nuget.org/packages/OpenFeature.Providers.GOFeatureFlag) ![NuGet Downloads](https://img.shields.io/nuget/dt/OpenFeature.Providers.GOFeatureFlag?style=flat-square)

The .NET provider for GO Feature Flag is a server provider that allows you to evaluate feature flags in your .NET application. It allows you to:

* Manage the integration of the OpenFeature .NET SDK and GO Feature Flag relay-proxy.

* 2 types of evaluations available:

  <!-- -->

  * **In process**: fetch the flag configuration from the GO Feature Flag relay-proxy API and evaluate the flags directly in the provider.
  * **Remote**: Call the GO Feature Flag relay-proxy for each flag evaluation.

* Collect and send evaluation data to the GO Feature Flag relay-proxy for statistics and monitoring purposes.

* Support the OpenFeature [tracking API](https://openfeature.dev/docs/reference/concepts/tracking/) to associate metrics or KPIs with feature flag evaluation contexts.

## Install dependencies[​](#install-dependencies "Direct link to Install dependencies")

The first things we will do is install the **Open Feature SDK** and the **GO Feature Flag provider**.

* .NET CLI
* Package Manager
* Package Reference
* Paket cli
* Cake

```
dotnet add package OpenFeature.Providers.GOFeatureFlag
```

```
NuGet\Install-Package OpenFeature.Providers.GOFeatureFlag
```

```
<PackageReference Include="OpenFeature.Providers.GOFeatureFlag" />
```

```
paket add OpenFeature.Providers.GOFeatureFlag
```

```
// Install OpenFeature.Providers.GOFeatureFlag as a Cake Addin

#addin nuget:?package=OpenFeature.Providers.GOFeatureFlag



// Install OpenFeature.Providers.GOFeatureFlag as a Cake Tool

#tool nuget:?package=OpenFeature.Providers.GOFeatureFlag
```

## Getting started[​](#getting-started "Direct link to Getting started")

### Initialize the provider[​](#initialize-the-provider "Direct link to Initialize the provider")

GO Feature Flag provider needs to be created and then set in the global OpenFeatureAPI.

The only required option to create a `GOFeatureFlagProvider` is the endpoint to your GO Feature Flag relay-proxy instance.

* C#

```
using OpenFeature;

using OpenFeature.Providers.GOFeatureFlag;



// ...



var options = new GOFeatureFlagProviderOptions { Endpoint = "https://my-instance.gofeatureflag.org" };

var provider = new GOFeatureFlagProvider(options);



// Associate the provider with the OpenFeature API

await Api.Instance.SetProviderAsync("my-app", provider);



// Create a client to perform feature flag evaluations

var client = Api.Instance.GetClient("my-app");



// targetingKey is mandatory for each evaluation

var evaluationContext = EvaluationContext.Builder()

  .SetTargetingKey("d45e303a-38c2-11ed-a261-0242ac120002")

  .Set("email", "john.doe@gofeatureflag.org")

  .Build();



// Example of a boolean flag evaluation

var myFeatureFlag = await client.GetBooleanDetailsAsync("my-feature-flag", false, evaluationContext);
```

The evaluation context is the way for the client to specify contextual data that GO Feature Flag uses to evaluate the feature flags, it allows to define rules on the flag.

The `targetingKey` is mandatory for GO Feature Flag in order to evaluate the feature flag, it could be the id of a user, a session ID or anything you find relevant to use as identifier during the evaluation.

### Configure the provider[​](#configure-the-provider "Direct link to Configure the provider")

You can configure the provider with several options to customize its behavior. The following options are available:

| name                              | mandatory | Description                                                                                                                                                                                                                                                                                                                                                             |
| --------------------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`Endpoint`**                    | `true`    | endpoint contains the DNS of your GO Feature Flag relay proxy *(ex: <https://mydomain.com/gofeatureflagproxy/>)*                                                                                                                                                                                                                                                        |
| **`EvaluationType`**              | `false`   | evaluationType is the type of evaluation you want to use.- If you want to have a local evaluation, you should use InProcess.<br />- If you want to have an evaluation on the relay-proxy directly, you should use Remote.Default: InProcess<br />                                                                                                                       |
| **`Timeout`**                     | `false`   | timeout in milliseconds we are waiting when calling the relay proxy API. *(default: `10000`)*                                                                                                                                                                                                                                                                           |
| **`ApiKey`**                      | `false`   | If the relay proxy is configured to authenticate the requests, you should provide an API Key to the provider. Please ask the administrator of the relay proxy to provide an API Key. (This feature is available only if you are using GO Feature Flag relay proxy v1.7.0 or above). *(default: null)*                                                                   |
| **`FlushIntervalMs`**             | `false`   | interval time we publish statistics collection data to the proxy. The parameter is used only if the cache is enabled, otherwise the collection of the data is done directly when calling the evaluation API. *(default: `1000` ms)*                                                                                                                                     |
| **`MaxPendingEvents`**            | `false`   | max pending events aggregated before publishing for collection data to the proxy. When event is added while events collection is full, event is omitted. *(default: `10000`)*                                                                                                                                                                                           |
| **`DisableDataCollection`**       | `false`   | set to true if you don't want to collect the usage of flags retrieved in the cache. *(default: `false`)*                                                                                                                                                                                                                                                                |
| **`ExporterMetadata`**            | `false`   | exporterMetadata is the metadata we send to the GO Feature Flag relay proxy when we report the evaluation data usage.                                                                                                                                                                                                                                                   |
| **`EvaluationFlagList`**          | `false`   | If you are using in process evaluation, by default we will load in memory all the flags available in the relay proxy. If you want to limit the number of flags loaded in memory, you can use this parameter. By setting this parameter, you will only load the flags available in the list.If null or empty, all the flags available in the relay proxy will be loaded. |
| **`FlagChangePollingIntervalMs`** | `false`   | interval time we poll the proxy to check if the configuration has changed. It is used for the in process evaluation to check if we should refresh our internal cache. *(default: `120000`)*                                                                                                                                                                             |

## Evaluate your flag[​](#evaluate-your-flag "Direct link to Evaluate your flag")

This code block explain how you can create an `EvaluationContext` and use it to evaluate your flag.

note

In this example we are evaluating a `boolean` flag, but other types are also available.

**Refer to the [Open Feature documentation](https://docs.openfeature.dev/docs/reference/concepts/evaluation-api#basic-evaluation) to know more about it.**

* C#

```
// Context of your flag evaluation.

// With GO Feature Flag you MUST have a targetingKey that is a unique identifier of the user.

var userContext = EvaluationContext.Builder()

    .SetTargetingKey("1d1b9238-2591-4a47-94cf-d2bc080892f1") // user unique identifier (mandatory)

    .Set("firstname", "john")

    .Set("lastname", "doe")

    .Set("email", "john.doe@gofeatureflag.org")

    .Set("admin", true) // this field is used in the targeting rule of the flag "flag-only-for-admin"

    .Set("anonymous", false)

    .Build();



var adminFlag = await client.GetBooleanValueAsync("flag-only-for-admin", false, userContext);

if (adminFlag) {

   // flag "flag-only-for-admin" is true for the user

} else {

  // flag "flag-only-for-admin" is false for the user

}
```

## Use with .NET Aspire[​](#use-with-net-aspire "Direct link to Use with .NET Aspire")

If your application is orchestrated by [.NET Aspire](https://aspire.dev), you don't have to configure the provider by hand. The `CommunityToolkit.Aspire.GoFeatureFlag` client integration registers a `GOFeatureFlagProvider` in the dependency injection container and adds a health check for you, and the relay proxy endpoint comes straight from your AppHost, so there is no hardcoded URL in your application.

* .NET CLI
* Package Reference

```
dotnet add package CommunityToolkit.Aspire.GoFeatureFlag
```

```
<PackageReference Include="CommunityToolkit.Aspire.GoFeatureFlag" />
```

The `connectionName` must match the name of the GO Feature Flag resource declared in your AppHost.

Program.cs

```
builder.AddGoFeatureFlagClient(connectionName: "goff");
```

ExampleService.cs

```
using OpenFeature.Providers.GOFeatureFlag;



public class ExampleService(GOFeatureFlagProvider provider)

{

    // Use provider to evaluate feature flags...

}
```

If you need several relay proxies, use `AddKeyedGoFeatureFlagClient` to register one provider per connection name and resolve them with `[FromKeyedServices("...")]`.

note

To learn how to model the relay proxy in your AppHost, see the [.NET Aspire deployment page](https://gofeatureflag.org/docs/relay-proxy/deployment/aspire.md) and the upstream [Connect to GO Feature Flag](https://aspire.dev/integrations/devtools/goff/goff-connect/) reference.

## Breaking changes[​](#breaking-changes "Direct link to Breaking changes")

### 1.0.0 - Introduce In Process Evaluation[​](#100---introduce-in-process-evaluation "Direct link to 1.0.0 - Introduce In Process Evaluation")

warning

This version of the provider requires to use GO Feature Flag relay-proxy `v1.45.0` or above. If you have an older version of the relay-proxy, please use the `OpenFeature.Contrib.Providers.GOFeatureFlag` package at version `0.2.1`.

The version `1.0.0` of the provider introduces the in process evaluation. This evaluation type is more efficient than the remote evaluation because it doesn't require to call the relay-proxy API for each flag evaluation.

When configured for in process evaluation, the provider fetches flag configuration from the GO Feature Flag relay-proxy API and evaluates flags directly in the provider using a WebAssembly module compiled from the GO Feature Flag source code.

### .NET Framework Compatibility[​](#net-framework-compatibility "Direct link to .NET Framework Compatibility")

To use the GO Feature Flag provider in **In Process** mode, you need **.NET Core SDK 3.0 or later**. For older .NET Framework versions, only remote evaluation mode is available.

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

You can find the source of the provider in the [`open-feature/dotnet-sdk-contrib`](https://github.com/open-feature/dotnet-sdk-contrib/tree/main/src/OpenFeature.Providers.GOFeatureFlag) repository.
