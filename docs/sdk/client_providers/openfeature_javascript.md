# Javascript / Typescript SDK usage

[![NPM Version](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue\&style=flat-square\&logo=npm)](https://www.npmjs.com/package/@openfeature/go-feature-flag-web-provider) ![NPM Downloads](https://img.shields.io/npm/d18m/%40openfeature%2Fgo-feature-flag-web-provider?style=flat-square)

This page describes how to use the OpenFeature Javascript web SDK for your client application.

## About this provider[​](#about-this-provider "Direct link to About this provider")

[GO Feature Flag](https://gofeatureflag.org) provider allows you to connect to your GO Feature Flag instance with the `@openfeature/web-sdk`.

The main difference between this provider and [`@openfeature/go-feature-flag-provider`](https://www.npmjs.com/package/@openfeature/go-feature-flag-provider) is that it uses a **static evaluation context**. This provider is more sustainable for client-side implementation.

If you want to know more about this pattern, I encourage you to read this [blog post](https://openfeature.dev/blog/catering-to-the-client-side/).

## Install the provider[​](#install-the-provider "Direct link to Install the provider")

```
npm install @openfeature/go-feature-flag-web-provider @openfeature/web-sdk
```

## How to use the provider?[​](#how-to-use-the-provider "Direct link to How to use the provider?")

```
const evaluationCtx: EvaluationContext = {

  targetingKey: 'user-key',

  email: 'john.doe@gofeatureflag.org',

  name: 'John Doe',

};



const goFeatureFlagWebProvider = new GoFeatureFlagWebProvider({

  endpoint: endpoint,

  // ...

}, logger);



await OpenFeature.setContext(evaluationCtx); // Set the static context for OpenFeature

OpenFeature.setProvider(goFeatureFlagWebProvider); // Attach the provider to OpenFeature

const client = await OpenFeature.getClient();



// You can now use the client to use your flags

if(client.getBooleanValue('my-new-feature', false)){

    //...

}



// You can add handlers to know what happen in the provider

client.addHandler(ProviderEvents.Ready, () => { ... });

client.addHandler(ProviderEvents.Error, () => { //... });

client.addHandler(ProviderEvents.Stale, () => { //... });

client.addHandler(ProviderEvents.ConfigurationChanged, () => { //... });
```

### Limit the flags to evaluate[​](#limit-the-flags-to-evaluate "Direct link to Limit the flags to evaluate")

By default, the provider will fetch all the flags configured in the GO Feature Flag server to be ready to evaluate them. If you know in advance, what are the flags you will evaluate in your application, you can specify the list of flags to evaluate in the context.

You need to add in the evaluation context the restricted key `gofeatureflag.flagList` with the list of flags you want to evaluate.

```
OpenFeature.setContext({

  // ...

  gofeatureflag: {

    flagList: ['flag1', 'flag2']

  }

});



await OpenFeature.setContext(evaluationCtx);
```

By setting the `gofeatureflag.flagList` key in the context, the provider will only fetch the flags specified in the list.

warning

When limiting the flags to evaluate, if you try to evaluate a flag not in the list, the provider will return the default value with the error `FLAG_NOT_FOUND`.

### Available options[​](#available-options "Direct link to Available options")

| Option name                   | Type   | Default        | Description                                                                                                                                                                                     |
| ----------------------------- | ------ | -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| endpoint                      | string |                | endpoint is the URL where your GO Feature Flag server is located.                                                                                                                               |
| apiTimeout                    | number | 0 = no timeout | (optional) timeout is the time in millisecond we wait for an answer from the server.                                                                                                            |
| apiKey                        | string |                | (optional) If GO Feature Flag is configured to authenticate the requests, you should provide an API Key to the provider. Please ask the administrator of the relay proxy to provide an API Key. |
| websocketRetryInitialDelay    | number | 100            | (optional) initial delay in millisecond to wait before retrying to connect the websocket                                                                                                        |
| websocketRetryDelayMultiplier | number | 2              | (optional) multiplier of websocketRetryInitialDelay after each failure *(example: 1st connection retry will be after 100ms, second after 200ms, third after 400ms ...)*                         |
| websocketMaxRetries           | number | 10             | (optional) maximum number of retries before considering the websocket unreachable                                                                                                               |
| customHeaders                 | object |                | (optional) custom HTTP headers to add to every request sent to the relay proxy *(e.g. `{"X-Custom-Header": "value"}`)*                                                                          |

### Reconnection[​](#reconnection "Direct link to Reconnection")

If the connection to the GO Feature Flag instance fails, the provider will attempt to reconnect with an exponential back-off. The `websocketMaxRetries` can be specified to customize reconnect behavior.

### Event streaming[​](#event-streaming "Direct link to Event streaming")

The `GoFeatureFlagWebProvider` receives events from GO Feature Flag with changes. Combined with the event API in the web SDK, this allows for subscription to flag value changes in clients.

```
client.addHandler(ProviderEvents.ConfigurationChanged, (ctx: EventDetails) => {

  // do something when the configuration has changed.

  // ctx.flagsChanged contains the list of changed flags.

});
```

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
