# 🚥 Flag Evaluation

## Overview[​](#overview "Direct link to Overview")

In a feature flag system, flag evaluation is the process of determining the current state *(which variant to use)* of a feature flag for a given [evaluation context](https://gofeatureflag.org/docs/concepts/evaluation-context.md).

This process happens at runtime, allowing you to dynamically control the behavior of your application without redeploying code.

## What Happens During Evaluation?[​](#what-happens-during-evaluation "Direct link to What Happens During Evaluation?")

The evaluation process typically involves these steps:

1. **Retrieving the Flag Configuration:** The system fetches the configuration for the specific flag being evaluated. This configuration includes information such as the flag's default value, targeting rules, and any other relevant settings.

2. **Contextual Input:** The evaluation receives context about the current request or user. This context can include information like:

   <!-- -->

   * A targeting key
   * Device type
   * Location
   * Custom attributes

3. **Rule Evaluation:** The system evaluates the targeting rules defined for the flag against the provided context. These rules determine which users or requests should see the feature enabled (or a specific variant). Common targeting strategies include:

   <!-- -->

   * **Percentage rollout:** Enable the feature for a percentage of users *(e.g., 50%)*.
   * **User targeting:** Enable the feature for specific users.
   * **Attribute-based targeting:** Enable the feature based on user attributes *(e.g., users with a "premium" subscription)*.

4. **Returning the Flag Value:** Based on the rule evaluation, the system returns the appropriate value for the flag (e.g., `true`, `false`, or a variant value like `"variantA"`).

## Supported type of flags[​](#supported-type-of-flags "Direct link to Supported type of flags")

GO Feature Flag supports a variety of flag types, including:

* **Boolean**
* **String**
* **Integer**
* **Double**
* **Object**

For each flag type, the SDKs are providing specific APIs to be able to retrieve your flag in the right type inside your code.

## Example[​](#example "Direct link to Example")

Let's say you have a feature flag called `new_checkout_flow` with the following configuration:

flags.goff.yaml

```
new_checkout_flow:

  variations:

    enabled: true

    disabled: false

  targeting:

    - query: user_segment eq "early_adopters"

      variation: enabled

  defaultRule:

    percentage:

      enabled: 20

      disabled: 80
```

If a user belongs to the `early_adopters` segment, the flag evaluation will return `true`.<br /><!-- -->If a user does not belong to that segment, there's a **20% chance** the flag will return `true` due to the percentage rollout rule. Otherwise, the flag will return the `false`.

## Basic evaluation with GO Feature Flag[​](#basic-evaluation-with-go-feature-flag "Direct link to Basic evaluation with GO Feature Flag")

After setting up your flag configuration, you can evaluate the flag in your application code. For this example *(in JS)*, we'll use the OpenFeature SDKs, with a GO Feature Flag provider:

```
const evaluationContext = {

  targetingKey: "<my-key>",

   "user_segment": "early_adopters",

};

const defaultValue = false;



// get a boolean value

const boolValue = await client.getBooleanValue(

  'new_checkout_flow',

  defaultValue,

  evaluationContext);
```

As you can see, the evaluation context API takes 3 parameters *(this is true in most languages)*:

* The name of the flag to evaluate.
* A default value to return if the flag is not found or the evaluation fails. The evaluation API will always return a value, even if there is an error.
* An evaluation context giving information about the current user or request *([more information about evaluation context](https://gofeatureflag.org/docs/concepts/evaluation-context.md))*.

You can also see that we are explicitly asking for a boolean value because `new_checkout_flow` is a boolean flag.

## Benefits of Dynamic Evaluation[​](#benefits-of-dynamic-evaluation "Direct link to Benefits of Dynamic Evaluation")

Dynamic flag evaluation provides several advantages:

* **Instant control:** Change the state of features without deploying new code.
* **A/B testing:** Run experiments and gather data on different feature variations.
* **Gradual rollouts:** Release features to a small group of users before wider deployment.
* **Kill switches:** Quickly disable problematic features in production.

## How Evaluation is performed inside your provider?[​](#how-evaluation-is-performed-inside-your-provider "Direct link to How Evaluation is performed inside your provider?")

Feature flag evaluation in GO Feature Flag depends on which type of SDK and provider you are using in your application.<br /><!-- -->There are two main categories **Client Providers** and **Server Providers** (see [SDK paradigms](https://gofeatureflag.org/docs/concepts/sdk-paradigms.md))

The evaluation process is handled differently for each, below we will detail all the flows.

### Client Provider[​](#client-provider "Direct link to Client Provider")

When using a client SDK (e.g., JavaScript/React, Android, Swift), since you have a single evaluation context, GO Feature Flag will evaluate all flags once and you will use the local cache of evaluation.

Details

### Server Provider: In Process Evaluation[​](#server-provider-in-process-evaluation "Direct link to Server Provider: In Process Evaluation")

* Your server-side provider fetches the flag configuration from the Relay Proxy once and caches it locally.
* Each evaluation is performed directly inside the provider — no per-evaluation network call to the Relay Proxy is needed.

Details

### Server Provider: Remote Evaluation[​](#server-provider-remote-evaluation "Direct link to Server Provider: Remote Evaluation")

* Your server-side provider is configured to call the Relay Proxy for flag evaluations.
* This can simplify config management and centralize evaluations, but adds network round-trips on every evaluation.

Details
