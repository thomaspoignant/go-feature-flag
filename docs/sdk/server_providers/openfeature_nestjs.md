# NestJS

[![NPM Version](https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-provider?color=blue\&style=flat-square)](https://www.npmjs.com/package/@openfeature/go-feature-flag-provider) ![NPM Downloads](https://img.shields.io/npm/d18m/%40openfeature%2Fgo-feature-flag-provider?style=flat-square)

The OpenFeature `NestJS` SDK is a package that provides a NestJS wrapper for the OpenFeature [`Node Server SDK`](https://gofeatureflag.org/docs/sdk/server_providers/openfeature_javascript.md).

info

We don't have a specific provider for NestJS, you have to use the `GO Feature Flag node provider` with the OpenFeature SDK.

We recommend you to check the [OpenFeature NestJS SDK Documentation](https://openfeature.dev/docs/reference/technologies/server/javascript/nestjs) to see all the available directives and how to use them.

## Install dependencies[​](#install-dependencies "Direct link to Install dependencies")

The first things we will do is install the **Open Feature SDK** and the **GO Feature Flag provider**.

* yarn
* npm

```
# yarn requires manual installation of the peer dependencies (see below) 

yarn add @openfeature/nestjs-sdk \

    @openfeature/server-sdk \

    @openfeature/core \

    @openfeature/go-feature-flag-provider
```

```
npm i @openfeature/nestjs-sdk @openfeature/go-feature-flag-provider
```

## Initialize your Open Feature client[​](#initialize-your-open-feature-client "Direct link to Initialize your Open Feature client")

To evaluate the flags you need to have an Open Feature configured in your app. This code block shows you how you can create a client that you can use in your application.

```
import { Module, OnModuleInit } from '@nestjs/common';

import { OpenFeatureModule } from '@openfeature/nestjs-sdk';

import { GoFeatureFlagProvider } from  '@openfeature/go-feature-flag-provider';

import { OpenFeature } from '@openfeature/server-sdk';



const goffProvider = new GoFeatureFlagProvider({

  endpoint: 'http://localhost:1031/', // DNS of your instance of relay proxy

});



@Module({

  imports: [

    ...

    OpenFeatureModule.forRoot({

      contextFactory: () => ({

        targetingKey: 'nestjs-test'

      }),

    }),

  ],

  controllers: [...],

  providers: [...],

})

export class AppModule implements OnModuleInit {

  async onModuleInit() {

    await OpenFeature.setProviderAndWait(goffProvider);

  }

}
```

## Evaluate your flag[​](#evaluate-your-flag "Direct link to Evaluate your flag")

To use the OpenFeature NestJS SDK in a service first inject the OpenFeatureClient into the service.<br /><!-- -->Then use a flag value by creating the EvaluationContext, and pass the flag key, default value, and EvaluationContext to one of the OpenFeature flag evaluation methods.

```
import { Injectable, Inject } from '@nestjs/common';

import { OpenFeatureClient, Client } from '@openfeature/nestjs-sdk';



const EVALUATION_CONTEXT = { targetingKey: '4b9e3d6e-cb2a-40fb-8834-5352c59f04c1' };



@Injectable()

export class ExampleService {

  constructor(

    @OpenFeatureClient() private ofClient: Client,

  ) {}



  async testFlag() {

    const testFlag = await this.ofClient.getBooleanValue(

      'test-flag',

      false,

      EVALUATION_CONTEXT,

    );

  }

}
```

## Required TargetingKey[​](#required-targetingkey "Direct link to Required TargetingKey")

GO Feature Flag needs a `targetingKey` to be set on the evaluation context.<br /><!-- -->This is used to bucket the users into the correct targeting group for the feature flag evaluation.

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
