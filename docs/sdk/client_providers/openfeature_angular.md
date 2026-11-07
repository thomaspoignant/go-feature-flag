# Angular SDK

[![NPM Version](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue\&style=flat-square\&logo=npm)](https://www.npmjs.com/package/@openfeature/go-feature-flag-web-provider) ![NPM Downloads](https://img.shields.io/npm/d18m/%40openfeature%2Fgo-feature-flag-web-provider?style=flat-square)

This page describes how to use the OpenFeature Angular SDK for your client application.

## About this provider[​](#about-this-provider "Direct link to About this provider")

OpenFeature Angular SDK allows you to connect to your GO Feature Flag instance with the `@openfeature/angular-sdk`.

If you work with Angular the `@openfeature/angular-sdk` will give you a better integration with your Angular application. To integrate it with GO Feature Flag, you need to use the `@openfeature/go-feature-flag-web-provider` provider.

info

We recommend you to check the [OpenFeature Angular SDK Documentation](https://openfeature.dev/docs/reference/technologies/client/web/angular) to see all the available directives and how to use them.

## Install the provider[​](#install-the-provider "Direct link to Install the provider")

```
npm install @openfeature/go-feature-flag-web-provider

npm install @openfeature/angular-sdk
```

## How to use the provider?[​](#how-to-use-the-provider "Direct link to How to use the provider?")

### Module[​](#module "Direct link to Module")

To include the OpenFeature Angular directives in your application, you need to import the `OpenFeatureModule` and configure it using the `forRoot` method.

```
import { NgModule } from '@angular/core';

import { CommonModule } from '@angular/common';

import { OpenFeatureModule } from '@openfeature/angular-sdk';

import { GoFeatureFlagWebProvider } from '@openfeature/go-feature-flag-web-provider';



@NgModule({

  declarations: [

    // Other components

  ],

  imports: [

    CommonModule,

    OpenFeatureModule.forRoot({

      provider: new GoFeatureFlagWebProvider({

        endpoint: 'http://localhost:1031'

      }),

    })

  ],

})

export class AppModule {

}
```

### Minimal Example[​](#minimal-example "Direct link to Minimal Example")

You don't need to provide all the templates. Here's a minimal example using a boolean feature flag:

If initializing and reconciling are not given, the feature flag value that is returned by the provider will determine what will be rendered.

```
<div *booleanFeatureFlag="'isFeatureEnabled'; default: true">

    This is shown when the feature flag is enabled.

</div>
```

This example shows content when the feature flag `isFeatureEnabled` is `true` with a default value of `true`. No else, initializing, or reconciling templates are required in this case.

### Available directives per types[​](#available-directives-per-types "Direct link to Available directives per types")

The OpenFeature SDK provides a set of directives `booleanFeatureFlag`, `numberFeatureFlag`, `stringFeatureFlag` and `objectFeatureFlag` to get the value of a feature flag. See the [OpenFeature Angular SDK Documentation](https://openfeature.dev/docs/reference/technologies/client/web/angular#how-to-use) for more details on how to use this directive, including how to use `<type>FeatureElse`, `<type>FeatureInitializing`, and `<type>FeatureReconciling`.

* Boolean
* Number
* String
* Object

```
<div

  *booleanFeatureFlag="'isFeatureEnabled'; default: true; domain: 'userDomain'; else: booleanFeatureElse; initializing: booleanFeatureInitializing; reconciling: booleanFeatureReconciling">

  This is shown when the feature flag is enabled.

</div>

<ng-template #booleanFeatureElse>

  This is shown when the feature flag is disabled.

</ng-template>

<ng-template #booleanFeatureInitializing>

  This is shown when the feature flag is initializing.

</ng-template>

<ng-template #booleanFeatureReconciling>

  This is shown when the feature flag is reconciling.

</ng-template>
```

```
<div

  *numberFeatureFlag="'discountRate'; value: 10; default: 5; domain: 'userDomain'; else: numberFeatureElse; initializing: numberFeatureInitializing; reconciling: numberFeatureReconciling">

  This is shown when the feature flag matches the specified discount rate.

</div>

<ng-template #numberFeatureElse>

  This is shown when the feature flag does not match the specified discount rate.

</ng-template>

<ng-template #numberFeatureInitializing>

  This is shown when the feature flag is initializing.

</ng-template>

<ng-template #numberFeatureReconciling>

  This is shown when the feature flag is reconciling.

</ng-template>
```

```
<div

  *stringFeatureFlag="'themeColor'; value: 'dark'; default: 'light'; domain: 'userDomain'; else: stringFeatureElse; initializing: stringFeatureInitializing; reconciling: stringFeatureReconciling">

  This is shown when the feature flag matches the specified theme color.

</div>

<ng-template #stringFeatureElse>

  This is shown when the feature flag does not match the specified theme color.

</ng-template>

<ng-template #stringFeatureInitializing>

  This is shown when the feature flag is initializing.

</ng-template>

<ng-template #stringFeatureReconciling>

  This is shown when the feature flag is reconciling.

</ng-template>
```

```
<div

  *objectFeatureFlag="'userConfig'; value: { theme: 'dark' }; default: { theme: 'light' }; domain: 'userDomain'; else: objectFeatureElse; initializing: objectFeatureInitializing; reconciling: objectFeatureReconciling">

  This is shown when the feature flag matches the specified user configuration.

</div>

<ng-template #objectFeatureElse>

  This is shown when the feature flag does not match the specified user configuration.

</ng-template>

<ng-template #objectFeatureInitializing>

  This is shown when the feature flag is initializing.

</ng-template>

<ng-template #objectFeatureReconciling>

  This is shown when the feature flag is reconciling.

</ng-template>
```

## Required TargetingKey[​](#required-targetingkey "Direct link to Required TargetingKey")

GO Feature Flag needs a `targetingKey` to be set on the evaluation context.<br /><!-- -->This is used to bucket the users into the correct targeting group for the feature flag evaluation.

Don't forget to set a context with a `targetingKey` in your application.

```
await OpenFeature.setContext({

  targetingKey: 'user-key',

  // ...

});
```

### Advanced usage[​](#advanced-usage "Direct link to Advanced usage")

You can check the [OpenFeature Angular SDK Documentation](https://openfeature.dev/docs/reference/technologies/client/web/angular) to see all the available directives and how to use them.

## Available options[​](#available-options "Direct link to Available options")

Check the available options for the provider in the [web provider page](https://gofeatureflag.org/docs/sdk/client_providers/openfeature_javascript.md).

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
