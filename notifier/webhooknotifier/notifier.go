package webhooknotifier

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/notifier"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/signer"
)

func NewNotifier(logger *log.Logger,
	httpClient internal.HTTPClient,
	endpointURL string,
	secret string,
	meta map[string]string,
) (Notifier, error) {
	// Deal with meta information
	if meta == nil {
		meta = make(map[string]string)
	}

	// if no hostname provided we return the hostname of the current machine
	if _, ok := meta["hostname"]; !ok {
		hostname, _ := os.Hostname()
		meta["hostname"] = hostname
	}

	parsedURL, err := url.Parse(endpointURL)
	if err != nil {
		return Notifier{}, err
	}

	w := Notifier{
		Logger:      logger,
		EndpointURL: *parsedURL,
		Secret:      secret,
		Meta:        meta,
		HTTPClient:  httpClient,
	}
	return w, nil
}

type webhookReqBody struct {
	Meta  map[string]string  `json:"meta"`
	Flags notifier.DiffCache `json:"flags"`
}

type Notifier struct {
	Logger      *log.Logger
	HTTPClient  internal.HTTPClient
	EndpointURL url.URL
	Secret      string
	Meta        map[string]string
}

func (c *Notifier) Notify(diff notifier.DiffCache, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create request body
	reqBody := webhookReqBody{
		Meta:  c.Meta,
		Flags: diff,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		fflog.Printf(c.Logger, "error: (Notifier) impossible to read differences; %v\n", err)
		return
	}

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	// if a secret is provided we sign the body and add this signature as a header.
	if c.Secret != "" {
		headers["X-Hub-Signature-256"] = []string{signer.Sign(payload, []byte(c.Secret))}
	}

	request := http.Request{
		Method: "POST",
		URL:    &c.EndpointURL,
		Header: headers,
		Body:   ioutil.NopCloser(bytes.NewReader(payload)),
	}
	response, err := c.HTTPClient.Do(&request)
	// Log if something went wrong while calling the webhook.
	if err != nil {
		fflog.Printf(c.Logger, "error: while calling webhook: %v\n", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode > 399 {
		fflog.Printf(c.Logger, "error: while calling webhook, statusCode = %d", response.StatusCode)
		return
	}
}
