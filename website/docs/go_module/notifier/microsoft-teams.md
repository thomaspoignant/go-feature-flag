---
sidebar_position: 11
---

# Microsoft Teams Notifier

The microsoft teams notifier allows to get notified on your favorite microsoft teams channel when an instance of `go-feature-flag` is
detecting changes in the configuration of your flags.

![Microsoft Teams Notification](/docs/notifier/microsoftteams1.png)

## Configure Microsoft Teams Notifications

https://learn.microsoft.com/en-us/graph/api/chatmessage-post

1. First, get your microsoft team access token.
   *Need a hand?*[Click here to see how it's done](https://learn.microsoft.com/en-us/graph/auth/auth-concepts?view=graph-rest-1.0#access-tokens)
2. Add the access token in your env file with the key `MICROSOFT_TEAMS_ACCESS_TOKEN`
3. Copy your Team ID (e.g XXXXXX) as well as well as the Channel ID (e.g YYYYYY)
4. Now, generate/copy your webhook URL.  
   It should look like: `https://graph.microsoft.com/teams/XXXXXX/channels/YYYYYY/messages`.
5.  In your init method add a microsoft teams notifier

```go
err := ffclient.Init(ffclient.Config
{
    // ...
    Notifiers: []notifier.Notifier{
        &microsoftteamsnotifier.Notifier{
            MicrosoftTeamsWebhookURL: "https://graph.microsoft.com/teams/XXXXXX/channels/YYYYYY/messages",
        },
    },
})
```

## **Configuration fields**

| **Field**           | **Description**                           |
|---------------------|-------------------------------------------|
| `MicrosoftTeamsWebhookURL` | The complete URL of your microsoft teams webhook. |
