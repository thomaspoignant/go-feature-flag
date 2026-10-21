//go:build docker

package postgresqlretriever_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testcontainerPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever"
)

var _ retriever.ChangeInfoProvider = (*postgresqlretriever.Retriever)(nil)

func TestRetriever_LastChange(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainerPostgres.Run(ctx, "postgres:15-alpine")
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })
	uri, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	var conn *pgx.Conn
	for i := 0; i < 5; i++ {
		if conn, err = pgx.Connect(ctx, uri); err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	require.NoError(t, err)
	defer func() { _ = conn.Close(ctx) }()
	_, err = conn.Exec(ctx, `
		CREATE TABLE go_feature_flag (
			flag_name TEXT, flagset TEXT, config JSONB, updated_by TEXT, updated_at TIMESTAMPTZ);
		INSERT INTO go_feature_flag VALUES
			('a', 'fs1', '{}', 'jane', '2026-01-01T00:00:00Z'),
			('b', 'fs1', '{}', 'john', '2026-02-01T00:00:00Z'),
			('c', 'other',   '{}', 'zoe',  '2026-03-01T00:00:00Z');`)
	require.NoError(t, err)

	tests := []struct {
		name       string
		columns    map[string]string
		flagset    string
		wantAuthor string
		wantNil    bool
	}{
		{
			// like Retrieve, the default flagset reads every row.
			name:       "default flagset looks at all the rows",
			columns:    map[string]string{"updated_by": "updated_by", "updated_at": "updated_at"},
			flagset:    "default",
			wantAuthor: "zoe",
		},
		{
			name:       "named flagset only looks at its rows",
			columns:    map[string]string{"updated_by": "updated_by", "updated_at": "updated_at"},
			flagset:    "fs1",
			wantAuthor: "john",
		},
		{name: "columns not configured", columns: nil, flagset: "default", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &postgresqlretriever.Retriever{URI: uri, Table: "go_feature_flag", Columns: tt.columns}
			require.NoError(t, r.Init(ctx, nil, &tt.flagset))
			defer func() { _ = r.Shutdown(ctx) }()

			got, err := r.LastChange(ctx)
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.Equal(t, tt.wantAuthor, got.Author)
			assert.False(t, got.ChangedAt.IsZero())
		})
	}
}
