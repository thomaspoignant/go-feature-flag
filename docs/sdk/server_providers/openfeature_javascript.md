# Node.js

[![NPM Version](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-provider?color=blue\&style=flat-square)](https://www.npmjs.com/package/@openfeature/go-feature-flag-provider) ![NPM Downloads](https://img.shields.io/npm/d18m/%40openfeature%2Fgo-feature-flag-provider?style=flat-square)

This provider supports both **in-process** and **remote** evaluation modes, offering flexibility for different deployment scenarios.

## Install dependencies[​](#install-dependencies "Direct link to Install dependencies")

The first things we will do is install the **Open Feature SDK** and the **GO Feature Flag provider**.

* yarn
* npm

```
yarn add @openfeature/server-sdk @openfeature/go-feature-flag-provider
```

```
npm i @openfeature/server-sdk @openfeature/go-feature-flag-provider
```

## Initialize your Open Feature client[​](#initialize-your-open-feature-client "Direct link to Initialize your Open Feature client")

To evaluate the flags you need to have an Open Feature configured in your app. This code block shows you how you can create a client that you can use in your application.

* JavaScript
* TypeScript

```
const { OpenFeature } = require("@openfeature/server-sdk");

const { GoFeatureFlagProvider, EvaluationType } = require("@openfeature/go-feature-flag-provider");



// Initialize the provider

const goFeatureFlagProvider = new GoFeatureFlagProvider({

  endpoint: 'http://localhost:1031/', // DNS of your instance of relay proxy

  evaluationType: EvaluationType.Remote,

});



// Register the provider

OpenFeature.setProvider(goFeatureFlagProvider);



// Get a client

const featureFlagClient = OpenFeature.getClient('my-app');
```

```
import { OpenFeature } from "@openfeature/server-sdk";

import { GoFeatureFlagProvider, EvaluationType } from "@openfeature/go-feature-flag-provider";



// Initialize the provider

const goFeatureFlagProvider = new GoFeatureFlagProvider({

  endpoint: 'http://localhost:1031/', // DNS of your instance of relay proxy

  evaluationType: EvaluationType.Remote,

});



// Register the provider

OpenFeature.setProvider(goFeatureFlagProvider);



// Get a client

const featureFlagClient = OpenFeature.getClient('my-app');
```

### In-Process Evaluation[​](#in-process-evaluation "Direct link to In-Process Evaluation")

For high-performance scenarios where you want to evaluate flags locally:

* JavaScript
* TypeScript

```
const { GoFeatureFlagProvider, EvaluationType } = require("@openfeature/go-feature-flag-provider");



const goFeatureFlagProvider = new GoFeatureFlagProvider({

  endpoint: 'http://localhost:1031/',

  evaluationType: EvaluationType.InProcess,

  flagChangePollingIntervalMs: 30000, // Poll every 30 seconds

});
```

```
import { GoFeatureFlagProvider, EvaluationType } from "@openfeature/go-feature-flag-provider";



const goFeatureFlagProvider = new GoFeatureFlagProvider({

  endpoint: 'http://localhost:1031/',

  evaluationType: EvaluationType.InProcess,

  flagChangePollingIntervalMs: 30000, // Poll every 30 seconds

});
```

## Configuration options[​](#configuration-options "Direct link to Configuration options")

| Option                        | Type               | Default      | Description                                     |
| ----------------------------- | ------------------ | ------------ | ----------------------------------------------- |
| `endpoint`                    | `string`           | **Required** | The endpoint of the GO Feature Flag relay-proxy |
| `evaluationType`              | `EvaluationType`   | `InProcess`  | Evaluation mode: `InProcess` or `Remote`        |
| `timeout`                     | `number`           | `10000`      | HTTP request timeout in milliseconds            |
| `flagChangePollingIntervalMs` | `number`           | `120000`     | Polling interval for configuration changes      |
| `dataFlushInterval`           | `number`           | `120000`     | Data collection flush interval                  |
| `maxPendingEvents`            | `number`           | `10000`      | Maximum pending events before flushing          |
| `disableDataCollection`       | `boolean`          | `false`      | Disable data collection entirely                |
| `apiKey`                      | `string`           | `undefined`  | API key for authentication                      |
| `exporterMetadata`            | `ExporterMetadata` | `undefined`  | Custom metadata for events                      |
| `fetchImplementation`         | `FetchAPI`         | `undefined`  | Custom fetch implementation                     |

### Evaluation types[​](#evaluation-types "Direct link to Evaluation types")

#### InProcess Evaluation[​](#inprocess-evaluation "Direct link to InProcess Evaluation")

* **Performance**: Fastest evaluation with local caching
* **Network**: Minimal network calls, only for configuration updates and tracking
* **Use Case**: High-performance applications, real-time evaluation

#### Remote Evaluation[​](#remote-evaluation "Direct link to Remote Evaluation")

* **Performance**: Network-dependent evaluation
* **Network**: Each evaluation requires a network call, works well with side-cars or in the edge
* **Use Case**: Centralized control

## Evaluate your flag[​](#evaluate-your-flag "Direct link to Evaluate your flag")

This code block explains how you can create an `EvaluationContext` and use it to evaluate your flag.

note

In this example, we are evaluating a `boolean` flag, but other types are also available.

**Refer to the [Open Feature documentation](https://docs.openfeature.dev/docs/reference/concepts/evaluation-api#basic-evaluation) to know more about it.**

* JavaScript
* TypeScript

```
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

} else {

  // flag "flag-only-for-admin" is false for the user

}
```

```
import { EvaluationContext } from "@openfeature/server-sdk";



// Context of your flag evaluation.

// With GO Feature Flag you MUST have a targetingKey that is a unique identifier of the user.

const userContext: EvaluationContext = {

  targetingKey: '1d1b9238-2591-4a47-94cf-d2bc080892f1', // user unique identifier

  firstname: 'john',

  lastname: 'doe',

  email: 'john.doe@gofeatureflag.org',

  admin: true, // this field is used in the targeting rule of the flag "flag-only-for-admin"

  // ...

};



const adminFlag = await featureFlagClient.getBooleanValue('flag-only-for-admin', false, userContext);

if (adminFlag) {

  // flag "flag-only-for-admin" is true for the user

} else {

  // flag "flag-only-for-admin" is false for the user

}
```

## Advanced usage[​](#advanced-usage "Direct link to Advanced usage")

### Custom context and targeting[​](#custom-context-and-targeting "Direct link to Custom context and targeting")

```
const context = {

  targetingKey: 'user-123',

  email: 'john.doe@example.com',

  firstname: 'John',

  lastname: 'Doe',

  anonymous: false,

  professional: true,

  rate: 3.14,

  age: 30,

  company_info: {

    name: 'my_company',

    size: 120,

  },

  labels: ['pro', 'beta'],

};



const flagValue = await featureFlagClient.getBooleanValue('my-feature-flag', false, context);
```

### Data collection and analytics[​](#data-collection-and-analytics "Direct link to Data collection and analytics")

The provider automatically collects evaluation data. You can customize this behavior:

```
const goFeatureFlagProvider = new GoFeatureFlagProvider({

  endpoint: 'http://localhost:1031/',

  evaluationType: EvaluationType.Remote,

  disableDataCollection: false, // Enable data collection

  dataFlushInterval: 20000, // Flush every 20 seconds

  maxPendingEvents: 5000, // Max 5000 pending events

});
```

### Custom exporter metadata[​](#custom-exporter-metadata "Direct link to Custom exporter metadata")

Add custom metadata to your evaluation events:

```
import { ExporterMetadata } from '@openfeature/go-feature-flag-provider';



const metadata = new ExporterMetadata()

  .add('environment', 'production')

  .add('version', '1.0.0')

  .add('region', 'us-east-1');



const goFeatureFlagProvider = new GoFeatureFlagProvider({

  endpoint: 'http://localhost:1031/',

  evaluationType: EvaluationType.Remote,

  exporterMetadata: metadata,

});
```

## Flag types supported[​](#flag-types-supported "Direct link to Flag types supported")

The provider supports all OpenFeature flag types:

### Boolean flags[​](#boolean-flags "Direct link to Boolean flags")

```
const isEnabled = await featureFlagClient.getBooleanValue('feature-flag', false, context);

const details = await featureFlagClient.getBooleanDetails('feature-flag', false, context);
```

### String flags[​](#string-flags "Direct link to String flags")

```
const message = await featureFlagClient.getStringValue('welcome-message', 'Hello!', context);

const details = await featureFlagClient.getStringDetails('welcome-message', 'Hello!', context);
```

### Number flags[​](#number-flags "Direct link to Number flags")

```
const percentage = await featureFlagClient.getNumberValue('discount-percentage', 0, context);

const details = await featureFlagClient.getNumberDetails('discount-percentage', 0, context);
```

### Object flags[​](#object-flags "Direct link to Object flags")

```
const config = await featureFlagClient.getObjectValue('user-config', {}, context);

const details = await featureFlagClient.getObjectDetails('user-config', {}, context);
```

## Tracking events[​](#tracking-events "Direct link to Tracking events")

The provider supports custom event tracking:

```
// Track a custom event

featureFlagClient.track('user_action', context, {

  action: 'button_click',

  page: 'homepage',

  timestamp: Date.now(),

});
```

## Next.js[​](#nextjs "Direct link to Next.js")

### Vercel Flags SDK[​](#vercel-flags-sdk "Direct link to Vercel Flags SDK")

The [Flags SDK](https://flags-sdk.dev) is Vercel's toolkit to declare and use feature flags in Next.js. It works with GO Feature Flag through its **OpenFeature adapter**, which wraps any OpenFeature client, including GOFF.

* yarn
* npm

```
yarn add flags @flags-sdk/openfeature @openfeature/server-sdk @openfeature/go-feature-flag-provider
```

```
npm i flags @flags-sdk/openfeature @openfeature/server-sdk @openfeature/go-feature-flag-provider
```

Create the adapter using the **asynchronous** form, since the GO Feature Flag provider initializes asynchronously:

```
import { createOpenFeatureAdapter } from "@flags-sdk/openfeature";

import { OpenFeature } from "@openfeature/server-sdk";

import { GoFeatureFlagProvider } from "@openfeature/go-feature-flag-provider";



export const openFeatureAdapter = createOpenFeatureAdapter(async () => {

  await OpenFeature.setProviderAndWait(

    new GoFeatureFlagProvider({

      endpoint: "http://localhost:1031/", // DNS of your instance of relay proxy

    }),

  );

  return OpenFeature.getClient();

});
```

You can now declare your flags and use them anywhere in your Next.js application:

```
import { flag } from "flags/next";

import type { EvaluationContext } from "@openfeature/server-sdk";



export const myFeatureFlag = flag<boolean, EvaluationContext>({

  key: "my-feature-flag",

  defaultValue: false,

  adapter: openFeatureAdapter.booleanValue(),

});
```

The adapter also exposes `stringValue()`, `numberValue()` and `objectValue()`, along with an `identify` function to build your evaluation context. Check the [Flags SDK OpenFeature documentation](https://flags-sdk.dev/providers/openfeature) for the full API and the [GO Feature Flag integration page](https://flags-sdk.dev/docs/providers/openfeature/go).

### Loading the WASM binary[​](#loading-the-wasm-binary "Direct link to Loading the WASM binary")

warning

If you're using NextJS with in-process evaluation mode, you might encounter issues loading the WASM binary.

In the server side, you might have an error like this *(only if you are using the in-process evaluation mode)*:

```
WasmNotLoadedException: Failed to load WASM binary: WASM file not found
```

Because of the specificity of **NextJS**, you need to add the following code to your `next.config.ts` file:

```
const nextConfig: NextConfig = {

  // ...

  serverExternalPackages: ["@openfeature/go-feature-flag-provider"],

  // ...

};
```

This will tell NextJS to use the `@openfeature/go-feature-flag-provider` package as an external package and not to bundle it with the application, which is the best way to avoid the WASM binary not found error.

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

You can find the source of the provider in the [`open-feature/js-sdk-contrib`](https://github.com/open-feature/js-sdk-contrib/tree/main/libs/providers/go-feature-flag) repository.
