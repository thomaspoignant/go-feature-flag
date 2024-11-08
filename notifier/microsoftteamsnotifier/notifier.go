package microsoftteamsnotifier

import (
	"fmt"
	"os"
	"strings"

	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/luci/go-render/render"
	"github.com/r3labs/diff/v3"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type Notifier struct {
	MicrosoftTeamsWebhookURL string

	teamsClient *goteamsnotify.TeamsClient
}

func (c *Notifier) Notify(diff notifier.DiffCache) error {
	if c.MicrosoftTeamsWebhookURL == "" {
		return fmt.Errorf("error: (Microsoft Teams Notifier) invalid notifier configuration, no " +
			"MicrosoftTeamsWebhookURL provided for the microsoft teams notifier")
	}
	msgText := convertToMicrosoftTeamsMessage(diff)
	msg, err := adaptivecard.NewSimpleMessage(msgText, "GO Feature Flag", true)
	if err != nil {
		return err
	}
	if c.teamsClient == nil {
		c.teamsClient = goteamsnotify.NewTeamsClient()
	}
	return c.teamsClient.Send(c.MicrosoftTeamsWebhookURL, msg)
}

func convertToMicrosoftTeamsMessage(diffCache notifier.DiffCache) string {
	hostname, _ := os.Hostname()
	msgText := fmt.Sprintf("Changes detected in your feature flag file on: **%s**", hostname)
	for key := range diffCache.Deleted {
		msgText += fmt.Sprintf("\n * ❌ Flag **%s** deleted", key)
	}

	for key := range diffCache.Added {
		msgText += fmt.Sprintf("\n * 🆕 Flag **%s** created", key)
	}

	for key, value := range diffCache.Updated {
		msgText += fmt.Sprintf("\n * ✏️ Flag **%s** updated", key)
		changelog, _ := diff.Diff(value.Before, value.After, diff.AllowTypeMismatch(true))
		for _, change := range changelog {
			if change.Type == "update" {
				msgText += fmt.Sprintf("\n   * %s: %s => %s", strings.Join(change.Path, "."),
					render.Render(change.From), render.Render(change.To))
			}
		}
	}
	return msgText
}
