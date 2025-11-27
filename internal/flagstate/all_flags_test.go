package flagstate_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flagstate"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func TestAllFlags(t *testing.T) {
	afs := flagstate.NewAllFlags()
	assert.NotNil(t, afs.GetFlags())
	assert.Equal(t, 0, len(afs.GetFlags()))
	assert.True(t, afs.IsValid())

	fs := flagstate.FlagState{
		Value:         20,
		Timestamp:     time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC).Unix(),
		VariationType: "var_a",
		TrackEvents:   false,
		Failed:        false,
		Reason:        flag.ReasonStatic,
	}
	afs.AddFlag("my-key", fs)
	assert.Equal(t, 1, len(afs.GetFlags()))
	assert.True(t, afs.IsValid())
	fs2 := flagstate.FlagState{
		Value:         20,
		Timestamp:     time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC).Unix(),
		VariationType: "var_a",
		TrackEvents:   false,
		Failed:        true,
		ErrorCode:     flag.ErrorCodeTargetingKeyMissing,
		ErrorDetails:  "The targeting key is missing",
		Reason:        flag.ReasonError,
	}
	afs.AddFlag("my-key-2", fs2)
	assert.Equal(t, 2, len(afs.GetFlags()))
	assert.False(t, afs.IsValid())

	want := "{\"flags\":{\"my-key\":{\"value\":20,\"timestamp\":1659312010,\"variationType\":\"var_a\",\"trackEvents\":false,\"errorCode\":\"\",\"reason\":\"STATIC\"},\"my-key-2\":{\"value\":20,\"timestamp\":1659312010,\"variationType\":\"var_a\",\"trackEvents\":false,\"errorCode\":\"TARGETING_KEY_MISSING\",\"errorDetails\":\"The targeting key is missing\",\"reason\":\"ERROR\"}},\"valid\":false}\n"
	got, err := afs.MarshalJSON()
	assert.NoError(t, err)
	assert.JSONEq(t, want, string(got))
}
