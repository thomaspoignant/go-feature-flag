---
sidebar_position: 1
---

# Notify flag changes
If you want to be informed when a flag has changed, you can configure a [**notifier**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#NotifierConfig).

A notifier will send one notification to the targeted system to inform them that a new flag configuration has been loaded.

:::info
`go-feature-flag` can handle more than one notifier at a time.
:::

Available notifiers are:

- [Slack](slack.md) - Get a slack message with the changes.
- [Webhook](webhook.md) - Call an API with the changes.
- [Discord](discord.md) - Get a discord message with the changes.
- [Microsoft Teams](microsoft-teams.md) - Get a Microsoft Teams message with the changes.
