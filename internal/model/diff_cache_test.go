package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiffCache_HasDiff(t *testing.T) {
	type fields struct {
		Deleted map[string]Flag
		Added   map[string]Flag
		Updated map[string]DiffUpdated
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "null fields",
			fields: fields{},
			want:   false,
		},
		{
			name: "empty fields",
			fields: fields{
				Deleted: map[string]Flag{},
				Added:   map[string]Flag{},
				Updated: map[string]DiffUpdated{},
			},
			want: false,
		},
		{
			name: "only Deleted",
			fields: fields{
				Deleted: map[string]Flag{
					"flag": {
						Percentage: 100,
						True:       true,
						False:      true,
						Default:    true,
					}},
				Added:   map[string]Flag{},
				Updated: map[string]DiffUpdated{},
			},
			want: true,
		},
		{
			name: "only Added",
			fields: fields{
				Added: map[string]Flag{
					"flag": {
						Percentage: 100,
						True:       true,
						False:      true,
						Default:    true,
					}},
				Deleted: map[string]Flag{},
				Updated: map[string]DiffUpdated{},
			},
			want: true,
		},
		{
			name: "only Updated",
			fields: fields{
				Added:   map[string]Flag{},
				Deleted: map[string]Flag{},
				Updated: map[string]DiffUpdated{
					"flag": {
						Before: Flag{
							Percentage: 100,
							True:       true,
							False:      true,
							Default:    true,
						},
						After: Flag{
							Percentage: 100,
							True:       true,
							False:      true,
							Default:    false,
						},
					},
				},
			},
			want: true,
		},
		{
			name: "all fields",
			fields: fields{
				Added: map[string]Flag{
					"flag": {
						Percentage: 100,
						True:       true,
						False:      true,
						Default:    true,
					},
				},
				Deleted: map[string]Flag{
					"flag": {
						Percentage: 100,
						True:       true,
						False:      true,
						Default:    true,
					}},
				Updated: map[string]DiffUpdated{
					"flag": {
						Before: Flag{
							Percentage: 100,
							True:       true,
							False:      true,
							Default:    true,
						},
						After: Flag{
							Percentage: 100,
							True:       true,
							False:      true,
							Default:    false,
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DiffCache{
				Deleted: tt.fields.Deleted,
				Added:   tt.fields.Added,
				Updated: tt.fields.Updated,
			}
			assert.Equal(t, tt.want, d.HasDiff())
		})
	}
}
