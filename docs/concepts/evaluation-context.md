# 🔎 Evaluation Context

## Overview[​](#overview "Direct link to Overview")

All GO Feature Flag SDKs use an evaluation context to determine which variation of a feature flag to serve.

An evaluation context in a feature flagging system is crucial for determining the output of a feature flag evaluation.

It's a collection of pertinent data about the conditions under which the evaluation is being made. This data can be supplied through a mix of static information *(server name, IP, etc ...)* and dynamic inputs *(information about the user performing the action, etc ...)*, along with state information that is implicitly carried through the execution of the program.

## About contexts[​](#about-contexts "Direct link to About contexts")

GO Feature Flag evaluation contexts are data objects representing users, devices, organizations, and other entities that are used to determine which variation of a feature flag to serve.

The context is used to evaluate the [targeting queries](https://gofeatureflag.org/docs/configure_flag/target-with-flags.md) and determine which variation of a feature flag to serve to a user.

info

**Only the evaluation context attributes you provide are available for targeting queries.**

If you want to use a specific attribute in your targeting queries, you must include it in the evaluation context. If not present the query will not apply.

## Example context: Mike Wazowski at Monsters, Inc.[​](#example-context-mike-wazowski-at-monsters-inc "Direct link to Example context: Mike Wazowski at Monsters, Inc.")

As an example, let's assume Mike Wazowski is one of your end users. He is a scare assistant who works on the scare floor at Monsters, Inc with James P. Sullivan.<br /><!-- -->Mike has two mobile devices, an Android phone and an iPad tablet. Mike uses your application on both devices as part of his work.

Given this information, you may know the following things about Mike Wazowski:

* his name, email and job title *("Scare assistant")*
* his employee ID *(used for the targetingKey)*
* his work station position *("scare floor")*
* his organization's name *("Monsters, Inc."),*
* his device's type *("iPad")*
* his coworker's name *("James P. Sullivan")*

Here is an example of what the data structure for Mike Wazowski evaluation context object might look like:

evaluation context

```
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

## Targeting Key[​](#targeting-key "Direct link to Targeting Key")

info

**Targeting Key is a mandatory field in GO Feature Flag.**

A **targeting key** is a unique identifier that represents the context of the evaluation *(email, session id, a fingerprint or anything that is consistent)*, ensuring that they are consistently exposed to the same variation of a feature, even across multiple visits or sessions.

The targeting key is used to ensure that a user consistently receives the same variation of a feature flag over time.<br /><!-- -->For instance, **GO Feature Flag** ensures that in cases where a feature is being rolled out to a percentage of users, based on the targeting key, they will see the same variation each time they encounter the feature flag.

## Reserved properties in the evaluation context[​](#reserved-properties-in-the-evaluation-context "Direct link to Reserved properties in the evaluation context")

danger

If you put a key named `gofeatureflag` in your evaluation context, it may break internal features of GO Feature Flag. This property name is reserved for internal use.

When you create an evaluation context some fields are reserved for GO Feature Flag.<br /><!-- -->Those fields are used by GO Feature Flag directly, you can use them as will in your targeting queries, but you should be aware that they are used internally for GO Feature Flag.

| Field                            | Description                                                                                                                                                                                                                    |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `gofeatureflag.currentDateTime`  | If this property is set, we will use this date as base for all the rollout strategies which implies dates *(experimentation, progressive and scheduled)*.<br />**Format:** Date following the RFC3339 format.                  |
| `gofeatureflag.flagList`         | If this property is set, in the bulk evaluation mode (for the client SDK) we will only evaluate the flags in this list.<br />If empty or not set the default behavior is to evaluate all the flags.<br />**Format:** \[]string |
| `gofeatureflag.exporterMetadata` | If this property is set, we will add all the fields in the feature event send to the provider.<br />**Format:** map\[string]string\|number\|bool                                                                               |
