package folderloader_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/internal/folderloader"
)

func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		itemName  string
		item      string
		prefix    string
		pattern   string
		recursive bool
		wantErr   string
	}{
		{name: "item only", item: "a.yaml"},
		{name: "nothing is the bucket root"},
		{name: "prefix only", prefix: "flags/"},
		{name: "root prefix recursive", prefix: "/", recursive: true},
		{name: "prefix with pattern", prefix: "flags/", pattern: "*.yaml", recursive: true},
		{name: "item and prefix", item: "a.yaml", prefix: "f/", wantErr: "item and prefix are mutually exclusive"},
		{
			name: "custom item name", itemName: "object", item: "a.yaml", prefix: "f/",
			wantErr: "object and prefix are mutually exclusive",
		},
		{
			name: "custom item name with pattern", itemName: "object", item: "a.yaml", pattern: "*",
			wantErr: "pattern and recursive can only be used without object",
		},
		{
			name: "item with pattern", item: "a.yaml", pattern: "*.yaml",
			wantErr: "pattern and recursive can only be used without item",
		},
		{
			name: "item with recursive", item: "a.yaml", recursive: true,
			wantErr: "pattern and recursive can only be used without item",
		},
		{name: "invalid pattern", prefix: "f/", pattern: "[a-", wantErr: `invalid pattern "[a-"`},
		{name: "pattern with a slash", prefix: "f/", pattern: "sub/*.yaml", wantErr: `can't contain "/"`},
		{name: "pattern with a leading slash", pattern: "/a.yaml", wantErr: `can't contain "/"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.itemName == "" {
				tt.itemName = "item"
			}
			err := folderloader.Validate(tt.itemName, tt.item, tt.prefix, tt.pattern, tt.recursive)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
