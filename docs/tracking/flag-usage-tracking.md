# 📈 Flag usage tracking

## What it is[​](#what-it-is "Direct link to What it is")

**Flag usage tracking** records which flag variation each user received whenever a flag is evaluated.<br /><!-- -->For each evaluation, a **feature** event is recorded and sent to the exporters you have configured.

Depending on the way you evaluate your flags, this event is tracked in the relay-proxy or in the OpenFeature provider.

All flag usage events are sent to the [exporters](https://gofeatureflag.org/docs/integrations/export-evaluation-data.md) you have configured.

## Why use it[​](#why-use-it "Direct link to Why use it")

* **Understand how flags are used**: See who saw what variation, power analytics and dashboards, and feed data into A/B or experimentation tools.
* **Correlate exposures with outcomes**: Essential when you need to link flag evaluations to business metrics or downstream events.

## How it works[​](#how-it-works "Direct link to How it works")

### With the relay-proxy[​](#with-the-relay-proxy "Direct link to With the relay-proxy")

If your app talks to the relay (direct API or via an OpenFeature SDK), evaluations can happen in two places:

1. **In the relay-proxy**: Each evaluation performed by the relay-proxy produces an event that the relay-proxy exports. You can think of the relay-proxy as the component that evaluates and records.
2. **In the OpenFeature provider**: The OpenFeature provider can evaluate locally the flag without calling the relay-proxy. In that case the OpenFeature provider later sends batched "I evaluated this flag for this user" events to the relay-proxy, and the relay-proxy exports them so you don't miss usage when in-process evaluation is used.

Regardless of where the evaluation happened, the relay-proxu is the place that forwards flag-usage events to your configured exporters.

### Summary[​](#summary "Direct link to Summary")

Flag usage tracking means: *every time someone gets a flag value, record that fact and send it to your chosen destinations.* Whether the evaluation runs in the relay-proxy or from the OpenFeature provider only affects who produces the event and who sends it to the relay-proxy; the concept is the same.

## What you need to do[​](#what-you-need-to-do "Direct link to What you need to do")

1. **Turn on tracking for the flags you care about**: Set `trackEvents: true` in the flag configuration of your feature flag(or omit it; the default is `true`). See [configure your flags](https://gofeatureflag.org/docs/configure_flag/create-flags.md) for the flag format.
2. **Configure at least one exporter for evaluation data**: Add a data exporter for feature events (the default event type). See [export evaluation data](https://gofeatureflag.org/docs/integrations/export-evaluation-data.md) for available exporters and how to configure them for the relay-proxy or GO module.

No code changes are required beyond configuration for the default "feature" event type.

## Event format[​](#event-format "Direct link to Event format")

Each flag-usage event (also called a **feature event**) has the following structure when sent to your exporters:

| Field            | Description                                                                                                                                           |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| **kind**         | Always `"feature"` for flag usage events.                                                                                                             |
| **contextKind**  | `"user"` for a known user, or `"anonymousUser"` for an anonymous context.                                                                             |
| **userKey**      | The targeting key (user or entity identifier) used in the evaluation.                                                                                 |
| **creationDate** | Unix timestamp (seconds) when the flag was evaluated.                                                                                                 |
| **key**          | The flag name that was evaluated.                                                                                                                     |
| **variation**    | The variation name returned.                                                                                                                          |
| **value**        | The resolved value returned to the application.                                                                                                       |
| **default**      | `true` if the evaluation failed and the default value was returned; otherwise `false`.                                                                |
| **version**      | Version of the flag configuration (if set).                                                                                                           |
| **source**       | `"SERVER"` when the evaluation was done in the relay-proxy; `"PROVIDER_CACHE"` when it was served from the SDK cache and later reported to the relay. |
| **metadata**     | Optional static key-value data added by the provider for context.                                                                                     |

### Example[​](#example "Direct link to Example")

A minimal feature event in JSON (as sent for example by the [webhook exporter](https://gofeatureflag.org/docs/integrations/export-evaluation-data/webhook.md)) looks like:

```
{

  "kind": "feature",

  "contextKind": "user",

  "userKey": "94a25909-20d8-40cc-8500-fee99b569345",

  "creationDate": 1618909178,

  "key": "my-feature-flag",

  "variation": "true-variation",

  "value": true,

  "default": false,

  "version": "1.0.0",

  "source": "SERVER"

}
```

To combine flag usage with custom business events (e.g. conversions or revenue), use the [OpenFeature Tracking API](https://gofeatureflag.org/docs/tracking/tracking-api.md).
