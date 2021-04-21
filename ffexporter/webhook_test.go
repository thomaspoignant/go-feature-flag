package ffexporter

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestWebhook_IsBulk(t *testing.T) {
	exporter := Webhook{}
	assert.True(t, exporter.IsBulk(), "Webhook exporter is not a bulk exporter")
}

func TestWebhook_Export(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	type fields struct {
		EndpointURL string
		Secret      string
		Meta        map[string]string
		httpClient  testutils.HTTPClientMock
	}
	type args struct {
		logger        *log.Logger
		featureEvents []exporter.FeatureEvent
	}
	type expected struct {
		bodyFilePath string
		signHeader   string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		expected expected
	}{
		{
			name: "Invalid EndpointURL",
			fields: fields{
				EndpointURL: " http://invalid.com",
			},
			wantErr: true,
		},
		{
			name: "valid without signature",
			fields: fields{
				EndpointURL: "http://valid.com/webhook",
				httpClient:  testutils.HTTPClientMock{StatusCode: 200, ForceError: false},
				Meta:        map[string]string{"hostname": "hostname"},
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.FeatureEvent{
					{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false},
					{Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false},
				},
			},
			expected: expected{
				bodyFilePath: "../testdata/ffexporter/webhook/valid_without_signature.json",
				signHeader:   "",
			},
			wantErr: false,
		},
		{
			name: "valid with signature",
			fields: fields{
				EndpointURL: "http://valid.com/webhook",
				httpClient:  testutils.HTTPClientMock{StatusCode: 200, ForceError: false},
				Meta:        map[string]string{"hostname": "hostname"},
				Secret:      "secret",
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.FeatureEvent{
					{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false},
					{Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false},
				},
			},
			expected: expected{
				bodyFilePath: "../testdata/ffexporter/webhook/valid_with_signature.json",
				signHeader:   "sha256=a52098f9d5b3dc26c0095b690ca8a62348e7142fb4c96a181e3e80d134cd9084",
			},
			wantErr: false,
		},
		{
			name: "http error while calling webhook",
			fields: fields{
				EndpointURL: "http://valid.com/webhook",
				httpClient:  testutils.HTTPClientMock{StatusCode: 400, ForceError: false},
				Meta:        map[string]string{"hostname": "hostname"},
				Secret:      "secret",
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.FeatureEvent{
					{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false},
					{Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false},
				},
			},
			wantErr: true,
		},
		{
			name: "error in the during the Do call",
			fields: fields{
				EndpointURL: "http://valid.com/webhook",
				httpClient:  testutils.HTTPClientMock{StatusCode: 200, ForceError: true},
				Meta:        map[string]string{"hostname": "hostname"},
				Secret:      "secret",
			},
			args: args{
				logger: logger,
				featureEvents: []exporter.FeatureEvent{
					{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
						Variation: "Default", Value: "YO", Default: false},
					{Kind: "feature", ContextKind: "anonymousUser", UserKey: "EFGH", CreationDate: 1617970701, Key: "random-key",
						Variation: "Default", Value: "YO2", Default: false},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Webhook{
				EndpointURL: tt.fields.EndpointURL,
				Secret:      tt.fields.Secret,
				Meta:        tt.fields.Meta,
				httpClient:  &tt.fields.httpClient,
			}
			err := f.Export(tt.args.logger, tt.args.featureEvents)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.expected.bodyFilePath != "" {
				c, _ := ioutil.ReadFile(tt.expected.bodyFilePath)
				assert.JSONEq(t, string(c), tt.fields.httpClient.Body)
			}

			if tt.expected.signHeader != "" {
				assert.Equal(t, tt.expected.signHeader, tt.fields.httpClient.Signature)
			}
		})
	}
}

func TestWebhook_Export_impossibleToParse(t *testing.T) {
	f := &Webhook{
		EndpointURL: " http://invalid.com/",
	}

	err := f.Export(log.New(os.Stdout, "", 0), []exporter.FeatureEvent{})
	assert.EqualError(t, err, "parse \" http://invalid.com/\": first path segment in URL cannot contain colon")

	// We should parse the URL only once
	err = f.Export(log.New(os.Stdout, "", 0), []exporter.FeatureEvent{})
	assert.EqualError(t, err, "no URL available for the webhook")
}
