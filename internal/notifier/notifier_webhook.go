package notifier

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

func NewWebhookNotifier(logger *log.Logger,
	httpClient internal.HTTPClient,
	payloadURL string,
	secret string,
	meta map[string]string,
) (WebhookNotifier, error) {
	// Deal with meta information
	if meta == nil {
		meta = make(map[string]string)
	}

	// if no hostname provided we return the hostname of the current machine
	if _, ok := meta["hostname"]; !ok {
		hostname, _ := os.Hostname()
		meta["hostname"] = hostname
	}

	parsedURL, err := url.Parse(payloadURL)
	if err != nil {
		return WebhookNotifier{}, err
	}

	w := WebhookNotifier{
		Logger:     logger,
		PayloadURL: *parsedURL,
		Secret:     secret,
		Meta:       meta,
		HTTPClient: httpClient,
	}
	return w, nil
}

type webhookReqBody struct {
	Meta  map[string]string `json:"meta"`
	Flags model.DiffCache   `json:"flags"`
}

type WebhookNotifier struct {
	Logger     *log.Logger
	HTTPClient internal.HTTPClient
	PayloadURL url.URL
	Secret     string
	Meta       map[string]string
}

func (c *WebhookNotifier) Notify(diff model.DiffCache, wg *sync.WaitGroup) {
	defer wg.Done()
	date := time.Now().Format(time.RFC3339)

	// Create request body
	reqBody := webhookReqBody{
		Meta:  c.Meta,
		Flags: diff,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		fflog.Printf(c.Logger, "[%v] error: (WebhookNotifier) impossible to read differences; %v\n", date, err)
		return
	}

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	// if a secret is provided we sign the body and add this signature as a header.
	if c.Secret != "" {
		headers["X-Hub-Signature-256"] = []string{signPayload(payload, []byte(c.Secret))}
	}

	request := http.Request{
		Method: "POST",
		URL:    &c.PayloadURL,
		Header: headers,
		Body:   ioutil.NopCloser(bytes.NewReader(payload)),
	}
	response, err := c.HTTPClient.Do(&request)
	// Log if something went wrong while calling the webhook.
	if err != nil {
		fflog.Printf(c.Logger, "[%v] error: while calling webhook: %v\n", date, err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode > 399 {
		fflog.Printf(c.Logger, "[%v] error: while calling webhook, statusCode = %d", date, response.StatusCode)
		return
	}
}

// signPayload is using the data and the secret to compute a HMAC(SHA256) to sign the body of the request.
// so the webhook can use this signature to verify that no data have been compromised.
func signPayload(payloadBody []byte, secretToken []byte) string {
	mac := hmac.New(sha256.New, secretToken)
	_, _ = mac.Write(payloadBody)
	expectedMAC := mac.Sum(nil)
	return "sha256=" + hex.EncodeToString(expectedMAC)
}
