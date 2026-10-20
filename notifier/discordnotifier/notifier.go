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
	"unicode/utf8"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

var _ notifier.Notifier = &Notifier{}

const (
	colorDeleted = 15158332
	colorUpdated = 16753920
	colorAdded   = 32768
	// Discord limits, see https://discord.com/developers/docs/resources/message#embed-object-embed-limits
	maxDiscordEmbeds      = 10
	maxDiscordTitle       = 256
	maxDiscordDescription = 4096
	maxDiscordTotal       = 6000
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

	if len(embeds) > maxDiscordEmbeds {
		embeds = embeds[:maxDiscordEmbeds-1]
		embeds = append(embeds, embed{
			Title: "Wow, lots of updates!\nToo many to fit here. 😔\n\nCheck the logs for the full list.",
			Color: colorDeleted,
		})
	}
	for i := range embeds {
		embeds[i].Title = truncate(embeds[i].Title, maxDiscordTitle)
	}
	fitTotalLength(embeds)

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
		e := embed{Title: fmt.Sprintf("✏️ Flag \"%s\" updated", key), Color: colorUpdated}
		changes, err := notifier.DescribeChanges(value.Before, value.After)
		if err != nil {
			e.Description = fmt.Sprintf("%s: %s", notifier.ChangesUnavailable, err.Error())
		} else {
			e.changes = changes
			e.Description = notifier.ChangeList(changes, "• ", maxDiscordDescription)
		}
		embeds = append(embeds, e)
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
	Title       string `json:"title"`
	Color       int    `json:"color"`
	Description string `json:"description,omitempty"`

	// changes are the changes listed in Description, kept to shorten the list if the message is too long.
	changes []string
}

// fitTotalLength shortens change lists so all embeds fit in the characters Discord accepts
// in one message. The budget is shared evenly: short lists stay whole and the longest are cut first.
func fitTotalLength(embeds []embed) {
	budget := maxDiscordTotal
	lists := make([]int, 0, len(embeds))
	for i, e := range embeds {
		budget -= utf8.RuneCountInString(e.Title)
		if len(e.changes) == 0 {
			budget -= utf8.RuneCountInString(e.Description)
			continue
		}
		lists = append(lists, i)
	}
	sort.SliceStable(lists, func(a, b int) bool {
		return utf8.RuneCountInString(embeds[lists[a]].Description) < utf8.RuneCountInString(embeds[lists[b]].Description)
	})
	for n, i := range lists {
		share := max(budget/(len(lists)-n), 1)
		if utf8.RuneCountInString(embeds[i].Description) > share {
			embeds[i].Description = notifier.ChangeList(embeds[i].changes, "• ", share)
		}
		budget -= utf8.RuneCountInString(embeds[i].Description)
	}
}

func truncate(text string, maxLength int) string {
	runes := []rune(text)
	if len(runes) <= maxLength {
		return text
	}
	return string(runes[:maxLength-1]) + "…"
}
