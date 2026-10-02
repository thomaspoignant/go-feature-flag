package microsoftteamsnotifier

import (
	"fmt"
	"os"
	"sort"

	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

var _ notifier.Notifier = &Notifier{}

// Notifier is the component in charge of sending flag changes to Microsoft Teams.
type Notifier struct {
	MicrosoftTeamsWebhookURL string

	teamsClient *goteamsnotify.TeamsClient
}

// Notify is the notifying all the changes to the notifier.
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

	// Deleted
	deletedKeys := make([]string, 0, len(diffCache.Deleted))
	for k := range diffCache.Deleted {
		deletedKeys = append(deletedKeys, k)
	}
	sort.Strings(deletedKeys)
	for _, key := range deletedKeys {
		msgText += fmt.Sprintf("\n * ❌ Flag **%s** deleted", key)
	}

	// Added
	addedKeys := make([]string, 0, len(diffCache.Added))
	for k := range diffCache.Added {
		addedKeys = append(addedKeys, k)
	}
	sort.Strings(addedKeys)
	for _, key := range addedKeys {
		msgText += fmt.Sprintf("\n * 🆕 Flag **%s** created", key)
	}

	// Updated
	updatedKeys := make([]string, 0, len(diffCache.Updated))
	for k := range diffCache.Updated {
		updatedKeys = append(updatedKeys, k)
	}
	sort.Strings(updatedKeys)
	for _, key := range updatedKeys {
		value := diffCache.Updated[key]
		msgText += fmt.Sprintf("\n * ✏️ Flag **%s** updated", key)
		for _, change := range notifier.FlagChanges(value.Before, value.After) {
			msgText += fmt.Sprintf("\n   * %s: %s => %s", change.Path, change.From, change.To)
		}
	}
	return msgText
}
