package ffexporter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal/signer"
)

// Webhook is the exporter of your data to a webhook.
//
// It calls the EndpointURL with a POST request with the following format:
//   {
//      "meta": {
//        "hostname": "server01",
//      },
//      "events": [
//        {
//           "kind": "feature",
//           "contextKind": "anonymousUser",
//           "userKey": "14613538188334553206",
//           "creationDate": 1618909178,
//           "key": "test-flag",
//           "variation": "Default",
//           "value": false,
//           "default": false
//        },
//      ]
//    }
//
type Webhook struct {
	// EndpointURL of your webhook
	EndpointURL string
	// Secret used to sign your request body.
	Secret string
	// Meta information that you want to send to your webhook (not mandatory)
	Meta map[string]string

	httpClient internal.HTTPClient
	parsedURL  *url.URL
	init       sync.Once
}

// webhookPayload contains the body of the webhook.
type webhookPayload struct {
	// Meta are the extra information added during the configuration
	Meta map[string]string `json:"meta"`

	// events is the list of the event we send in the payload
	Events []exporter.FeatureEvent `json:"events"`
}

// Export is sending a collection of events in a webhook call.
func (f *Webhook) Export(logger *log.Logger, featureEvents []exporter.FeatureEvent) error {
	var err error
	f.init.Do(func() {
		if f.httpClient == nil {
			f.httpClient = internal.DefaultHTTPClient()
		}

		if f.Meta == nil {
			f.Meta = make(map[string]string)
		}
		// if no hostname provided we return the hostname of the current machine
		if _, ok := f.Meta["hostname"]; !ok {
			hostname, _ := os.Hostname()
			f.Meta["hostname"] = hostname
		}
		f.parsedURL, err = url.Parse(f.EndpointURL)
	})
	if err != nil {
		return err
	}

	// if we were not able to parse the URL the 1st time
	if f.parsedURL == nil {
		return errors.New("no URL available for the webhook")
	}

	body := webhookPayload{
		Meta:   f.Meta,
		Events: featureEvents,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	// if a secret is provided we sign the body and add this signature as a header.
	if f.Secret != "" {
		headers["X-Hub-Signature-256"] = []string{signer.Sign(payload, []byte(f.Secret))}
	}

	request := http.Request{
		Method: "POST",
		URL:    f.parsedURL,
		Header: headers,
		Body:   ioutil.NopCloser(bytes.NewReader(payload)),
	}

	response, err := f.httpClient.Do(&request)
	// Log if something went wrong while calling the webhook.
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode > 399 {
		return fmt.Errorf(
			"error while calling the webhook, HTTP Code %d received, response: %v", response.StatusCode, response.Body)
	}
	return nil
}

// IsBulk return false if we should directly send the data as soon as it is produce
// and true if we collect the data to send them in bulk.
func (f *Webhook) IsBulk() bool {
	return true
}
