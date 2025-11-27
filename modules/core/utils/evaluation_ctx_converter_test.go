package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func Test_ConvertEvaluationCtxFromReq(t *testing.T) {
	type fields struct {
		Key    string
		Custom map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   ffcontext.Context
	}{
		{
			name: "simple case",
			fields: fields{
				Key: "2323f37b-eef7-4bbc-856f-7d16c67de3ae",
				Custom: map[string]interface{}{
					"company_name": "go feature flag",
				},
			},
			want: ffcontext.
				NewEvaluationContextBuilder("2323f37b-eef7-4bbc-856f-7d16c67de3ae").
				AddCustom("company_name", "go feature flag").
				Build(),
		},
		{
			name: "should return a float if the value is a float",
			fields: fields{
				Key: "2323f37b-eef7-4bbc-856f-7d16c67de3ae",
				Custom: map[string]interface{}{
					"company_name": "go feature flag",
					"company_id":   1.1,
				},
			},
			want: ffcontext.
				NewEvaluationContextBuilder("2323f37b-eef7-4bbc-856f-7d16c67de3ae").
				AddCustom("company_name", "go feature flag").
				AddCustom("company_id", 1.1).
				Build(),
		},
		{
			name: "should return an int if the value is a float but without decimal",
			fields: fields{
				Key: "2323f37b-eef7-4bbc-856f-7d16c67de3ae",
				Custom: map[string]interface{}{
					"company_name": "go feature flag",
					"company_id":   1.0,
				},
			},
			want: ffcontext.
				NewEvaluationContextBuilder("2323f37b-eef7-4bbc-856f-7d16c67de3ae").
				AddCustom("company_name", "go feature flag").
				AddCustom("company_id", 1).
				Build(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.ConvertEvaluationCtxFromRequest(tt.fields.Key, tt.fields.Custom)
			assert.Equal(t, tt.want, got)
		})
	}
}
