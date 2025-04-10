package ffcontext_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

func TestUser_AddCustomAttribute(t *testing.T) {
	type args struct {
		name  string
		value interface{}
	}
	tests := []struct {
		name string
		user ffcontext.EvaluationContext
		args args
		want map[string]interface{}
	}{
		{
			name: "trying to add nil value",
			user: ffcontext.NewEvaluationContext("123"),
			args: args{},
			want: map[string]interface{}{},
		},
		{
			name: "add valid element",
			user: ffcontext.NewEvaluationContext("123"),
			args: args{
				name:  "test",
				value: "test",
			},
			want: map[string]interface{}{
				"test": "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.user.AddCustomAttribute(tt.args.name, tt.args.value)
			assert.Equal(t, tt.want, tt.user.GetCustom())
		})
	}
}

func Test_ExtractGOFFProtectedFields(t *testing.T) {
	tests := []struct {
		name string
		ctx  ffcontext.EvaluationContext
		want ffcontext.GoffContextSpecifics
	}{
		{
			name: "context goff specifics as map[string]string",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", map[string]string{
					"currentDateTime": time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC).
						Format(time.RFC3339),
				}).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics as map[string]interface and date as time.Time",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", map[string]interface{}{
					"currentDateTime": time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC),
				}).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics as map[string]interface and date as *time.Time",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", map[string]interface{}{
					"currentDateTime": testconvert.Time(
						time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC),
					),
				}).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics as map[string]interface",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", map[string]interface{}{
					"currentDateTime": time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC).
						Format(time.RFC3339),
				}).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics nil",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", nil).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: nil,
			},
		},
		{
			name: "no context goff specifics",
			ctx:  ffcontext.NewEvaluationContextBuilder("my-targetingKey").Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: nil,
			},
		},
		{
			name: "context goff specifics as GoffContextSpecifics type",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", ffcontext.GoffContextSpecifics{
					CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
				}).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics as GoffContextSpecifics type contains flagList",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", ffcontext.GoffContextSpecifics{
					CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
					FlagList:        []string{"flag1", "flag2"},
				}).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
				FlagList:        []string{"flag1", "flag2"},
			},
		},
		{
			name: "context goff specifics as map[string]interface type contains flagList",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", map[string]interface{}{
					"currentDateTime": testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)).
						Format(time.RFC3339),
					"flagList": []string{"flag1", "flag2"},
				}).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
				FlagList:        []string{"flag1", "flag2"},
			},
		},
		{
			name: "context goff specifics only flagList",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", map[string]interface{}{
					"flagList": []string{"flag1", "flag2"},
				}).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				FlagList: []string{"flag1", "flag2"},
			},
		},
		{
			name: "context goff specifics with exporter metadata",
			ctx: ffcontext.NewEvaluationContextBuilder("my-targetingKey").
				AddCustom("gofeatureflag", map[string]interface{}{
					"exporterMetadata": map[string]interface{}{
						"toto": 123,
						"titi": 123.45,
						"tutu": true,
						"tata": "bonjour",
					},
				}).
				Build(),
			want: ffcontext.GoffContextSpecifics{
				ExporterMetadata: map[string]interface{}{
					"toto": 123,
					"titi": 123.45,
					"tutu": true,
					"tata": "bonjour",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.ExtractGOFFProtectedFields()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEvaluationContext_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		context  ffcontext.EvaluationContext
		expected string
	}{
		{
			name:     "marshal with empty attributes",
			context:  ffcontext.NewEvaluationContext("test-key"),
			expected: `{"targetingKey":"test-key","attributes":{}}`,
		},
		{
			name: "marshal with attributes",
			context: ffcontext.NewEvaluationContextBuilder("test-key").
				AddCustom("attr1", "value1").
				AddCustom("attr2", 123).
				Build(),
			expected: `{"targetingKey":"test-key","attributes":{"attr1":"value1","attr2":123}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.context.MarshalJSON()
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestEvaluationContext_ToMap(t *testing.T) {
	tests := []struct {
		name     string
		context  ffcontext.EvaluationContext
		expected map[string]interface{}
	}{
		{
			name:     "empty attributes",
			context:  ffcontext.NewEvaluationContext("test-key"),
			expected: map[string]interface{}{"targetingKey": "test-key"},
		},
		{
			name: "attributes with values",
			context: ffcontext.NewEvaluationContextBuilder("test-key").
				AddCustom("attr1", "value1").
				AddCustom("attr2", 123).
				Build(),
			expected: map[string]interface{}{
				"targetingKey": "test-key",
				"attr1":        "value1",
				"attr2":        123,
			},
		},
		{
			name: "attributes with nested map",
			context: ffcontext.NewEvaluationContextBuilder("test-key").
				AddCustom("nested", map[string]interface{}{
					"key1": "value1",
					"key2": 42,
				}).
				Build(),
			expected: map[string]interface{}{
				"targetingKey": "test-key",
				"nested": map[string]interface{}{
					"key1": "value1",
					"key2": 42,
				},
			},
		},
		{
			name: "attributes with nil value",
			context: ffcontext.NewEvaluationContextBuilder("test-key").
				AddCustom("attr1", nil).
				Build(),
			expected: map[string]interface{}{
				"targetingKey": "test-key",
				"attr1":        nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.context.ToMap()
			assert.Equal(t, tt.expected, got)
		})
	}
}
