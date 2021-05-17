package notifier

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestLogNotifier_Notify(t *testing.T) {
	type args struct {
		diff model.DiffCache
		wg   *sync.WaitGroup
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "Flag deleted",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]model.Flag{
						"test-flag": &model.FlagData{
							Percentage: testconvert.Float64(100),
							True:       testconvert.Interface(true),
							False:      testconvert.Interface(false),
							Default:    testconvert.Interface(false),
						},
					},
					Updated: map[string]model.DiffUpdated{},
					Added:   map[string]model.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag removed",
		},
		{
			name: "Update flag",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]model.Flag{},
					Updated: map[string]model.DiffUpdated{
						"test-flag": {
							Before: &model.FlagData{
								Rule:       testconvert.String("key eq \"random-key\""),
								Percentage: testconvert.Float64(100),
								True:       testconvert.Interface(true),
								False:      testconvert.Interface(false),
								Default:    testconvert.Interface(false),
							},
							After: &model.FlagData{
								Percentage: testconvert.Float64(100),
								True:       testconvert.Interface(true),
								False:      testconvert.Interface(false),
								Default:    testconvert.Interface(false),
							},
						},
					},
					Added: map[string]model.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag updated, old=\\[percentage=100%, rule=\"key eq \"random-key\"\", true=\"true\", false=\"false\", default=\"false\", disable=\"false\"\\], new=\\[percentage=100%, true=\"true\", false=\"false\", default=\"false\", disable=\"false\"\\]",
		},
		{
			name: "Disable flag",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]model.Flag{},
					Updated: map[string]model.DiffUpdated{
						"test-flag": {
							Before: &model.FlagData{
								Rule:       testconvert.String("key eq \"random-key\""),
								Percentage: testconvert.Float64(100),
								True:       testconvert.Interface(true),
								False:      testconvert.Interface(false),
								Default:    testconvert.Interface(false),
							},
							After: &model.FlagData{
								Rule:       testconvert.String("key eq \"random-key\""),
								Disable:    testconvert.Bool(true),
								Percentage: testconvert.Float64(100),
								True:       testconvert.Interface(true),
								False:      testconvert.Interface(false),
								Default:    testconvert.Interface(false),
							},
						},
					},
					Added: map[string]model.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag is turned OFF",
		},
		{
			name: "Add flag",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]model.Flag{},
					Updated: map[string]model.DiffUpdated{},
					Added: map[string]model.Flag{
						"add-test-flag": &model.FlagData{
							Rule:       testconvert.String("key eq \"random-key\""),
							Percentage: testconvert.Float64(100),
							True:       testconvert.Interface(true),
							False:      testconvert.Interface(false),
							Default:    testconvert.Interface(false),
						},
					},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag add-test-flag added",
		},
		{
			name: "Enable flag",
			args: args{
				diff: model.DiffCache{
					Deleted: map[string]model.Flag{},
					Updated: map[string]model.DiffUpdated{
						"test-flag": {
							After: &model.FlagData{
								Rule:       testconvert.String("key eq \"random-key\""),
								Percentage: testconvert.Float64(100),
								True:       testconvert.Interface(true),
								False:      testconvert.Interface(false),
								Default:    testconvert.Interface(false),
							},
							Before: &model.FlagData{
								Rule:       testconvert.String("key eq \"random-key\""),
								Disable:    testconvert.Bool(true),
								Percentage: testconvert.Float64(100),
								True:       testconvert.Interface(true),
								False:      testconvert.Interface(false),
								Default:    testconvert.Interface(false),
							},
						},
					},
					Added: map[string]model.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "^\\[" + testutils.RFC3339Regex + "\\] flag test-flag is turned ON \\(flag=\\[percentage=100%, rule=\"key eq \"random-key\"\", true=\"true\", false=\"false\", default=\"false\", disable=\"false\"\\]\\)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logOutput, _ := ioutil.TempFile("", "")
			defer os.Remove(logOutput.Name())

			c := &LogNotifier{
				Logger: log.New(logOutput, "", 0),
			}
			tt.args.wg.Add(1)
			c.Notify(tt.args.diff, tt.args.wg)
			log, _ := ioutil.ReadFile(logOutput.Name())
			assert.Regexp(t, tt.expected, string(log))
		})
	}
}
