package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
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

	date := time.Now().Format(time.RFC3339)
	reqBody := convertToSlackMessage(diff)
	payload, err := json.Marshal(reqBody)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Printf("[%v] error: (SlackNotifier) impossible to read differences; %v\n", date, err)
		}
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
		c.Logger.Printf("[%v] error: (SlackNotifier) error: while calling webhook: %v\n", date, err)
		return
	}

	defer response.Body.Close()
	if response.StatusCode > 399 && c.Logger != nil {
		c.Logger.Printf("[%v] error: (SlackNotifier) while calling slack webhook, statusCode = %d",
			date, response.StatusCode)
		return
	}
}

func convertToSlackMessage(diff model.DiffCache) slackMessage {
	hostname, _ := os.Hostname()

	res := slackMessage{
		Text:        fmt.Sprintf("Changes detected in your feature flag file on: *%s*", hostname),
		IconURL:     goFFLogo,
		Attachments: []attachment{},
	}

	// deleted flags
	for key := range diff.Deleted {
		attachment := attachment{
			Title:      fmt.Sprintf("❌ Flag \"%s\" deleted", key),
			Color:      colorDeleted,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
		}
		res.Attachments = append(res.Attachments, attachment)
	}

	// updated flags
	for key, value := range diff.Updated {
		attachment := attachment{
			Title:      fmt.Sprintf("✏️ Flag \"%s\" updated", key),
			Color:      colorUpdated,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
			Fields:     []Field{},
		}

		if value.Before.Rule != value.After.Rule {
			attachment.Fields = append(attachment.Fields, Field{Title: "Rule", Short: false,
				Value: fmt.Sprintf("%s => %s", value.Before.Rule, value.After.Rule)})
		}

		if value.Before.Percentage != value.After.Percentage {
			attachment.Fields = append(attachment.Fields, Field{Title: "Percentage",
				Short: true,
				Value: fmt.Sprintf("%d%% => %d%%",
					int64(math.Round(value.Before.Percentage)), int64(math.Round(value.After.Percentage))),
			})
		}

		if value.Before.True != value.After.True {
			attachment.Fields = append(attachment.Fields, Field{Title: "True",
				Short: true, Value: fmt.Sprintf("%v => %v", value.Before.True, value.After.True)})
		}

		if value.Before.False != value.After.False {
			attachment.Fields = append(attachment.Fields, Field{Title: "False",
				Short: true, Value: fmt.Sprintf("%v => %v", value.Before.False, value.After.False)})
		}

		if value.Before.Default != value.After.Default {
			attachment.Fields = append(attachment.Fields, Field{Title: "False",
				Short: true, Value: fmt.Sprintf("%v => %v", value.Before.Default, value.After.Default)})
		}
		res.Attachments = append(res.Attachments, attachment)
	}

	// added flags
	for key, value := range diff.Added {
		attachment := attachment{
			Title:      fmt.Sprintf("🆕 Flag \"%s\" created", key),
			Color:      colorAdded,
			FooterIcon: goFFLogo,
			Footer:     slackFooter,
			Fields:     []Field{},
		}

		// display rule only if available
		if value.Rule != "" {
			attachment.Fields = append(attachment.Fields, Field{Title: "Rule", Short: false, Value: value.Rule})
		}

		attachment.Fields = append(attachment.Fields, Field{Title: "Percentage",
			Short: true, Value: fmt.Sprintf("%d%%", int64(math.Round(value.Percentage)))})
		attachment.Fields = append(attachment.Fields, Field{Title: "True",
			Short: true, Value: fmt.Sprintf("%v", value.True)})
		attachment.Fields = append(attachment.Fields, Field{Title: "False",
			Short: true, Value: fmt.Sprintf("%v", value.False)})
		attachment.Fields = append(attachment.Fields, Field{Title: "Default",
			Short: true, Value: fmt.Sprintf("%v", value.Default)})
		attachment.FooterIcon = goFFLogo
		res.Attachments = append(res.Attachments, attachment)
	}

	return res
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
