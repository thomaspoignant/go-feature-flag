# Notifiers
If you want to be informed when a flag has changed, you can configure a [**notifier**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#NotifierConfig).

A notifier will send one notification to the targeted system to inform them that a new flag configuration has been loaded.

!!! Info
    `go-feature-flag` can handle more than one notifier at a time.

Available notifiers are:

- [Slack](slack.md) - Get a slack message with the changes.
- [Webhook](webhook.md) - Call an API with the changes.
