# 🎨 A first look at GO Feature Flag Studio

November 4, 2026 ·

<!-- -->

3 min read

[![Thomas Poignant](https://github.com/thomaspoignant.png)](https://github.com/thomaspoignant)

[Thomas Poignant](https://github.com/thomaspoignant)

Creator of GO Feature Flag

Since the beginning, GO Feature Flag has been about keeping things simple: your flags live in a file, the relay proxy reads it, and you have no backend to run. This works great, but one question keeps coming back: *"Is there a UI to manage my flags?"*

Today I want to show you what we are building to answer it: **GO Feature Flag Studio**.

[GO Feature Flag Studio demo](https://www.youtube.com/embed/t7V5nVGWgCE)

Work in progress

Studio is under active development. Features, APIs and the data model can still change, there is no release yet, and it is **not ready for production**. We are sharing it early because we want your feedback while it is still easy to change things.

## What is Studio?[​](#what-is-studio "Direct link to What is Studio?")

[GO Feature Flag Studio](https://github.com/go-feature-flag/studio) is a web interface for GO Feature Flag. It is the place where humans manage flags:

* **Create and edit flags** with a guided wizard: variations, targeting rules, percentage and progressive rollouts, without writing YAML by hand (you still see the exact payload before saving).
* **Organize by team and flagset**, with roles to control who can change what.
* **Keep a history of every change**, with a diff for each version and the option to revert.
* **Test a flag** against an evaluation context before shipping it.
* **Copy a code sample** for your language to start evaluating the flag with OpenFeature.

## How does it plug into GO Feature Flag?[​](#how-does-it-plug-into-go-feature-flag "Direct link to How does it plug into GO Feature Flag?")

Studio does not replace anything in GO Feature Flag, it sits on top of it.

Studio stores your flags in PostgreSQL, and the relay proxy reads them with the existing [PostgreSQL retriever](https://gofeatureflag.org/docs/integrations/store-flags-configuration/postgresql.md), exactly like it would read a file from S3 or GitHub. Studio even generates the `goff-proxy.yaml` for you, with one flagset per Studio flagset.

Your applications don't change either: they keep using the same OpenFeature SDKs and the same relay proxy.

## It is optional, and it will stay optional[​](#it-is-optional-and-it-will-stay-optional "Direct link to It is optional, and it will stay optional")

This is important to me: **GO Feature Flag will stay file-first**. If you are happy managing your flags as code in Git, S3, Kubernetes or anything else, nothing changes for you and you don't need Studio.

Studio is for the teams who want a UI on top of GO Feature Flag, for example to let product managers or support engineers change a rollout without opening a pull request. It is up to you to use it or not.

## Give us your feedback[​](#give-us-your-feedback "Direct link to Give us your feedback")

The code is open source and available at [github.com/go-feature-flag/studio](https://github.com/go-feature-flag/studio). If you try it, or if you have ideas about what a UI for GO Feature Flag should (or should not) do, please [open an issue](https://github.com/go-feature-flag/studio/issues). This is the best moment to shape it.

**Tags:**

* [studio](https://gofeatureflag.org/blog/tags/studio)
* [feature flags](https://gofeatureflag.org/blog/tags/feature-flags)
* [ui](https://gofeatureflag.org/blog/tags/ui)
* [relay proxy](https://gofeatureflag.org/blog/tags/relay-proxy)
* [postgresql](https://gofeatureflag.org/blog/tags/postgresql)

[Edit this page](https://github.com/thomaspoignant/go-feature-flag/tree/main/website/blog/2026-11-04-introducing-goff-studio/index.md)
