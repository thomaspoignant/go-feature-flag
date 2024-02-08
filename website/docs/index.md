---
title: Home
description:  go-feature-flag is a simple and complete feature flag solution, without any complex backend system to install. You need only a file as your backend.
sidebar_position: 1
---

<p align="center">
  <img width="250" height="238" src="/img/logo/logo.png" alt="go-feature-flag logo" />
</p>

## What is GO Feature Flag?
GO Feature Flag is a completely open-source, simple and lightweight feature flag solution.

The solution has been built for application of feature flags in your code without the need of any vendor.

**GO Feature Flag** was initially developed for the GO language, but with the new standardisation of feature flags by [Openfeature](https://openfeature.dev/) project, it now supports multiple languages _(`JAVA`, `typescript`, `javascript`, ...)_ with a simple server to host.

:::info
If you are not familiar with feature flags, also called feature toggles, you can read this [article from Martin Fowler](https://www.martinfowler.com/articles/feature-toggles.html)
where he explains why this is a great pattern.

I've also written an [article](https://medium.com/better-programming/feature-flags-and-how-to-iterate-quickly-7e3371b9986) which explains why feature flags can accelerate your iteration cycle.
:::

## What can I do with GO Feature Flag?

- Storing your configuration flags file on various locations (`HTTP`, `S3`, `Kubernetes`, [_see full list_](configure_flag/store_your_flags.mdx)).
- Configuring your flags in various [format](configure_flag/flag_format.mdx) (`JSON`, `TOML` and `YAML`).
- Adding complex [rules](configure_flag/flag_format.mdx#rule-format) to target your users.
- Use complex rollout strategy for your flags :
    - [Run A/B testing experimentation](configure_flag/rollout/experimentation.mdx).
    - [Progressively rollout a feature](configure_flag/rollout/progressive.mdx).
    - [Schedule your flag updates](configure_flag/rollout/scheduled.mdx).
- Exporting your flags usage data ([`s3`](go_module/data_collection/s3.md), [`log`](go_module/data_collection/log.md), [`file`](go_module/data_collection/file.md), [_see full list_](configure_flag/export_flags_usage.mdx)).
- Getting notified when a flag has been changed ([`webhook`](go_module/notifier/webhook.md) and [`slack`](go_module/notifier/slack.md)).
- **Cross-Language Support:** Available for use across several programming languages.
