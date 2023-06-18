---
slug: automate-your-product-release-cycles-using-go-feature-flag
title: "Automate Your Product Release Cycles Using Go Feature Flag"
authors: [thomaspoignant]
tags: [GO Feature Flag, v0.x.x]
---
![](./head.jpg)

When you build a new feature, orchestrating the actual launch schedule between the Product, Engineering, and Marketing teams can be challenging.

While it seems easy to launch something new, a poorly executed rollout can end up being your worst nightmare.

In this article, I will present to you how to use the Go module go-feature-flag to roll out your new features smoothly and help you be confident during the rollout phase. If you are not familiar with the concept of feature flags or feature toggles, I encourage you to read this [article by Martin Fowler](https://www.martinfowler.com/articles/feature-toggles.html).

go-feature-flag is a Go module to easily manage your flags. You can refer to this article I wrote a few months ago to understand how it works.
<!-- truncate -->

---

## How To Use go-feature-flag

The library is super simple to use:

1. Install the module:
```go
go get github.com/thomaspoignant/go-feature-flag
```

2. Init the client with the location of your configuration file for your flags:

```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 60,
    Retriever: &ffclient.HTTPRetriever{
        URL:    "http://example.com/flag-config.yaml",
    },
})
defer ffclient.Close()
```

3. Put your new features conditionally based on the flag value:

```go
user := ffcontext.NewEvaluationContext("user-unique-key")
hasFlag, _ := ffclient.BoolVariation("test-flag", user, false)
if hasFlag {
    // flag "test-flag" is true for the user
} else {
    // flag "test-flag" is false for the user
}
```

You can have way more configuration, but I’ll let you check the [documentation](https://thomaspoignant.github.io/go-feature-flag/) for that.

---

## Progressive Rollout

When you release a new feature that can have a strong impact on your system, you probably don’t want to go all-in with this new feature for all your users.

For example, you are deploying something that can be CPU-consuming and you want to have time to check if your servers handle it correctly. This is typically a great use case for a progressive rollout.

It progressively increases how many users are impacted by your feature flag to avoid a big-bang rollout for all your users at once. During that time, it allows you to monitor your system and be confident that your infrastructure can handle this new load.

```yaml
progressive-flag:
  true: true
  false: false
  default: false
  rollout:
    progressive:
      percentage:
        initial: 0
        end: 100
      releaseRamp:
        start: 2021-03-20T00:00:00.10-05:00
        end: 2021-03-21T00:00:00.10-05:00
```

To do that in go-feature-flag, you will configure your flag like in the example above. You set up a progressive rollout with an initial percentage value and a release ramp. Over time, more and more users can be affected by the flag and will have the new feature.

Depending on how critical this feature is, you can have a long or a short release ramp. If something goes wrong, you can edit your flag to stop the rollout at any time.

---

## Scheduling Workflows

Scheduling introduces the ability for users to change their flags for future points in time. While this sounds deceptively straightforward, it unlocks the potential for users to create complex release strategies by scheduling the incremental steps in advance.

For example, you may want to turn a feature on for internal testing tomorrow and then enable it for your “beta” user segment four days later.

```yaml
scheduled-flag:
  true: true
  false: false
  default: false
  percentage: 0
  rollout:
    scheduled:
      steps:
        - date: 2020-04-10T00:00:00.00+02:00
          rule: internal eq true
          percentage: 100        - date: 2020-04-14T00:00:00.00+02:00
          rule: internal eq true and beta eq true        - date: 2020-04-18T00:00:00.00+02:00
          rule: ""
```

In this example, you can see that we are updating the flag multiple times to perform actions in the future. Let’s detail what will happen in this configuration:

1. Before `2020–04–10`, the flag is not served.
2. After the first update of the flag (`2020–04–10`), we have 100% of the internal users who received the flag as true.
3. Four days later, we add the users who have a `bet`a flag as `true`.
4. Finally, four days later, we open the feature to all users.

As you can see, this is really powerful because your release management is now ready without doing any manual deployment/action, and this scheduling can be done by a non-technical user (aka your product manager).

---

## Experimentation Rollout

Sometimes you also want to experiment, collect the data, and decide later if you want to roll out the feature to all your users.

To do that correctly, you can configure your flag with a start date and an end date for a subset of your users.

```yaml
experimentation-flag:
  percentage: 50
  true: true
  false: false
  default: false
  rule: userId sw "9"
  rollout:
    experimentation:
      start: 2021-03-20T00:00:00.10-05:00
      end: 2021-03-21T00:00:00.10-05:00
```

In this example, 50% of your users with a userId that starts with 9 will be impacted by the flag between the start and the end dates of the experimentation.

With the module, you can also collect the data of your variation (see the [documentation](https://thomaspoignant.github.io/go-feature-flag/data_collection/) for more info) to join them with the data of what you are testing.

So you can see the results of your experimentation and decide whether you want to roll out this flag for real or not.

---

## Conclusion

Using feature flags really is a great thing, but it becomes even better if you use some advanced rollout strategies.

If you start using them, you will love it because you decouple the deployment and the release and you can have fine-grained control over what your users can do.
