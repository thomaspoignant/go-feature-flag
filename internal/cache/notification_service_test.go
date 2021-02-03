package cache

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

func Test_notificationService_getDifferences(t *testing.T) {
	type fields struct {
		Notifiers []notifier.Notifier
	}
	type args struct {
		oldCache FlagsCache
		newCache FlagsCache
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   model.DiffCache
	}{
		{
			name: "Delete flag",
			args: args{
				oldCache: FlagsCache{
					"test-flag": model.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
					"test-flag2": model.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				newCache: FlagsCache{
					"test-flag": model.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
			},
			want: model.DiffCache{
				Deleted: map[string]model.Flag{
					"test-flag2": {
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				Added:   map[string]model.Flag{},
				Updated: map[string]model.DiffUpdated{},
			},
		},
		{
			name: "Added flag",
			args: args{
				oldCache: FlagsCache{
					"test-flag": model.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				newCache: FlagsCache{
					"test-flag": model.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
					"test-flag2": model.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
			},
			want: model.DiffCache{
				Added: map[string]model.Flag{
					"test-flag2": {
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				Deleted: map[string]model.Flag{},
				Updated: map[string]model.DiffUpdated{},
			},
		},
		{
			name: "Updated flag",
			args: args{
				oldCache: FlagsCache{
					"test-flag": model.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    false,
					},
				},
				newCache: FlagsCache{
					"test-flag": model.Flag{
						Percentage: 100,
						True:       true,
						False:      false,
						Default:    true,
					},
				},
			},
			want: model.DiffCache{
				Added:   map[string]model.Flag{},
				Deleted: map[string]model.Flag{},
				Updated: map[string]model.DiffUpdated{
					"test-flag": {
						Before: model.Flag{
							Percentage: 100,
							True:       true,
							False:      false,
							Default:    false,
						},
						After: model.Flag{
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
