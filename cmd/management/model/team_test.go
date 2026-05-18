package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{"admin", RoleAdmin, true},
		{"editor", RoleEditor, true},
		{"viewer", RoleViewer, true},
		{"empty", Role(""), false},
		{"unknown", Role("owner"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.role.IsValid())
		})
	}
}

func TestRole_AtLeast(t *testing.T) {
	tests := []struct {
		name string
		have Role
		min  Role
		want bool
	}{
		{"admin >= viewer", RoleAdmin, RoleViewer, true},
		{"admin >= editor", RoleAdmin, RoleEditor, true},
		{"admin >= admin", RoleAdmin, RoleAdmin, true},
		{"editor >= viewer", RoleEditor, RoleViewer, true},
		{"editor >= editor", RoleEditor, RoleEditor, true},
		{"editor < admin", RoleEditor, RoleAdmin, false},
		{"viewer >= viewer", RoleViewer, RoleViewer, true},
		{"viewer < editor", RoleViewer, RoleEditor, false},
		{"viewer < admin", RoleViewer, RoleAdmin, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.have.AtLeast(tt.min))
		})
	}
}
