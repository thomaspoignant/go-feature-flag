---
sidebar_position: 1
---

# Slack Notifier
The **Slack** notifier allows you to get notification on your favorite slack channel when an instance of `go-feature-flag` is detecting changes in the configuration file.

<center><img src="/docs/notifier/slack_notification.png" /></center>

## Configure Slack Notification
1. First, you need to create an incoming webhook on your slack instance.  
   *You can follow this [documentation to see how to do it](https://api.slack.com/messaging/webhooks#getting_started)*
2. Copy your webhook URL.  
   It should look like: `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX`.
3. In your init method add a slack notifier

```go {5}showLineNumbers
ffclient.Config{ 
    // ...
    Notifiers: []notifier.Notifier{
        &slacknotifier.Notifier{
            SlackWebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
        },
        // ...
    },
}
```

### Configuration fields

| Field  | Description  |
|---|---|
|`SlackWebhookURL`   | The complete URL of your incoming webhook configured in Slack.  |
