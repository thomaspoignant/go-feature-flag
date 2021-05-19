<p align="center">
  <img width="250" height="238" src="assets/logo.png" alt="go-feature-flag logo" />
</p>

<p align="center">
    <a href="https://app.circleci.com/pipelines/github/thomaspoignant/go-feature-flag"><img src="https://img.shields.io/circleci/build/github/thomaspoignant/go-feature-flag" alt="Build Status" /></a>
    <a href="https://coveralls.io/github/thomaspoignant/go-feature-flag"><img src="https://coveralls.io/repos/github/thomaspoignant/go-feature-flag/badge.svg" alt="Coverage Status" /></a>
    <a href="https://sonarcloud.io/dashboard?id=thomaspoignant_go-feature-flag"><img src="https://sonarcloud.io/api/project_badges/measure?project=thomaspoignant_go-feature-flag&metric=alert_status" alt="Sonarcloud Status" /></a>
    <a href="https://github.com/thomaspoignant/go-feature-flag/actions?query=workflow%3ACodeQL"><img src="https://github.com/thomaspoignant/go-feature-flag/workflows/CodeQL/badge.svg" alt="Build Status" /></a>
    <a href="https://app.fossa.com/projects/git%2Bgithub.com%2Fthomaspoignant%2Fgo-feature-flag"><img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2Fthomaspoignant%2Fgo-feature-flag.svg?type=shield" alt="FOSSA Status"/></a>
    <br/>
    <a href="https://github.com/thomaspoignant/go-feature-flag/releases"><img src="https://img.shields.io/github/v/release/thomaspoignant/go-feature-flag" alt="Release version" /></a>
    <a href="https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag"><img src="https://godoc.org/github.com/thomaspoignant/go-feature-flag?status.svg" alt="GoDoc" /></a>
    <img src="https://img.shields.io/github/go-mod/go-version/thomaspoignant/go-feature-flag?logo=go%20version" alt="Go version"/>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/thomaspoignant/go-feature-flag" alt="License"/></a>
    <a href="https://github.com/avelino/awesome-go/#server-applications"><img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go"></a>
</p>

# go-feature-flag

**Feature flags with no complex system to maintain!**

## What is go-feature-flag?

A simple and complete feature flag solution, without any complex backend system to install, you need only a file as your backend.

No server is needed, just add a file in a central system and all your services will react to the changes of this file. 

If you are not familiar with feature flags also called feature Toggles you can read this [article of Martin Fowler](https://www.martinfowler.com/articles/feature-toggles.html)
that explains why this is a great pattern.

I've also written an [article](https://medium.com/better-programming/feature-flags-and-how-to-iterate-quickly-7e3371b9986) that explains why feature flags can help you to iterate quickly.

**go-feature-flags supports:**

- Storing your configuration flags file on various locations ([`HTTP`](./flag_file/http.md), [`S3`](.flag_file/s3.md), [`GitHub`](./flag_file/github.md), [`file`](./flag_file/file.md)).
- Configuring your flags in various [format](flag_format.md) (`JSON`, `TOML` and `YAML`).
- Adding complex [rules](flag_format.md#rule-format) to target your users.
- Use complex rollout strategy for your flags
    - [Run A/B testing experimentation](rollout/experimentation.md).
    - [Progressively rollout a feature](rollout/progressive.md).
    - [Schedule your flag updates](rollout/scheduled.md).
- Exporting your flags usage data ([`s3`](data_collection/s3.md), [`log`](data_collection/log.md) and [`file`](data_collection/file.md)).
- Getting notified when a flag has changed ([`webhook`](notifier/webhook.md) and [`slack`](notifier/slack.md)).


