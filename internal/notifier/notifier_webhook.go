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
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

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
	if err != nil && c.Logger != nil {
		c.Logger.Printf("[%v] error: (WebhookNotifier) impossible to read differences; %v\n", date, err)
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
		if c.Logger != nil {
			c.Logger.Printf("[%v] error: while calling webhook: %v\n", date, err)
		}
		return
	}
	defer response.Body.Close()
	if response.StatusCode > 399 && c.Logger != nil {
		c.Logger.Printf("[%v] error: while calling webhook, statusCode = %d", date, response.StatusCode)
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
