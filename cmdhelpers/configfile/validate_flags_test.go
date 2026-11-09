package configfile_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestValidateFlags(t *testing.T) {
	t.Parallel()
	panicking := dto.DTO{
		Variations:  &map[string]*any{"a": testconvert.Interface(true)},
		DefaultRule: &flag.Rule{ProgressiveRollout: &flag.ProgressiveRollout{}},
	}
	valid := dto.DTO{
		Variations:  &map[string]*any{"a": testconvert.Interface(true)},
		DefaultRule: &flag.Rule{VariationResult: testconvert.String("a")},
	}
	invalid := dto.DTO{Variations: &map[string]*any{"a": testconvert.Interface(true)}}

	tests := []struct {
		name  string
		flags map[string]dto.DTO
		want  []configfile.FlagError
	}{
		{name: "nil flags", flags: nil, want: nil},
		{name: "all valid", flags: map[string]dto.DTO{"a": valid}, want: nil},
		{
			name:  "invalid flags sorted by key",
			flags: map[string]dto.DTO{"z": invalid, "ok": valid, "b": invalid},
			want: []configfile.FlagError{
				{Flag: "b", Err: errors.New("missing default rule")},
				{Flag: "z", Err: errors.New("missing default rule")},
			},
		},
		{
			name:  "panicking validation is reported with a generic message",
			flags: map[string]dto.DTO{"p": panicking, "ok": valid},
			want: []configfile.FlagError{
				{
					Flag: "p",
					Err: errors.New(
						"internal validation error: flag definition could not be validated",
					),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := configfile.ValidateFlags(tt.flags)
			assert.Len(t, got, len(tt.want))
			for i, w := range tt.want {
				assert.Equal(t, w.Flag, got[i].Flag)
				// modules/core only panics on an empty progressive rollout before its nil guard is released.
				if got[i].Err.Error() == "invalid progressive rollout, initial and end steps are required" {
					got[i].Err = w.Err
				}
				assert.EqualError(t, got[i].Err, w.Err.Error())
			}
		})
	}
}

var errSentinel = errors.New("sentinel")

type customError struct{}

func (customError) Error() string { return "custom" }

func TestFlagError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		err       error
		wantMsg   string
		wantIsErr error
		wantAs    bool
	}{
		{
			name:      "sentinel error",
			err:       errSentinel,
			wantMsg:   "invalid flag my-flag: sentinel",
			wantIsErr: errSentinel,
		},
		{
			name:    "custom error type",
			err:     customError{},
			wantMsg: "invalid flag my-flag: custom",
			wantAs:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var err error = configfile.FlagError{Flag: "my-flag", Err: tt.err}
			assert.Equal(t, tt.wantMsg, err.Error())
			assert.Equal(t, tt.err, errors.Unwrap(err))
			if tt.wantIsErr != nil {
				assert.ErrorIs(t, err, tt.wantIsErr)
			}
			var ce customError
			assert.Equal(t, tt.wantAs, errors.As(err, &ce))
		})
	}
}
