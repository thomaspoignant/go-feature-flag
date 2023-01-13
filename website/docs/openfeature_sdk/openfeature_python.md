---
sidebar_position: 53
description: How to use the OpenFeature Python SDK
---

# Python

## Install dependencies

The first things we will do is install the **Open Feature SDK** and the **GO Feature Flag provider**.

```shell
 TODO
```

## Initialize your Open Feature client

To evaluate the flags you need to have an Open Feature configured in you app.  
This code block shows you how you can create a client that you can use in your application.

```ptython
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from open_feature import open_feature_api
from open_feature.evaluation_context.evaluation_context import EvaluationContext

// ...

goff_provider = GoFeatureFlagProvider(
    options=GoFeatureFlagOptions(endpoint="https://gofeatureflag.org/")
)
open_feature_api.set_provider(goff_provider)
client = open_feature_api.get_client(name="test-client")
```

## Evaluate your flag

This code block explain how you can create an `EvaluationContext` and use it to evaluate your flag.


> In this example we are evaluating a `boolean` flag, but other types are available.
> 
> **Refer to the [Open Feature documentation](https://docs.openfeature.dev/docs/reference/concepts/evaluation-api#basic-evaluation) to know more about it.**

```csharp
// Context of your flag evaluation.
// With GO Feature Flag you MUST have a targetingKey that is a unique identifier of the user.
user_context = EvaluationContext(
                 targeting_key="d45e303a-38c2-11ed-a261-0242ac120002",
                 attributes={
                     "email": "john.doe@gofeatureflag.org",
                     "firstname": "john",
                     "lastname": "doe",
                     "anonymous": False,
                     "admin": True,
                 })
                 
admin_flag = client.get_boolean_value(
          flag_key=flag_key,
          default_value=default_value,
          evaluation_context=ctx,
      )
      
if admin_flag:
  # flag "flag-only-for-admin" is true for the user
else:
  # flag "flag-only-for-admin" is false for the user
```
