package webhooknotifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/internal/signer"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

var _ notifier.Notifier = &Notifier{}

// webhookReqBody is the format we are sending to the webhook
type webhookReqBody struct {
	Meta  map[string]string  `json:"meta"`
	Flags notifier.DiffCache `json:"flags"`
}

// Notifier will call your endpoint URL with a POST request with the following format
//
//	{
//	  "meta": {
//	    "hostname": "server01"
//	  },
//	  "flags": {
//	    "deleted": {
//	      "test-flag": {
//	        "rule": "key eq \"random-key\"",
//	        "percentage": 100,
//	        "true": true,
//	        "false": false,
//	        "default": false
//	      }
//	    },
//	    "added": {
//	      "test-flag3": {
//	        "percentage": 5,
//	        "true": "test",
//	        "false": "false",
//	        "default": "default"
//	      }
//	    },
//	    "updated": {
//	      "test-flag2": {
//	        "old_value": {
//	          "rule": "key eq \"not-a-key\"",
//	          "percentage": 100,
//	          "true": true,
//	          "false": false,
//	          "default": false
//	        },
//	        "new_value": {
//	          "disable": true,
//	          "rule": "key eq \"not-a-key\"",
//	          "percentage": 100,
//	          "true": true,
//	          "false": false,
//	          "default": false
//	        }
//	      }
//	    }
//	  }
//	}
type Notifier struct {
	// EndpointURL of your webhook
	EndpointURL string

	// Optional: Secret used to sign your request body.
	Secret string

	// Meta (optional) information that you want to send to your webhook
	Meta map[string]string

	// Headers (optional) the list of Headers to send to the endpoint
	Headers map[string][]string

	httpClient internal.HTTPClient
	init       sync.Once
}

// Notify is the notifying all the changes to the notifier.
func (c *Notifier) Notify(diff notifier.DiffCache) error {
	if c.EndpointURL == "" {
		return fmt.Errorf(
			"invalid notifier configuration, no endpointURL provided for the webhook notifier",
		)
	}

	// init the notifier
	c.init.Do(func() {
		if c.httpClient == nil {
			c.httpClient = internal.DefaultHTTPClient()
		}

		if c.Meta == nil {
			c.Meta = make(map[string]string)
		}

		// if no hostname provided we return the hostname of the current machine
		if _, ok := c.Meta["hostname"]; !ok {
			hostname, _ := os.Hostname()
			c.Meta["hostname"] = hostname
		}
	})

	endpointURL, err := url.Parse(c.EndpointURL)
	if err != nil {
		return fmt.Errorf("error: (Webhook Notifier) invalid EnpointURL:%v", c.EndpointURL)
	}

	// Create request body
	reqBody := webhookReqBody{
		Meta:  c.Meta,
		Flags: diff,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error: (Webhook Notifier) impossible to read differences; %v", err)
	}

	if c.Headers == nil {
		c.Headers = map[string][]string{}
	}
	c.Headers["Content-Type"] = []string{"application/json"}

	// if a secret is provided we sign the body and add this signature as a header.
	if c.Secret != "" {
		c.Headers["X-Hub-Signature-256"] = []string{signer.Sign(payload, []byte(c.Secret))}
	}

	request := http.Request{
		Method: "POST",
		URL:    endpointURL,
		Header: c.Headers,
		Body:   io.NopCloser(bytes.NewReader(payload)),
	}
	response, err := c.httpClient.Do(&request)
	// Log if something went wrong while calling the webhook.
	if err != nil {
		return fmt.Errorf("error: while calling webhook: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode > 399 {
		return fmt.Errorf("error: while calling webhook, statusCode = %d", response.StatusCode)
	}

	return nil
}
