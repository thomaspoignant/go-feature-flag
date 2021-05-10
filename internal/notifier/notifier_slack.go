package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

const goFFLogo = "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/logo_128.png"
const slackFooter = "go-feature-flag"
const colorDeleted = "#FF0000"
const colorUpdated = "#FFA500"
const colorAdded = "#008000"

func NewSlackNotifier(logger *log.Logger, httpClient internal.HTTPClient, webhookURL string) SlackNotifier {
	slackURL, _ := url.Parse(webhookURL)
	return SlackNotifier{
		Logger:     logger,
		HTTPClient: httpClient,
		WebhookURL: *slackURL,
	}
}

type SlackNotifier struct {
	Logger     *log.Logger
	HTTPClient internal.HTTPClient
	WebhookURL url.URL
}

func (c *SlackNotifier) Notify(diff model.DiffCache, wg *sync.WaitGroup) {
	defer wg.Done()

	reqBody := convertToSlackMessage(diff)
	payload, err := json.Marshal(reqBody)
	if err != nil {
		fflog.Printf(c.Logger, "error: (SlackNotifier) impossible to read differences; %v\n", err)
		return
	}
	request := http.Request{
		Method: http.MethodPost,
		URL:    &c.WebhookURL,
		Body:   ioutil.NopCloser(bytes.NewReader(payload)),
		Header: map[string][]string{"Content-type": {"application/json"}},
	}
	response, err := c.HTTPClient.Do(&request)
	if err != nil {
		fflog.Printf(c.Logger, "error: (SlackNotifier) error: while calling webhook: %v\n", err)
		return
	}

	defer response.Body.Close()
	if response.StatusCode > 399 {
		fflog.Printf(c.Logger, "error: (SlackNotifier) while calling slack webhook, statusCode = %d",
			response.StatusCode)
		return
	}
}

func convertToSlackMessage(diff model.DiffCache) slackMessage {
	hostname, _ := os.Hostname()
	attachments := convertDeletedFlagsToSlackMessage(diff)
	attachments = append(attachments, convertUpdatedFlagsToSlackMessage(diff)...)
	attachments = append(attachments, convertAddedFlagsToSlackMessage(diff)...)
	res := slackMessage{
		Text:        fmt.Sprintf("Changes detected in your feature flag file on: *%s*", hostname),
		IconURL:     goFFLogo,
		Attachments: attachments,
	}
	return res
}

func convertDeletedFlagsToSlackMessage(diff model.DiffCache) []attachment {
	var attachments = make([]attachment, 0)
	for key := range diff.Deleted {
		attachment := attachment{
			Title:      fmt.Sprintf("❌ Flag \"%s\" deleted", key),
			Color:      colorDeleted,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
		}
		attachments = append(attachments, attachment)
	}
	return attachments
}

func convertUpdatedFlagsToSlackMessage(diff model.DiffCache) []attachment {
	var attachments = make([]attachment, 0)
	// updated flags - use reflection to list the flags
	for key, value := range diff.Updated {
		attachment := attachment{
			Title:      fmt.Sprintf("✏️ Flag \"%s\" updated", key),
			Color:      colorUpdated,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
			Fields:     []Field{},
		}

		before := value.Before
		after := value.After
		const compareFormat = "%v => %v"

		// rule
		if before.GetRule() != after.GetRule() {
			attachment.Fields = append(attachment.Fields, Field{Title: "Rule", Short: false,
				Value: fmt.Sprintf(compareFormat, before.GetRule(), after.GetRule())})
		}

		// Percentage
		if before.GetPercentage() != after.GetPercentage() {
			attachment.Fields = append(attachment.Fields, Field{Title: "Percentage", Short: true,
				Value: fmt.Sprintf(compareFormat, before.GetPercentage(), after.GetPercentage())})
		}

		// True
		if before.GetTrue() != after.GetTrue() {
			attachment.Fields = append(attachment.Fields, Field{Title: "True", Short: true,
				Value: fmt.Sprintf(compareFormat, before.GetTrue(), after.GetTrue())})
		}

		// False
		if before.GetFalse() != after.GetFalse() {
			attachment.Fields = append(attachment.Fields, Field{Title: "False", Short: true,
				Value: fmt.Sprintf(compareFormat, before.GetFalse(), after.GetFalse())})
		}

		// Default
		if before.GetDefault() != after.GetDefault() {
			attachment.Fields = append(attachment.Fields, Field{Title: "Default", Short: true,
				Value: fmt.Sprintf(compareFormat, before.GetDefault(), after.GetDefault())})
		}

		// TrackEvents
		if before.GetTrackEvents() != after.GetTrackEvents() {
			attachment.Fields = append(attachment.Fields, Field{Title: "TrackEvents", Short: true,
				Value: fmt.Sprintf(compareFormat, before.GetTrackEvents(), after.GetTrackEvents())})
		}

		// Disable
		if before.GetDisable() != after.GetDisable() {
			attachment.Fields = append(attachment.Fields, Field{Title: "Disable", Short: true,
				Value: fmt.Sprintf(compareFormat, before.GetDisable(), after.GetDisable())})
		}

		// Rollout
		if before.GetRollout() != after.GetRollout() {
			attachment.Fields = append(attachment.Fields, Field{Title: "Rollout", Short: false,
				Value: fmt.Sprintf(compareFormat, before.GetRollout(), after.GetRollout())})
		}
		attachments = append(attachments, attachment)
	}
	return attachments
}

func convertAddedFlagsToSlackMessage(diff model.DiffCache) []attachment {
	var attachments = make([]attachment, 0)
	for key, value := range diff.Added {
		attachment := attachment{
			Title:      fmt.Sprintf("🆕 Flag \"%s\" created", key),
			Color:      colorAdded,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
			Fields:     []Field{},
		}

		attachment.Fields = append(attachment.Fields, Field{Title: "Rule", Short: false,
			Value: fmt.Sprintf("%v", value.GetRule())})
		attachment.Fields = append(attachment.Fields, Field{Title: "Percentage", Short: true,
			Value: fmt.Sprintf("%v", value.GetPercentage())})
		attachment.Fields = append(attachment.Fields, Field{Title: "True", Short: true,
			Value: fmt.Sprintf("%v", value.GetTrue())})
		attachment.Fields = append(attachment.Fields, Field{Title: "False", Short: true,
			Value: fmt.Sprintf("%v", value.GetFalse())})
		attachment.Fields = append(attachment.Fields, Field{Title: "Default", Short: true,
			Value: fmt.Sprintf("%v", value.GetDefault())})
		attachment.Fields = append(attachment.Fields, Field{Title: "TrackEvents", Short: true,
			Value: fmt.Sprintf("%v", value.GetTrackEvents())})
		attachment.Fields = append(attachment.Fields, Field{Title: "Disable", Short: true,
			Value: fmt.Sprintf("%v", value.GetDisable())})
		attachments = append(attachments, attachment)
	}
	return attachments
}

type slackMessage struct {
	IconURL     string       `json:"icon_url"`
	Text        string       `json:"text"`
	Attachments []attachment `json:"attachments"`
}

type attachment struct {
	Color      string  `json:"color"`
	Title      string  `json:"title"`
	Fields     []Field `json:"fields"`
	FooterIcon string  `json:"footer_icon,omitempty"`
	Footer     string  `json:"footer,omitempty"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}
