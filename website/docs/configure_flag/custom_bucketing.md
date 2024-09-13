---
sidebar_position: 22
description: How to bucket users based on a custom identifier
---

# Custom bucketing

When evaluating flags, the `targetingKey` is usually given a user ID. This key ensures that a user will always be in the same group for each flag.

Sometimes, it is desireable to _bucket_ users based on a different value. The `bucketingKey` field in the flag configuration allows you to define a different identifier to be used instead. For example:

```yaml
first-flag:
  bucketingKey: "teamId"
  variations:
    A: false
    B: true
  defaultRule: # When no targeting match we use the defaultRule
    percentage:
        A: 50
        B: 50
```

With this flag configuration, the `teamId` value will be used for hashing instead of `targetingKey`. The value must be provided to the evaluation context:


```go
user = ffcontext.NewEvaluationContextBuilder("user126")
  .AddCustom("teamId", "f74b72")
  .Build()

ffclient.BoolVariation("first-flag", user, false)
```

As a result, users who are members of the same team will receive the same flag variation, consistently. A different `bucketingKey` can be used per experiment, though normally you'll only have a handful of possible values.

This is useful for A/B testing, permissions management and other use cases where targeting a consistent group of users is required.

**Note**: if a value in the corresponding `bucketingKey` is not found in the evaluation context, the flag rules will not be evaluated, and the SDK will return the default value.