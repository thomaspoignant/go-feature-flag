package cache

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/flags"
	"github.com/thomaspoignant/go-feature-flag/testutil"
)

func TestLogNotifier_Notify(t *testing.T) {
	type args struct {
		diff diffCache
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
				diff: diffCache{
					Deleted: map[string]flags.Flag{
						"test-flag": {
							Percentage: 100,
							True:       true,
							False:      false,
							Default:    false,
						},
					},
					Updated: map[string]diffUpdated{},
					Added:   map[string]flags.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag test-flag removed",
		},
		{
			name: "Update flag",
			args: args{
				diff: diffCache{
					Deleted: map[string]flags.Flag{},
					Updated: map[string]diffUpdated{
						"test-flag": {
							Before: flags.Flag{
								Rule:       "key eq \"random-key\"",
								Percentage: 100,
								True:       true,
								False:      false,
								Default:    false,
							},
							After: flags.Flag{
								Percentage: 100,
								True:       true,
								False:      false,
								Default:    false,
							},
						},
					},
					Added: map[string]flags.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag test-flag updated, old=\\[percentage=100%, rule=\"key eq \"random-key\"\", true=\"true\", false=\"false\", true=\"false\", disable=\"false\"\\], new=\\[percentage=100%, true=\"true\", false=\"false\", true=\"false\", disable=\"false\"\\]",
		},
		{
			name: "Disable flag",
			args: args{
				diff: diffCache{
					Deleted: map[string]flags.Flag{},
					Updated: map[string]diffUpdated{
						"test-flag": {
							Before: flags.Flag{
								Rule:       "key eq \"random-key\"",
								Percentage: 100,
								True:       true,
								False:      false,
								Default:    false,
							},
							After: flags.Flag{
								Rule:       "key eq \"random-key\"",
								Disable:    true,
								Percentage: 100,
								True:       true,
								False:      false,
								Default:    false,
							},
						},
					},
					Added: map[string]flags.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag test-flag is turned OFF",
		},
		{
			name: "Add flag",
			args: args{
				diff: diffCache{
					Deleted: map[string]flags.Flag{},
					Updated: map[string]diffUpdated{},
					Added: map[string]flags.Flag{
						"add-test-flag": {
							Rule:       "key eq \"random-key\"",
							Percentage: 100,
							True:       true,
							False:      false,
							Default:    false,
						},
					},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag add-test-flag added",
		},
		{
			name: "Enable flag",
			args: args{
				diff: diffCache{
					Deleted: map[string]flags.Flag{},
					Updated: map[string]diffUpdated{
						"test-flag": {
							After: flags.Flag{
								Rule:       "key eq \"random-key\"",
								Percentage: 100,
								True:       true,
								False:      false,
								Default:    false,
							},
							Before: flags.Flag{
								Rule:       "key eq \"random-key\"",
								Disable:    true,
								Percentage: 100,
								True:       true,
								False:      false,
								Default:    false,
							},
						},
					},
					Added: map[string]flags.Flag{},
				},
				wg: &sync.WaitGroup{},
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag test-flag is turned ON \\(flag=\\[percentage=100%, rule=\"key eq \"random-key\"\", true=\"true\", false=\"false\", true=\"false\", disable=\"false\"\\]\\)",
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
			fmt.Println(string(log))
			assert.Regexp(t, tt.expected, string(log))
		})
	}
}
