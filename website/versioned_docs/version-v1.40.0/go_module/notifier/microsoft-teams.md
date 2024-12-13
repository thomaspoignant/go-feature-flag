---
sidebar_position: 11
---

# Microsoft Teams Notifier

The microsoft teams notifier allows to get notified on your favorite microsoft teams channel when an instance of `go-feature-flag` is
detecting changes in the configuration of your flags.


## Configure Microsoft Teams Notifications

1. First create a Webhook in the channel you want to send notifications to.  
   *Need a hand?*[Click here to see how it's done](https://support.microsoft.com/en-us/office/create-incoming-webhooks-with-workflows-for-microsoft-teams-8ae491c7-0394-4861-ba59-055e33f75498)
2. Copy the webhook URL
3. Now you can configure your notifier

```go
err := ffclient.Init(ffclient.Config
{
    // ...
    Notifiers: []notifier.Notifier{
        &microsoftteamsnotifier.Notifier{
			MicrosoftTeamsWebhookURL: "https://xxx.xxx/..."
		},
    },
})
```

## **Configuration fields**

| **Field**                  | **Description**                                   |
|----------------------------|---------------------------------------------------|
| `MicrosoftTeamsWebhookURL` | The complete URL of your microsoft teams webhook. |
