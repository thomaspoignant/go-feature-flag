package model_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

func TestExperimentation_String(t *testing.T) {
	type fields struct {
		StartDate *time.Time
		EndDate   *time.Time
		Start     *time.Time
		End       *time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "both dates - deprecated fields",
			fields: fields{
				StartDate: testconvert.Time(time.Unix(1095379400, 0)),
				EndDate:   testconvert.Time(time.Unix(1095379500, 0)),
			},
			want: "start:[2004-09-17T00:03:20Z] end:[2004-09-17T00:05:00Z]",
		},
		{
			name: "only start date - deprecated fields",
			fields: fields{
				StartDate: testconvert.Time(time.Unix(1095379400, 0)),
			},
			want: "start:[2004-09-17T00:03:20Z]",
		},
		{
			name: "only end date - deprecated fields",
			fields: fields{
				EndDate: testconvert.Time(time.Unix(1095379500, 0)),
			},
			want: "end:[2004-09-17T00:05:00Z]",
		},
		{
			name: "both dates",
			fields: fields{
				Start: testconvert.Time(time.Unix(1095379400, 0)),
				End:   testconvert.Time(time.Unix(1095379500, 0)),
			},
			want: "start:[2004-09-17T00:03:20Z] end:[2004-09-17T00:05:00Z]",
		},
		{
			name: "only start date",
			fields: fields{
				Start: testconvert.Time(time.Unix(1095379400, 0)),
			},
			want: "start:[2004-09-17T00:03:20Z]",
		},
		{
			name: "only end date",
			fields: fields{
				End: testconvert.Time(time.Unix(1095379500, 0)),
			},
			want: "end:[2004-09-17T00:05:00Z]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := model.Experimentation{
				StartDate: tt.fields.StartDate,
				EndDate:   tt.fields.EndDate,
				End:       tt.fields.End,
				Start:     tt.fields.Start,
			}
			got := e.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRollout_String(t *testing.T) {
	tests := []struct {
		name    string
		rollout model.Rollout
		want    string
	}{
		{
			name: "experimentation",
			rollout: model.Rollout{Experimentation: &model.Experimentation{
				Start: testconvert.Time(time.Unix(1095379400, 0)),
				End:   testconvert.Time(time.Unix(1095379500, 0)),
			}},
			want: "experimentation: start:[2004-09-17T00:03:20Z] end:[2004-09-17T00:05:00Z]",
		},
		{
			name:    "empty",
			rollout: model.Rollout{},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rollout.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
