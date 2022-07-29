package slacknotifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/notifier"

	"github.com/gdexlab/go-render/render"
	"github.com/r3labs/diff/v3"
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
		Body:   ioutil.NopCloser(bytes.NewReader(payload)),
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

func convertToSlackMessage(diffCache notifier.DiffCache) slackMessage {
	hostname, _ := os.Hostname()
	attachments := convertDeletedFlagsToSlackMessage(diffCache)
	attachments = append(attachments, convertUpdatedFlagsToSlackMessage(diffCache)...)
	attachments = append(attachments, convertAddedFlagsToSlackMessage(diffCache)...)
	res := slackMessage{
		Text:        fmt.Sprintf("Changes detected in your feature flag file on: *%s*", hostname),
		IconURL:     goFFLogo,
		Attachments: attachments,
	}
	return res
}

func convertDeletedFlagsToSlackMessage(diffCache notifier.DiffCache) []attachment {
	attachments := make([]attachment, 0)
	for key := range diffCache.Deleted {
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

func convertUpdatedFlagsToSlackMessage(diffCache notifier.DiffCache) []attachment {
	attachments := make([]attachment, 0)
	for key, value := range diffCache.Updated {
		attachment := attachment{
			Title:      fmt.Sprintf("✏️ Flag \"%s\" updated", key),
			Color:      colorUpdated,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
			Fields:     []Field{},
		}

		changelog, err := diff.Diff(value.Before, value.After, diff.AllowTypeMismatch(true))
		if err != nil {
			fmt.Println(err)
		}

		for _, change := range changelog {
			if change.Type == "update" {
				value := fmt.Sprintf("%s => %s", render.Render(change.From), render.Render(change.To))
				short := len(value) < longSlackAttachment
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

func convertAddedFlagsToSlackMessage(diff notifier.DiffCache) []attachment {
	attachments := make([]attachment, 0)
	for key := range diff.Added {
		attachment := attachment{
			Title:      fmt.Sprintf("🆕 Flag \"%s\" created", key),
			Color:      colorAdded,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
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

type ByTitle []Field

func (a ByTitle) Len() int           { return len(a) }
func (a ByTitle) Less(i, j int) bool { return a[i].Title < a[j].Title }
func (a ByTitle) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
