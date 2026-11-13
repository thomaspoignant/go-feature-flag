# Slack

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Send notifications to a Slack channel.

<!-- -->

## About <!-- -->Slack<!-- --> Notifier[​](#about-notifiername-notifier "Direct link to about-notifiername-notifier")

The **Slack** notifier allows you to get notification on your favorite slack channel when an instance of `go-feature-flag` is detecting changes in the configuration file.

![Slack Notification](/docs/notifier/slack_change_notification.png)

Each notification lists the flags that were created, updated or deleted. For an updated flag, every change is described as a short sentence, for example:

```
✏️ Flag "dark-mode" updated

• Version set to `2`

• Targeting rule `beta-users` now matches `beta eq true or company eq "acme"` (was `beta eq true`)

• Targeting rule `internal` added: serves `enabled` when `email ew "@example.com"`

• Default rule now serves `disabled` 50%, `enabled` 50% (was `disabled` 90%, `enabled` 10%)
```

A flag shows at most 50 changes, and a last line counts the changes left out.

## Configure Slack Notification[​](#configure-slack-notification "Direct link to Configure Slack Notification")

1. First, you need to create an incoming webhook on your slack instance. *You can follow this [documentation to see how to do it](https://api.slack.com/messaging/webhooks#getting_started)*
2. Copy your webhook URL. It should look like: `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX`.
3. Now you can configure your notifier

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Slack<!-- --> notifier, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

notifier:

  - kind: slack

    webhookUrl: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

# ...
```

| Field name   | Mandatory | Type   | Default  | Description                                                                                             |
| ------------ | --------- | ------ | -------- | ------------------------------------------------------------------------------------------------------- |
| `kind`       | **        | string | **none** | **Value should be `slack`**.<br />*This field is mandatory and describe which retriever you are using.* |
| `webhookUrl` | **        | string | **none** | The complete URL of your incoming webhook configured in Slack.                                          |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Slack<!-- --> notifier, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
err := ffclient.Init(ffclient.Config{

    // ...

    Notifiers: []notifier.Notifier{

        &slacknotifier.Notifier{

            SlackWebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",

        },

        // ...

    },

})

defer ffclient.Close()
```

| **Field**         | **Mandatory** | **Description**                         |
| ----------------- | ------------- | --------------------------------------- |
| `SlackWebhookURL` | **            | The complete URL of your slack webhook. |
