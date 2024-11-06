package microsoftteamsnotifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/luci/go-render/render"
	"github.com/r3labs/diff/v3"
	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

const (
	colorDeleted                 = "#FF0000"
	colorUpdated                 = "#FFA500"
	colorAdded                   = "#008000"
	goFFLogo                     = "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/logo_128.png"
	microsoftteamsFooter         = "go-feature-flag"
	longMicrosoftTeamsAttachment = 100
)

type Notifier struct {
	MicrosoftTeamsWebhookURL string

	httpClient internal.HTTPClient
	init       sync.Once
}

func (c *Notifier) Notify(diff notifier.DiffCache) error {
	if c.MicrosoftTeamsWebhookURL == "" {
		return fmt.Errorf("error: (Microsoft Teams Notifier) invalid notifier configuration, no " +
			"MicrosoftTeamsWebhookURL provided for the microsoft teams notifier")
	}

	// init the notifier
	c.init.Do(func() {
		if c.httpClient == nil {
			c.httpClient = internal.DefaultHTTPClient()
		}
	})

	microsoftteamsURL, err := url.Parse(c.MicrosoftTeamsWebhookURL)
	if err != nil {
		return fmt.Errorf("error: (Microsoft Teams Notifier) invalid MicrosoftTeamsWebhookURL: %v",
			c.MicrosoftTeamsWebhookURL)
	}

	reqBody := convertToMicrosoftTeamsMessage(diff)
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error: (Microsoft Teams Notifier) impossible to read differences; %v", err)
	}
	microsoftTeamAccessToken := os.Getenv("MICROSOFT_TEAMS_ACCESS_TOKEN")
	request := http.Request{
		Method: http.MethodPost,
		URL:    microsoftteamsURL,
		Body:   io.NopCloser(bytes.NewReader(payload)),
		Header: map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {"Bearer " + microsoftTeamAccessToken},
		},
	}
	response, err := c.httpClient.Do(&request)
	if err != nil {
		return fmt.Errorf("error: (Microsoft Teams Notifier) error: while calling webhook: %v", err)
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode > 399 {
		return fmt.Errorf("error: (Microsoft Teams Notifier) while calling microsoft teams webhook, statusCode = %d",
			response.StatusCode)
	}

	return nil
}

func convertToMicrosoftTeamsMessage(diffCache notifier.DiffCache) AdaptiveCard {
	hostname, _ := os.Hostname()
	attachments := convertDeletedFlagsToMicrosoftTeamsMessage(diffCache)
	attachments = append(attachments, convertUpdatedFlagsToMicrosoftTeamsMessage(diffCache)...)
	attachments = append(attachments, convertAddedFlagsToMicrosoftTeamsMessage(diffCache)...)
	sections := attachmentsToSections(attachments)
	res := AdaptiveCard{
		Type:     "MessageCard",
		Context:  "https://schema.org/extensions",
		Summary:  fmt.Sprintf("Changes detected in your feature flag file on: *%s*", hostname),
		Sections: sections,
	}
	return res
}

func convertDeletedFlagsToMicrosoftTeamsMessage(diffCache notifier.DiffCache) []attachment {
	attachments := make([]attachment, 0)
	for key := range diffCache.Deleted {
		attachment := attachment{
			Title:      fmt.Sprintf("❌ Flag \"%s\" deleted", key),
			Color:      colorDeleted,
			FooterIcon: goFFLogo,
			Footer:     microsoftteamsFooter,
		}
		attachments = append(attachments, attachment)
	}
	return attachments
}

func convertUpdatedFlagsToMicrosoftTeamsMessage(diffCache notifier.DiffCache) []attachment {
	attachments := make([]attachment, 0)
	for key, value := range diffCache.Updated {
		attachment := attachment{
			Title:      fmt.Sprintf("✏️ Flag \"%s\" updated", key),
			Color:      colorUpdated,
			FooterIcon: goFFLogo,
			Footer:     microsoftteamsFooter,
			Fields:     []Field{},
		}

		changelog, _ := diff.Diff(value.Before, value.After, diff.AllowTypeMismatch(true))
		for _, change := range changelog {
			if change.Type == "update" {
				value := fmt.Sprintf("%s => %s", render.Render(change.From), render.Render(change.To))
				short := len(value) < longMicrosoftTeamsAttachment
				attachment.Fields = append(
					attachment.Fields,
					Field{Title: strings.Join(change.Path, "."), Short: short, Value: value},
				)
			}
		}

		sort.Sort(ByTitle(attachment.Fields))

		attachments = append(attachments, attachment)
	}
	return attachments
}

func convertAddedFlagsToMicrosoftTeamsMessage(diff notifier.DiffCache) []attachment {
	attachments := make([]attachment, 0)
	for key := range diff.Added {
		attachment := attachment{
			Title:      fmt.Sprintf("🆕 Flag \"%s\" created", key),
			Color:      colorAdded,
			FooterIcon: goFFLogo,
			Footer:     microsoftteamsFooter,
		}
		attachments = append(attachments, attachment)
	}
	return attachments
}

func attachmentsToSections(attachments []attachment) []Section {
	sections := make([]Section, len(attachments))
	for i, att := range attachments {
		facts := make([]Fact, len(att.Fields))
		for j, field := range att.Fields {
			facts[j] = Fact(field)
		}

		sections[i] = Section{
			Title: att.Title,
			Color: att.Color,
			Facts: facts,
		}
	}
	return sections
}

type AdaptiveCard struct {
	Type     string    `json:"@type"`
	Context  string    `json:"@context"`
	Summary  string    `json:"summary"`
	Sections []Section `json:"sections"`
}

type Section struct {
	Title string `json:"title,omitempty"`
	Text  string `json:"text,omitempty"`
	Color string `json:"color,omitempty"`
	Facts []Fact `json:"facts,omitempty"`
}

type Fact struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}
type attachment struct {
	Title      string  `json:"title"`
	Color      string  `json:"color"`
	FooterIcon string  `json:"footer_icon"`
	Footer     string  `json:"footer"`
	Fields     []Field `json:"fields,omitempty"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type ByTitle []Field

func (a ByTitle) Len() int           { return len(a) }
func (a ByTitle) Less(i, j int) bool { return a[i].Title < a[j].Title }
func (a ByTitle) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
