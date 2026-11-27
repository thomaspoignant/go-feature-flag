# 📣 Notify flag changes

If you want to be informed when a flag has changed, you can configure a [**notifier**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#NotifierConfig).

A notifier will send one notification to the targeted system to inform them that a new flag configuration has been loaded.

info

`go-feature-flag` can handle more than one notifier at a time.

## Available notifiers[​](#available-notifiers "Direct link to Available notifiers")

* [Slack](https://gofeatureflag.org/../integrations/notify-flags-changes/slack)
* [Discord](https://gofeatureflag.org/../integrations/notify-flags-changes/discord)
* [Microsoft Teams](https://gofeatureflag.org/../integrations/notify-flags-changes/microsoft-teams)
* [Webhook](https://gofeatureflag.org/../integrations/notify-flags-changes/webhook)

info

In each notifier documentation, you will find the configuration needed to set up the notifier.
