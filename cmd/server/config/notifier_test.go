package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/server/config"
)

func TestNotifierConf_IsValid(t *testing.T) {
	type fields struct {
		Kind            string
		SlackWebhookURL string
		EndpointURL     string
		Secret          string
		Meta            map[string]string
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		errValue string
	}{
		{
			name: "Invalid kind",
			fields: fields{
				Kind: "invalid",
			},
			wantErr:  true,
			errValue: "invalid notifier: kind \"invalid\" is not supported",
		},
		{
			name:     "no fields",
			fields:   fields{},
			wantErr:  true,
			errValue: "invalid notifier: kind \"\" is not supported",
		},
		{
			name: "kind slack without URL",
			fields: fields{
				Kind:            "slack",
				SlackWebhookURL: "",
			},
			wantErr:  true,
			errValue: "invalid notifier: no \"slackWebhookUrl\" property found for kind \"slack\"",
		},
		{
			name: "kind webhook without EndpointURL",
			fields: fields{
				Kind:        "webhook",
				EndpointURL: "",
			},
			wantErr:  true,
			errValue: "invalid notifier: no \"endpointUrl\" property found for kind \"webhook\"",
		},
		{
			name: "valid use-case slack",
			fields: fields{
				Kind:            "slack",
				SlackWebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
			},
			wantErr: false,
		},
		{
			name: "valid use-case webhook",
			fields: fields{
				Kind:        "webhook",
				EndpointURL: "https://hooktest.com/",
				Secret:      "xxxx",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config.NotifierConf{
				Kind:            config.NotifierKind(tt.fields.Kind),
				SlackWebhookURL: tt.fields.SlackWebhookURL,
				EndpointURL:     tt.fields.EndpointURL,
				Secret:          tt.fields.Secret,
				Meta:            tt.fields.Meta,
			}
			err := c.IsValid()
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				assert.Equal(t, tt.errValue, err.Error())
			}
		})
	}
}
