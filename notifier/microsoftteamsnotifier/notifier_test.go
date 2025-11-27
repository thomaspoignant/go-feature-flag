package microsoftteamsnotifier

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func TestMicrosoftTeamsNotifier_Notify(t *testing.T) {
	tests := []struct {
		name         string
		diff         notifier.DiffCache
		roundTripper *mockRoundTripper
		wantErr      assert.ErrorAssertionFunc
		want         string
		webhookURL   string
	}{
		{
			name: "should call webhook and have valid results",
			diff: notifier.DiffCache{
				Added: map[string]flag.Flag{
					"test-flag3": &flag.InternalFlag{
						Rules: &[]flag.Rule{
							{
								Name:  testconvert.String("rule1"),
								Query: testconvert.String("key eq \"random-key\""),
								Percentages: &map[string]float64{
									"False": 95,
									"True":  5,
								},
							},
						},
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface("default"),
							"False":   testconvert.Interface("false"),
							"True":    testconvert.Interface("test"),
						},
						DefaultRule: &flag.Rule{
							Name:            testconvert.String("defaultRule"),
							VariationResult: testconvert.String("Default"),
						},
						TrackEvents: testconvert.Bool(true),
						Disable:     testconvert.Bool(false),
						Version:     testconvert.String("1.1"),
					},
				},
				Deleted: map[string]flag.Flag{
					"test-flag": &flag.InternalFlag{
						Rules: &[]flag.Rule{
							{
								Name:  testconvert.String("rule1"),
								Query: testconvert.String("key eq \"random-key\""),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
							},
						},
						Variations: &map[string]*interface{}{
							"Default": testconvert.Interface(false),
							"False":   testconvert.Interface(false),
							"True":    testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{
							Name:            testconvert.String("defaultRule"),
							VariationResult: testconvert.String("Default"),
						},
					},
				},
				Updated: map[string]notifier.DiffUpdated{
					"test-flag2": {
						Before: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(false),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							DefaultRule: &flag.Rule{
								Name: testconvert.String("defaultRule"),
								Percentages: &map[string]float64{
									"False": 0,
									"True":  100,
								},
							},
							Experimentation: &flag.ExperimentationRollout{
								Start: testconvert.Time(time.Unix(1095379400, 0)),
								End:   testconvert.Time(time.Unix(1095371000, 0)),
							},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*interface{}{
								"Default": testconvert.Interface(true),
								"False":   testconvert.Interface(false),
								"True":    testconvert.Interface(true),
							},
							Rules: &[]flag.Rule{
								{
									Name:  testconvert.String("rule1"),
									Query: testconvert.String("key eq \"not-a-ke\""),
									Percentages: &map[string]float64{
										"False": 20,
										"True":  80,
									},
								},
							},
							DefaultRule: &flag.Rule{
								Name:            testconvert.String("defaultRule"),
								VariationResult: testconvert.String("Default"),
							},
							Disable:     testconvert.Bool(true),
							TrackEvents: testconvert.Bool(false),
							Version:     testconvert.String("1.1"),
						},
					},
				},
			},
			roundTripper: newMockRoundTripper(http.StatusAccepted),
			wantErr:      assert.NoError,
			want:         "./testdata/should_call_webhook_and_have_valid_results.json",
			webhookURL:   "https://prod-22.francecentral.logic.azure.com:443/workflows/XXXXXX/triggers/manual/paths/invoke?api-version=2016-06-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=XXX",
		},
		{
			name:         "should err if http code is superior to 399",
			roundTripper: newMockRoundTripper(http.StatusBadRequest),
			wantErr:      assert.Error,
			webhookURL:   "https://prod-22.francecentral.logic.azure.com:443/workflows/XXXXXX/triggers/manual/paths/invoke?api-version=2016-06-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=XXX",
		},
		{
			name:         "missing microsoft teams url",
			diff:         notifier.DiffCache{},
			roundTripper: newMockRoundTripper(http.StatusAccepted),
			wantErr:      assert.Error,
		},
		{
			name:         "invalid microsoft teams url",
			diff:         notifier.DiffCache{},
			roundTripper: newMockRoundTripper(http.StatusAccepted),
			wantErr:      assert.Error,
			webhookURL:   "https://{}}/workflows/XXXXXX/triggers/manual/paths/invoke?api-version=2016-06-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=XXX",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mstClient := goteamsnotify.NewTeamsClient()
			if tt.roundTripper != nil {
				mstClient.SetHTTPClient(&http.Client{Transport: tt.roundTripper})
			}

			n := &Notifier{
				MicrosoftTeamsWebhookURL: tt.webhookURL,
				teamsClient:              mstClient,
			}

			err := n.Notify(tt.diff)
			tt.wantErr(t, err)

			if err != nil {
				return
			}

			hostname, _ := os.Hostname()
			wantBody, err := os.ReadFile(tt.want)
			require.NoError(t, err)
			assert.JSONEq(
				t,
				strings.ReplaceAll(string(wantBody), "{{hostname}}", hostname),
				tt.roundTripper.requestBody,
			)
		})
	}
}

func newMockRoundTripper(status int) *mockRoundTripper {
	return &mockRoundTripper{
		response: &http.Response{
			StatusCode: status,
			Body:       nil,
		},
	}
}

type mockRoundTripper struct {
	response    *http.Response
	requestBody string
}

func (rt *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	rt.requestBody = string(body)
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return rt.response, nil
}
