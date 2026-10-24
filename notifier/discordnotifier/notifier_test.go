package discordnotifier

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestDiscordNotifier_Notify(t *testing.T) {
	genericFlag := &flag.InternalFlag{
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
		Variations: &map[string]*any{
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
	}

	type args struct {
		diff       notifier.DiffCache
		statusCode int
		forceError bool
		url        string
	}
	type expected struct {
		err       bool
		errMsg    string
		bodyPath  string
		signature string
	}
	tests := []struct {
		name     string
		args     args
		expected expected
	}{
		{
			name: "should call webhook and have valid results",
			expected: expected{
				bodyPath: "./testdata/should_call_webhook_and_have_valid_results.json",
			},
			args: args{
				url:        "https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
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
							Variations: &map[string]*any{
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
							Variations: &map[string]*any{
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
								Variations: &map[string]*any{
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
									Start: testconvert.Time(time.Unix(1095379400, 0).UTC()),
									End:   testconvert.Time(time.Unix(1095371000, 0).UTC()),
								},
							},
							After: &flag.InternalFlag{
								Variations: &map[string]*any{
									"Default": testconvert.Interface("strDefault"),
									"False":   testconvert.Interface("strFalse"),
									"True":    testconvert.Interface("strTrue"),
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
			},
		},
		{
			name: "should not display all the flags if reached the limit of embeds",
			expected: expected{
				bodyPath: "./testdata/should_call_webhook_and_have_valid_results_limit_max.json",
			},
			args: args{
				url:        "https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
				diff: notifier.DiffCache{
					Added: map[string]flag.Flag{
						"test-flag3": genericFlag,
						"test-flag4": genericFlag,
						"test-flag5": genericFlag,
						"test-flag6": genericFlag,
					},
					Deleted: map[string]flag.Flag{
						"test-flag":   genericFlag,
						"test-flag1":  genericFlag,
						"test-flag2":  genericFlag,
						"test-flag7":  genericFlag,
						"test-flag8":  genericFlag,
						"test-flag9":  genericFlag,
						"test-flag10": genericFlag,
					},
				},
			},
		},
		{
			name: "should err if http code is superior to 399",
			expected: expected{
				err:    true,
				errMsg: "error: (Discord Notifier) webhook call failed with statusCode = 400",
			},
			args: args{
				url:        "https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusBadRequest,
				diff:       notifier.DiffCache{},
			},
		},
		{
			name: "should err if error while calling webhook",
			expected: expected{
				err:    true,
				errMsg: "error: (Discord Notifier) error calling webhook: random error",
			},
			args: args{
				url:        "https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
				forceError: true,
			},
		},
		{
			name: "missing discord url",
			expected: expected{
				err:    true,
				errMsg: "error: (Discord Notifier) invalid notifier configuration, no DiscordWebhookURL provided",
			},
			args: args{
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
			},
		},
		{
			name: "invalid discord url",
			expected: expected{
				err:    true,
				errMsg: "error: (Discord Notifier) invalid DiscordWebhookURL: https://{}discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
			},
			args: args{
				url:        "https://{}discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				statusCode: http.StatusOK,
				diff:       notifier.DiffCache{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPClient := &testutils.HTTPClientMock{
				StatusCode: tt.args.statusCode,
				ForceError: tt.args.forceError,
			}

			c := Notifier{
				DiscordWebhookURL: tt.args.url,
				httpClient:        mockHTTPClient,
			}

			err := c.Notify(tt.args.diff)

			if tt.expected.err {
				assert.ErrorContains(t, err, tt.expected.errMsg)
			} else {
				assert.NoError(t, err)
				hostname, _ := os.Hostname()
				content, err := os.ReadFile(tt.expected.bodyPath)
				require.NoError(t, err)
				expectedContent := strings.ReplaceAll(string(content), "{{hostname}}", hostname)
				assert.JSONEq(t, expectedContent, mockHTTPClient.Body)
				assert.Equal(t, tt.expected.signature, mockHTTPClient.Signature)
			}
		})
	}
}

func TestConvertUpdatedFlagsToDiscordEmbed_capsDescription(t *testing.T) {
	tests := []struct {
		name        string
		changes     int
		valueLength int
		wantLast    string
	}{
		{name: "should list every change below the limits", changes: 3, valueLength: 1, wantLast: "• Metadata `key02` set to `\"x\"`"},
		{name: "should cap the number of changes", changes: 60, valueLength: 1, wantLast: "…and 10 more changes"},
		{name: "should stay within the description length", changes: 40, valueLength: 200, wantLast: "…and 23 more changes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			after := map[string]any{}
			for i := 0; i < tt.changes; i++ {
				after[fmt.Sprintf("key%02d", i)] = strings.Repeat("x", tt.valueLength)
			}
			embeds := convertUpdatedFlagsToDiscordEmbed(notifier.DiffCache{
				Updated: map[string]notifier.DiffUpdated{
					"test-flag": {
						Before: &flag.InternalFlag{Metadata: &map[string]any{}},
						After:  &flag.InternalFlag{Metadata: &after},
					},
				},
			})
			require.Len(t, embeds, 1)
			description := embeds[0].Description
			assert.LessOrEqual(t, utf8.RuneCountInString(description), maxDiscordDescription)
			lines := strings.Split(description, "\n")
			assert.Equal(t, tt.wantLast, lines[len(lines)-1])
		})
	}
}

// otherFlag is a flag.Flag that DescribeChanges cannot describe.
type otherFlag struct{ flag.Flag }

func TestConvertUpdatedFlagsToDiscordEmbed_diffError(t *testing.T) {
	broken := otherFlag{}
	embeds := convertUpdatedFlagsToDiscordEmbed(notifier.DiffCache{
		Updated: map[string]notifier.DiffUpdated{"test-flag": {Before: broken, After: broken}},
	})
	require.Len(t, embeds, 1)
	assert.Contains(t, embeds[0].Description, notifier.ChangesUnavailable)
	assert.Contains(t, embeds[0].Description, "impossible to describe changes")
}

func TestConvertToDiscordMessage_fitsDiscordLimits(t *testing.T) {
	updated := func(changes, valueLength int) notifier.DiffUpdated {
		after := map[string]any{}
		for i := 0; i < changes; i++ {
			after[fmt.Sprintf("key%02d", i)] = strings.Repeat("x", valueLength)
		}
		return notifier.DiffUpdated{
			Before: &flag.InternalFlag{Metadata: &map[string]any{}},
			After:  &flag.InternalFlag{Metadata: &after},
		}
	}
	tests := []struct {
		name          string
		updated       map[string]notifier.DiffUpdated
		wantWhole     []string
		wantTruncated []string
	}{
		{
			name:      "should keep lists whole when they fit",
			updated:   map[string]notifier.DiffUpdated{"a": updated(3, 10), "b": updated(3, 10)},
			wantWhole: []string{"a", "b"},
		},
		{
			name:          "should cut long lists to fit the total length",
			updated:       map[string]notifier.DiffUpdated{"a": updated(40, 200), "b": updated(40, 200), "c": updated(40, 200)},
			wantTruncated: []string{"a", "b", "c"},
		},
		{
			name: "should keep short lists whole and cut the long ones",
			updated: map[string]notifier.DiffUpdated{
				"small": updated(3, 10), "big1": updated(40, 200), "big2": updated(40, 200),
			},
			wantWhole:     []string{"small"},
			wantTruncated: []string{"big1", "big2"},
		},
		{
			name:      "should shorten a title over the limit",
			updated:   map[string]notifier.DiffUpdated{strings.Repeat("k", 300): updated(1, 10)},
			wantWhole: []string{strings.Repeat("k", 300)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := convertToDiscordMessage(notifier.DiffCache{Updated: tt.updated})
			total := 0
			descriptions := map[string]string{}
			for _, e := range msg.Embeds {
				assert.LessOrEqual(t, utf8.RuneCountInString(e.Title), maxDiscordTitle)
				assert.LessOrEqual(t, utf8.RuneCountInString(e.Description), maxDiscordDescription)
				total += utf8.RuneCountInString(e.Title) + utf8.RuneCountInString(e.Description)
				descriptions[e.Title] = e.Description
			}
			assert.LessOrEqual(t, total, maxDiscordTotal)
			description := func(key string) string {
				d, ok := descriptions[truncate(fmt.Sprintf("✏️ Flag \"%s\" updated", key), maxDiscordTitle)]
				require.True(t, ok, "no embed for %s", key)
				return d
			}
			for _, key := range tt.wantWhole {
				assert.NotContains(t, description(key), "more changes")
			}
			for _, key := range tt.wantTruncated {
				assert.Contains(t, description(key), "more changes")
			}
			assert.Len(t, descriptions, len(tt.updated))
		})
	}
}
