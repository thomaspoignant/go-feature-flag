---
sidebar_position: 20
description: An evaluation context in a feature flagging system is crucial for determining the output of a feature flag evaluation. It's a collection of pertinent data about the conditions under which the evaluation is being made.
---
# 🔎 Evaluation Context

## Overview
All GO Feature Flag SDKs use an evaluation context to determine which variation of a feature flag to serve.

An evaluation context in a feature flagging system is crucial for determining the output of a feature flag evaluation.

It's a collection of pertinent data about the conditions under which the evaluation is being made.
This data can be supplied through a mix of static information _(server name, IP, etc ...)_ and dynamic inputs
_(information about the user performing the action, etc ...)_, along with state information that is implicitly carried
through the execution of the program.

## About contexts
GO Feature Flag evaluation contexts are data objects representing users, devices, organizations, and other entities that are used to determine which variation of a feature flag to serve. 

The context is used to evaluate the [targeting queries](../configure_flag/target-with-flags.mdx) and determine which variation of a feature flag to serve to a user.

:::info
**Only the evaluation context attributes you provide are available for targeting queries.**

If you want to use a specific attribute in your targeting queries, you must include it in the evaluation context.
If not present the query will not apply.
:::

## Example context: Mike Wazowski at Monsters, Inc.

As an example, let's assume Mike Wazowski is one of your end users. He is a scare assistant who works on the scare floor at Monsters, Inc with James P. Sullivan.   
Mike has two mobile devices, an Android phone and an iPad tablet. Mike uses your application on both devices as part of his work.

Given this information, you may know the following things about Mike Wazowski:

- his name, email and job title _("Scare assistant")_
- his employee ID _(used for the targetingKey)_
- his work station position _("scare floor")_
- his organization's name _("Monsters, Inc."),_
- his device's type _("iPad")_
- his coworker's name _("James P. Sullivan")_

Here is an example of what the data structure for Mike Wazowski evaluation context object might look like:

```json title="evaluation context"
{
  "targetingKey": "34c7f8ab-6d14-4aa6-a77f-effc6245da6f",
  "firstname": "Mike",
  "lastname": "Wazowski",
  "email": "mike.wazowski@monster.inc",
  "organization": "Monsters, Inc.",
  "jobFunction": "Scare assistant",
  "location": "scare floor",
  "coworker": "James P. Sullivan",
  "device": "iPad"
}
```
## Targeting Key
:::info
**The targeting key requirement depends on the flag and provider.**
:::

Percentage and progressive rollouts require a non-empty key for bucketing.
This includes rules introduced by scheduled rollout steps.
By default, the `targetingKey` supplies this value; a configured [custom bucketing key](../configure_flag/custom-bucketing) can supply it instead.

Static flags and attribute-based targeting without bucketing can be evaluated without a targeting key.

A **targeting key** is a unique identifier that represents the context of the evaluation _(email, session id, a fingerprint or anything that is consistent)_,
ensuring that they are consistently exposed to the same variation of a feature, even across multiple visits or sessions.

The targeting key is used to ensure that a user consistently receives the same variation of a feature flag over time.  
For instance, **GO Feature Flag** ensures that in cases where a feature is being rolled out to a percentage of users, based on the targeting key, they will see the same variation each time they encounter the feature flag.

## Reserved properties in the evaluation context
:::danger
The `gofeatureflag` key is reserved. Only set the properties listed below, any other usage may break internal features of GO Feature Flag.

**Do not use `gofeatureflag` or any of its properties in your targeting queries.**  
They are internal to GO Feature Flag and may be absent depending on the provider and the evaluation mode _(remote or in-process)_, so the same rule could return different results.  
If you need to target on one of these values, add it to the evaluation context under your own key.
:::

You can set these properties to change how GO Feature Flag behaves:

| Field                            | Description                                                                                                                                                                                                                  |
|----------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `gofeatureflag.currentDateTime`  | If this property is set, we will use this date as base for all the rollout strategies which implies dates _(experimentation, progressive and scheduled)_.<br/>**Format:** Date following the RFC3339 format.                  |
| `gofeatureflag.flagList`         | If this property is set, in the bulk evaluation mode (for the client SDK) we will only evaluate the flags in this list.<br/>If empty or not set the default behavior is to evaluate all the flags.<br/>**Format:** []string |
| `gofeatureflag.exporterMetadata` | Metadata added to the feature events sent to the exporter when the flag is evaluated by the relay proxy or the GO module.<br/>OpenFeature providers may set it automatically, use the exporter metadata option of your provider instead.<br/>**Format:** map[string]string\|number\|bool |
