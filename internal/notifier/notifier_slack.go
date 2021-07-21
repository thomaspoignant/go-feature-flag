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
	"sort"
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
const longSlackAttachment = 35

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
