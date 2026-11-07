# React SDK usage

[![NPM Version](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue\&style=flat-square\&logo=npm)](https://www.npmjs.com/package/@openfeature/go-feature-flag-web-provider) ![NPM Downloads](https://img.shields.io/npm/d18m/%40openfeature%2Fgo-feature-flag-web-provider?style=flat-square)

This page describes how to use the OpenFeature React SDK for your client application.

## About this provider[​](#about-this-provider "Direct link to About this provider")

OpenFeature React SDK allows you to connect to your GO Feature Flag instance with the `@openfeature/react-sdk`.

If you work with React the `@openfeature/react-sdk` will give you a better integration with your React application. To integrate it with GO Feature Flag, you need to use the `@openfeature/go-feature-flag-web-provider` provider.

## Install the provider[​](#install-the-provider "Direct link to Install the provider")

```
npm install @openfeature/go-feature-flag-web-provider

npm install @openfeature/web-sdk

npm install @openfeature/react-sdk

npm install @openfeature/core
```

## How to use the provider?[​](#how-to-use-the-provider "Direct link to How to use the provider?")

### OpenFeatureProvider context provider[​](#openfeatureprovider-context-provider "Direct link to OpenFeatureProvider context provider")

The OpenFeatureProvider is a React context provider which represents a scope for feature flag evaluations within a React application. It binds an OpenFeature client to all evaluations within child components, and allows the use of evaluation hooks.

```
import { EvaluationContext, OpenFeature, OpenFeatureProvider, useFlag } from "@openfeature/react-sdk";

import { GoFeatureFlagWebProvider } from "@openfeature/go-feature-flag-web-provider";



const goFeatureFlagWebProvider = new GoFeatureFlagWebProvider({

  endpoint: "http://localhost:1031"

});



// Set the initial context for your evaluations

OpenFeature.setContext({

  targetingKey: "user-1",

  admin: false

});



// Instantiate and set our provider (be sure this only happens once)!

// Note: there's no need to await its initialization, the React SDK handles re-rendering and suspense for you!

OpenFeature.setProvider(goFeatureFlagWebProvider);



// Enclose your content in the configured provider

function App() {

  return (

    <OpenFeatureProvider>

      <Page />

    </OpenFeatureProvider>

  );

}
```

### Evaluation hooks[​](#evaluation-hooks "Direct link to Evaluation hooks")

Within the provider, you can use the various evaluation hooks to evaluate flags.

```
function Page() {

  // Use the "query-style" flag evaluation hook, specifying a flag-key and a default value.

  const { value: showNewMessage } = useFlag('new-message', true);

  return (

    <div className="App">

      <header className="App-header">

        {showNewMessage ? <p>Welcome to this OpenFeature-enabled React app!</p> : <p>Welcome to this React app.</p>}

      </header>

    </div>

  )

}
```

### Advanced usage[​](#advanced-usage "Direct link to Advanced usage")

You can check the [OpenFeature React SDK documentation](https://openfeature.dev/docs/reference/technologies/client/web/react) to see all the available hooks and how to use them.

## Available options[​](#available-options "Direct link to Available options")

Check the available options for the provider in the [web provider page](https://gofeatureflag.org/docs/sdk/client_providers/openfeature_javascript.md).

## Example[​](#example "Direct link to Example")

If you want to see some code, you can check the [example](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples/openfeature_react) in the GO Feature Flag repository.

## Features status[​](#features-status "Direct link to Features status")

| Status | Feature                      | Description                                                                                                                                                                |
| ------ | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **     | Remote Evaluation            | The provider is calling the remote server to evaluate the feature flags.                                                                                                   |
| **     | Tracking Flag Evaluation     | The provider is tracking all the evaluations of your feature flags and you can export them using an exporter.                                                              |
| **     | Tracking Custom Events       | The provider is tracking custom events through the track() function of your SDK. All those events are send to the exporter for you to forward them where you want.         |
| **     | Configuration Change Updates | The provider is able to update the configuration based on the configuration, it means that the provider is able to react to any feature flag change on your configuration. |
| **     | Provider Events Reactions    | You can add an event handler to the provider to react to the provider events.                                                                                              |

**Implemented**In-progress**Not implemented yet

## Contribute to the provider[​](#contribute-to-the-provider "Direct link to Contribute to the provider")

You can find the source of the provider in the [`open-feature/js-sdk-contrib`](https://github.com/open-feature/js-sdk-contrib/tree/main/libs/providers/go-feature-flag-web) repository.
