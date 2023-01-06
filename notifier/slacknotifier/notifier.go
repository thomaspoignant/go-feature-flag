package slacknotifier

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

	"github.com/thomaspoignant/go-feature-flag/notifier"

	"github.com/thomaspoignant/go-feature-flag/internal"
)

const (
	goFFLogo            = "https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/logo_128.png"
	slackFooter         = "go-feature-flag"
	colorDeleted        = "#FF0000"
	colorUpdated        = "#FFA500"
	colorAdded          = "#008000"
	longSlackAttachment = 35
)

type Notifier struct {
	SlackWebhookURL string

	httpClient internal.HTTPClient
	init       sync.Once
}

func (c *Notifier) Notify(diff notifier.DiffCache, wg *sync.WaitGroup) error {
	defer wg.Done()

	if c.SlackWebhookURL == "" {
		return fmt.Errorf("error: (Slack Notifier) invalid notifier configuration, no " +
			"SlackWebhookURL provided for the slack notifier")
	}

	// init the notifier
	c.init.Do(func() {
		if c.httpClient == nil {
			c.httpClient = internal.DefaultHTTPClient()
		}
	})

	slackURL, err := url.Parse(c.SlackWebhookURL)
	if err != nil {
		return fmt.Errorf("error: (Slack Notifier) invalid SlackWebhookURL: %v", c.SlackWebhookURL)
	}

	reqBody := convertToSlackMessage(diff)
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error: (Slack Notifier) impossible to read differences; %v", err)
	}
	request := http.Request{
		Method: http.MethodPost,
		URL:    slackURL,
		Body:   io.NopCloser(bytes.NewReader(payload)),
		Header: map[string][]string{"Content-type": {"application/json"}},
	}
	response, err := c.httpClient.Do(&request)
	if err != nil {
		return fmt.Errorf("error: (Slack Notifier) error: while calling webhook: %v", err)
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode > 399 {
		return fmt.Errorf("error: (Slack Notifier) while calling slack webhook, statusCode = %d",
			response.StatusCode)
	}

	return nil
}

func convertToSlackMessage(diff notifier.DiffCache) slackMessage {
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

func convertDeletedFlagsToSlackMessage(diff notifier.DiffCache) []attachment {
	attachments := make([]attachment, 0)
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

func convertUpdatedFlagsToSlackMessage(diff notifier.DiffCache) []attachment {
	attachments := make([]attachment, 0)
	for key, value := range diff.Updated {
		attachment := attachment{
			Title:      fmt.Sprintf("✏️ Flag \"%s\" updated", key),
			Color:      colorUpdated,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
			Fields:     []Field{},
		}

		before := value.Before.GetRawValues()
		after := value.After.GetRawValues()
		sortedKey := sortedKeys(before)
		for _, bKey := range sortedKey {
			if before[bKey] != after[bKey] {
				// format output if empty
				if before[bKey] == "" {
					before[bKey] = "<empty>"
				}
				if after[bKey] == "" {
					after[bKey] = "<empty>"
				}

				value := fmt.Sprintf("%v => %v", before[bKey], after[bKey])
				short := len(value) < longSlackAttachment
				attachment.Fields = append(attachment.Fields, Field{Title: bKey, Short: short, Value: value})
			}
		}
		attachments = append(attachments, attachment)
	}
	return attachments
}

func convertAddedFlagsToSlackMessage(diff notifier.DiffCache) []attachment {
	attachments := make([]attachment, 0)
	for key, value := range diff.Added {
		attachment := attachment{
			Title:      fmt.Sprintf("🆕 Flag \"%s\" created", key),
			Color:      colorAdded,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
			Fields:     []Field{},
		}

		rawValues := value.GetRawValues()
		sortedKey := sortedKeys(rawValues)
		for _, bKey := range sortedKey {
			if rawValues[bKey] != "" {
				value := fmt.Sprintf("%v", rawValues[bKey])
				short := len(value) < longSlackAttachment
				attachment.Fields = append(attachment.Fields, Field{Title: bKey, Short: short, Value: value})
			}
		}
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

func sortedKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
