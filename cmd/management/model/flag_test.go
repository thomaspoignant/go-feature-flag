package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	coredto "github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func strPtr(s string) *string { return &s }

func TestValidateFlagName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "my-flag", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateFlagName(tt.input)
			assert.Equal(t, tt.wantErr, len(errs) > 0)
		})
	}
}

func TestValidateFlagPayload(t *testing.T) {
	twoVars := map[string]*any{
		"on":  any3(true),
		"off": any3(false),
	}
	defaultRule := &flag.Rule{VariationResult: strPtr("on")}

	tests := []struct {
		name    string
		dto     coredto.DTO
		wantErr bool
	}{
		{
			name: "valid minimal",
			dto: coredto.DTO{
				Variations:  &twoVars,
				DefaultRule: defaultRule,
			},
			wantErr: false,
		},
		{
			name: "missing variations",
			dto: coredto.DTO{
				DefaultRule: defaultRule,
			},
			wantErr: true,
		},
		{
			name: "missing defaultRule",
			dto: coredto.DTO{
				Variations: &twoVars,
			},
			wantErr: true,
		},
		{
			name: "rule without query",
			dto: coredto.DTO{
				Variations:  &twoVars,
				DefaultRule: defaultRule,
				Rules: &[]flag.Rule{
					{VariationResult: strPtr("on")},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateFlagPayload(tt.dto)
			assert.Equal(t, tt.wantErr, len(errs) > 0, "errors: %+v", errs)
		})
	}
}

func any3(v any) *any { return &v }
