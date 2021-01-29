package cache

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/flags"
)

func Test_notificationService_getDifferences(t *testing.T) {
	type fields struct {
		Notifiers []Notifier
	}
	type args struct {
		oldCache FlagsCache
		newCache FlagsCache
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   diffCache
	}{
		{
			name: "Delete flag",
			args: args{
				oldCache: FlagsCache{
					"test-flag": flags.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
					"test-flag2": flags.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				newCache: FlagsCache{
					"test-flag": flags.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
			},
			want: diffCache{
				Deleted: map[string]flags.Flag{
					"test-flag2": {
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				Added:   map[string]flags.Flag{},
				Updated: map[string]diffUpdated{},
			},
		},
		{
			name: "Added flag",
			args: args{
				oldCache: FlagsCache{
					"test-flag": flags.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				newCache: FlagsCache{
					"test-flag": flags.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
					"test-flag2": flags.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
			},
			want: diffCache{
				Added: map[string]flags.Flag{
					"test-flag2": {
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				Deleted: map[string]flags.Flag{},
				Updated: map[string]diffUpdated{},
			},
		},
		{
			name: "Updated flag",
			args: args{
				oldCache: FlagsCache{
					"test-flag": flags.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				newCache: FlagsCache{
					"test-flag": flags.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    true,
					},
				},
			},
			want: diffCache{
				Added:   map[string]flags.Flag{},
				Deleted: map[string]flags.Flag{},
				Updated: map[string]diffUpdated{
					"test-flag": {
						Before: flags.Flag{
							Percentage: 100,
							True:       true,
							False:      false,
							Default:    false,
						},
						After: flags.Flag{
							Percentage: 100,
							True:       true,
							False:      false,
							Default:    true,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &notificationService{
				Notifiers: tt.fields.Notifiers,
				waitGroup: &sync.WaitGroup{},
			}
			got := c.getDifferences(tt.args.oldCache, tt.args.newCache)
			assert.Equal(t, tt.want, got)
		})
	}
}
