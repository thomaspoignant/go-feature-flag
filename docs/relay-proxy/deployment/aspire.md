# .NET Aspire

## Overview[​](#overview "Direct link to Overview")

[.NET Aspire](https://aspire.dev) is a code-first way to describe the services, containers and dependencies of your application in a single place.

The **Aspire GO Feature Flag integration** lets you model the relay proxy as a resource in your AppHost, which means:

* The `gofeatureflag/go-feature-flag` container is started for you, no `docker run` and no hardcoded URL.
* The relay proxy endpoint is injected into every app that references it as environment variables, so it works from C#, TypeScript, Python, Go or any other language.
* A health check on the [`/health`](https://gofeatureflag.org/docs/relay-proxy/relay_proxy_endpoints.md) endpoint is registered automatically, so the Aspire dashboard knows when the relay proxy is ready.
* Logs, traces and metrics of the relay proxy are surfaced in the Aspire dashboard through the OpenTelemetry exporter.

info

The Aspire integration is maintained by the **Aspire Community Toolkit**, not by the GO Feature Flag team. The full reference lives on [aspire.dev](https://aspire.dev/integrations/devtools/goff/goff-get-started/).

note

This integration targets your **local development inner loop**. To run the relay proxy in production, have a look at [Helm](https://gofeatureflag.org/docs/relay-proxy/deployment/helm.md), [AWS Lambda](https://gofeatureflag.org/docs/relay-proxy/deployment/aws-lambda.md) or start the [container](https://gofeatureflag.org/docs/relay-proxy/install_relay_proxy.md) directly.

## Prerequisites[​](#prerequisites "Direct link to Prerequisites")

* The [Aspire CLI](https://aspire.dev/get-started/install-cli/) installed on your machine.
* A container runtime such as Docker or Podman.

## Step 1: Add the hosting integration[​](#step-1-add-the-hosting-integration "Direct link to Step 1: Add the hosting integration")

Add the hosting integration to your AppHost project:

```
aspire add CommunityToolkit.Aspire.Hosting.GoFeatureFlag
```

## Step 2: Declare the relay proxy resource[​](#step-2-declare-the-relay-proxy-resource "Direct link to Step 2: Declare the relay proxy resource")

Declare a GO Feature Flag resource and reference it from the apps that need to evaluate flags.

* C#
* TypeScript

AppHost.cs

```
var builder = DistributedApplication.CreateBuilder(args);



var goff = builder.AddGoFeatureFlag("goff", pathToConfigFile: "/goff/goff-proxy.yaml")

    .WithGoffBindMount("./goff");



var apiService = builder.AddProject<Projects.ExampleProject>("apiservice")

    .WithReference(goff);



builder.Build().Run();
```

apphost.mts

```
import { createBuilder } from './.aspire/modules/aspire.mjs';



const builder = await createBuilder();



const goff = await builder.addGoFeatureFlag('goff', {

  pathToConfigFile: '/goff/goff-proxy.yaml',

});

await goff.withGoffBindMount('./goff');



const api = await builder.addProject(

  'apiservice',

  '../ExampleProject/ExampleProject.csproj'

);

await api.withReference(goff);



await builder.build().run();
```

* `WithGoffBindMount` mounts a local folder into the `/goff` path inside the container, this is where the relay proxy reads its configuration from.
* `pathToConfigFile` is passed to the relay proxy as its `--config` argument.
* `WithReference` injects the relay proxy connection information into the consuming app.

Data volumes, log levels and fixed host ports are described in the [hosting integration reference](https://aspire.dev/integrations/devtools/goff/goff-host/).

## Step 3: Provide the configuration[​](#step-3-provide-the-configuration "Direct link to Step 3: Provide the configuration")

The bind-mounted folder must contain the relay proxy configuration and your flag definitions.

goff/goff-proxy.yaml

```
retrievers:

  - kind: file

    path: /goff/flags.yaml
```

goff/flags.yaml

```
display-banner:

  variations:

    enabled: true

    disabled: false

  defaultRule:

    variation: enabled
```

All the available relay proxy options are listed in [Configure the relay proxy](https://gofeatureflag.org/docs/relay-proxy/configure-relay-proxy.md), and the flag format is described in [Create your flags](https://gofeatureflag.org/docs/configure_flag/create-flags.md).

## Step 4: Connect your application[​](#step-4-connect-your-application "Direct link to Step 4: Connect your application")

Aspire exposes the connection information of a resource as `[RESOURCE]_[PROPERTY]` environment variables. For a resource named `goff` you get:

| Environment variable      | Description                                                     |
| ------------------------- | --------------------------------------------------------------- |
| `GOFF_URI`                | Base URL of the relay proxy (`http://{host}:{port}`)            |
| `GOFF_HOST`               | Hostname or IP address of the relay proxy                       |
| `GOFF_PORT`               | Port the relay proxy is listening on                            |
| `ConnectionStrings__goff` | Connection string, formatted as `Endpoint=http://{host}:{port}` |

In a .NET application, the [`CommunityToolkit.Aspire.GoFeatureFlag`](https://www.nuget.org/packages/CommunityToolkit.Aspire.GoFeatureFlag) client integration reads it for you and registers the provider in the dependency injection container:

Program.cs

```
builder.AddGoFeatureFlagClient(connectionName: "goff");
```

In any other language, read `GOFF_URI` and use it as the endpoint of the [OpenFeature provider](https://gofeatureflag.org/docs/sdk.md) of your choice. Note that the providers expect a trailing `/`, so append it to the value of `GOFF_URI`.

## Going further[​](#going-further "Direct link to Going further")

* [Get started with GO Feature Flag integrations](https://aspire.dev/integrations/devtools/goff/goff-get-started/)
* [Set up GO Feature Flag in the AppHost](https://aspire.dev/integrations/devtools/goff/goff-host/)
* [Connect to GO Feature Flag](https://aspire.dev/integrations/devtools/goff/goff-connect/)
