package ffclient

import (
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

func TestGoFeatureFlag_getNotifiers(t *testing.T) {
	urlStr := "http://webhook.com/hook"
	parsedURL, _ := url.Parse(urlStr)
	hostname, _ := os.Hostname()

	type fields struct {
		config Config
	}
	tests := []struct {
		name    string
		fields  fields
		want    []notifier.Notifier
		wantErr bool
	}{
		{
			name: "log + webhook notifier",
			fields: fields{
				config: Config{
					Logger: log.New(os.Stdout, "", 0),
					Notifiers: []NotifierConfig{
						&WebhookConfig{
							PayloadURL: urlStr,
							Secret:     "Secret",
							Meta: map[string]string{
								"my-app":   "go-ff-test",
								"hostname": hostname,
							},
						},
					},
				},
			},
			want: []notifier.Notifier{
				&notifier.LogNotifier{Logger: log.New(os.Stdout, "", 0)},
				&notifier.WebhookNotifier{
					Logger: log.New(os.Stdout, "", 0),
					HTTPClient: &http.Client{
						Timeout: 10 * time.Second,
					},
					PayloadURL: *parsedURL,
					Secret:     "Secret",
					Meta: map[string]string{
						"my-app":   "go-ff-test",
						"hostname": hostname,
					},
				},
			},
		},
		{
			name: "log + webhook notifier - deprecated webhook",
			fields: fields{
				config: Config{
					Logger: log.New(os.Stdout, "", 0),
					Webhooks: []WebhookConfig{
						{
							PayloadURL: urlStr,
							Secret:     "Secret",
							Meta: map[string]string{
								"my-app":   "go-ff-test",
								"hostname": hostname,
							},
						},
					},
				},
			},
			want: []notifier.Notifier{
				&notifier.LogNotifier{Logger: log.New(os.Stdout, "", 0)},
				&notifier.WebhookNotifier{
					Logger: log.New(os.Stdout, "", 0),
					HTTPClient: &http.Client{
						Timeout: 10 * time.Second,
					},
					PayloadURL: *parsedURL,
					Secret:     "Secret",
					Meta: map[string]string{
						"my-app":   "go-ff-test",
						"hostname": hostname,
					},
				},
			},
		},
		{
			name: "error in DNS",
			fields: fields{
				config: Config{
					Logger: log.New(os.Stdout, "", 0),
					Notifiers: []NotifierConfig{
						&WebhookConfig{
							PayloadURL: " https://example.com/hook",
							Secret:     "Secret",
							Meta: map[string]string{
								"my-app":   "go-ff-test",
								"hostname": hostname,
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNotifiers(tt.fields.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
