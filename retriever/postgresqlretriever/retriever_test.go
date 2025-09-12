//go:build docker
// +build docker

package postgresqlretriever_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	testcontainerPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever"
)

var postgreSQLContainerList = make(map[string]*testcontainerPostgres.PostgresContainer)

func TestPostgreSQLRetriever(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		table_name string
		columns    map[string]string
		flagset    string
		assertErr  assert.ErrorAssertionFunc
		want       string
	}{
		{
			name: "valid set of flag",
			files: []string{
				"sql/init.sql",
				"sql/insert_data.sql",
			},
			table_name: "go_feature_flag",
			assertErr:  assert.NoError,
			want:       "response/valid.json",
		},
		{
			name: "invalid table name",
			files: []string{
				"sql/init.sql",
				"sql/insert_data.sql",
			},
			table_name: "invalid_table_name",
			assertErr:  assert.Error,
		},
		{
			name: "invalid column names",
			files: []string{
				"sql/init.sql",
				"sql/insert_data.sql",
			},
			assertErr: assert.Error,
			columns: map[string]string{
				"flag_name": "invalid_flag_name",
				"flagset":   "invalid_flagset",
				"config":    "invalid_config",
			},
		},
		{
			name: "valid set of flag with flagset",
			files: []string{
				"sql/init.sql",
				"sql/insert_data.sql",
				"sql/insert_alternative_flagset.sql",
			},
			table_name: "go_feature_flag",
			assertErr:  assert.NoError,
			flagset:    "team-A",
			want:       "response/valid_alternative_flagset.json",
		},
		{
			name: "valid set of flag with empty flagset",
			files: []string{
				"sql/init.sql",
				"sql/insert_data.sql",
			},
			table_name: "go_feature_flag",
			assertErr:  assert.NoError,
			flagset:    "empty-flagset",
			want:       "response/empty-flagset.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connectionString := startPostgreSQLAndAddData(t, t.Name(), tt.files)
			defer stopPostgreSQL(t, t.Name())

			retriever := postgresqlretriever.Retriever{
				URI:     connectionString,
				Table:   tt.table_name,
				Columns: tt.columns,
			}
			assert.NoError(t, retriever.Init(context.TODO(), nil, &tt.flagset))
			defer func() {
				assert.NoError(t, retriever.Shutdown(context.TODO()))
			}()

			got, err := retriever.Retrieve(context.TODO())
			tt.assertErr(t, err)
			if err != nil {
				return
			}

			want, err := os.ReadFile(filepath.Join("testdata", tt.want))
			assert.NoError(t, err)
			assert.JSONEq(t, string(want), string(got))

		})
	}
}

func startPostgreSQLAndAddData(t *testing.T, testName string, files []string) string {
	ctx := context.TODO()
	postgreSQLContainer, err := testcontainerPostgres.Run(ctx, "postgres:15-alpine")
	assert.NoError(t, err)

	connectionString, err := postgreSQLContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)
	postgreSQLContainerList[testName] = postgreSQLContainer

	// Retry connection with backoff
	var conn *pgx.Conn
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		conn, err = pgx.Connect(ctx, connectionString)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	if err != nil {
		t.Fatalf("Failed to connect to database after %d retries: %v", maxRetries, err)
	}
	defer func() {
		if conn != nil {
			conn.Close(ctx)
		}
	}()

	for _, file := range files {
		content, err := os.ReadFile(filepath.Join("testdata", file))
		assert.NoError(t, err)

		// Execute the SQL content
		_, err = conn.Exec(ctx, string(content))
		assert.NoError(t, err)
	}

	return connectionString
}

func stopPostgreSQL(t *testing.T, testName string) {
	postgreSQLContainer := postgreSQLContainerList[testName]
	if postgreSQLContainer != nil {
		if err := postgreSQLContainer.Terminate(context.TODO()); err != nil {
			assert.NoError(t, err)
		}
	}
}
