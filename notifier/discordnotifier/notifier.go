package discordnotifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

var _ notifier.Notifier = &Notifier{}

const (
	colorDeleted     = 15158332
	colorUpdated     = 16753920
	colorAdded       = 32768
	longDiscordField = 35
)

// Notifier is the component in charge of sending flag changes to Discord.
type Notifier struct {
	DiscordWebhookURL string

	httpClient internal.HTTPClient
	init       sync.Once
}

// Notify is the notifying all the changes to the notifier.
func (c *Notifier) Notify(diff notifier.DiffCache) error {
	if c.DiscordWebhookURL == "" {
		return fmt.Errorf(
			"error: (Discord Notifier) invalid notifier configuration, no DiscordWebhookURL provided",
		)
	}

	// init the notifier
	c.init.Do(func() {
		if c.httpClient == nil {
			c.httpClient = internal.DefaultHTTPClient()
		}
	})

	discordURL, err := url.Parse(c.DiscordWebhookURL)
	if err != nil {
		return fmt.Errorf(
			"error: (Discord Notifier) invalid DiscordWebhookURL: %v",
			c.DiscordWebhookURL,
		)
	}

	reqBody := convertToDiscordMessage(diff)
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error: (Discord Notifier) failed to create message payload; %v", err)
	}
	request := http.Request{
		Method: http.MethodPost,
		URL:    discordURL,
		Body:   io.NopCloser(bytes.NewReader(payload)),
		Header: map[string][]string{"Content-Type": {"application/json"}},
	}
	response, err := c.httpClient.Do(&request)
	if err != nil {
		return fmt.Errorf("error: (Discord Notifier) error calling webhook: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode > 399 {
		return fmt.Errorf(
			"error: (Discord Notifier) webhook call failed with statusCode = %d",
			response.StatusCode,
		)
	}
	return nil
}

func convertToDiscordMessage(diffCache notifier.DiffCache) discordMessage {
	hostname, _ := os.Hostname()
	embeds := convertDeletedFlagsToDiscordEmbed(diffCache)
	embeds = append(embeds, convertUpdatedFlagsToDiscordEmbed(diffCache)...)
	embeds = append(embeds, convertAddedFlagsToDiscordEmbed(diffCache)...)

	sort.Slice(embeds, func(i, j int) bool {
		return embeds[i].Title < embeds[j].Title
	})

	if len(embeds) > 10 {
		embeds = embeds[:9]
		embeds = append(embeds, embed{
			Title: "Wow, lots of updates!\nToo many to fit here. 😔\n\nCheck the logs for the full list.",
			Color: colorDeleted,
		})
	}

	return discordMessage{
		Content: fmt.Sprintf("Changes detected in your feature flag file on: **%s**", hostname),
		Embeds:  embeds,
	}
}

func convertDeletedFlagsToDiscordEmbed(diffCache notifier.DiffCache) []embed {
	embeds := make([]embed, 0, len(diffCache.Deleted))
	for key := range diffCache.Deleted {
		embeds = append(embeds, embed{
			Title: fmt.Sprintf("❌ Flag \"%s\" deleted", key),
			Color: colorDeleted,
		})
	}
	return embeds
}

func convertUpdatedFlagsToDiscordEmbed(diffCache notifier.DiffCache) []embed {
	embeds := make([]embed, 0, len(diffCache.Updated))
	for key, value := range diffCache.Updated {
		changes := notifier.FlagChanges(value.Before, value.After)
		fields := make([]embedField, 0, len(changes))
		for _, change := range changes {
			fieldValue := fmt.Sprintf("%s => %s", change.From, change.To)
			short := len(fieldValue) < longDiscordField
			fields = append(fields, embedField{
				Name:   change.Path,
				Value:  fieldValue,
				Inline: short,
			})
		}
		sort.Sort(byTitle(fields))
		embeds = append(embeds, embed{
			Title:  fmt.Sprintf("✏️ Flag \"%s\" updated", key),
			Color:  colorUpdated,
			Fields: fields,
		})
	}
	return embeds
}

func convertAddedFlagsToDiscordEmbed(diff notifier.DiffCache) []embed {
	embeds := make([]embed, 0, len(diff.Added))
	for key := range diff.Added {
		embeds = append(embeds, embed{
			Title: fmt.Sprintf("🆕 Flag \"%s\" created", key),
			Color: colorAdded,
		})
	}
	return embeds
}

type discordMessage struct {
	Content string  `json:"content"`
	Embeds  []embed `json:"embeds"`
}

type embed struct {
	Title  string       `json:"title"`
	Color  int          `json:"color"`
	Fields []embedField `json:"fields,omitempty"`
}

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type byTitle []embedField

func (a byTitle) Len() int           { return len(a) }
func (a byTitle) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a byTitle) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
