---
sidebar_position: 10
---

# Discord Notifier

The discord notifier allows to get notified on your favorite discord channel when an instance of `go-feature-flag` is
detecting changes in the configuration of your flags.

![Discord Notification](/docs/notifier/discord1.png)

## Configure Discord Notifications

1. Connect to your discord account and go on the channel where you want to send the notifications.
2. Go on the settings menu of your channel.

   ![Discord Config](/docs/notifier/discord2.png)

3. Under your channel’s settings, go to the "Integrations" section and create a new webhook. To create it, please follow
   the [official documentation](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks).
4. Copy the webhook URL

   ![Discord WebHook](/docs/notifier/discord3.png)

5. Now you can configure your notifier

```go
err := ffclient.Init(ffclient.Config
{
    // ...
    Notifiers: []notifier.Notifier{
        &discordnotifier.Notifier{
            DiscordWebhookURL: "https://discord.com/api/webhooks/yyyy/xxxxxxx",
        },
    },
})
```

## **Configuration fields**

| **Field**           | **Description**                           |
|---------------------|-------------------------------------------|
| `DiscordWebhookURL` | The complete URL of your discord webhook. |
