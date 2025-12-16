//go:build docker

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
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever"
)

var postgreSQLContainerList = make(map[string]*testcontainerPostgres.PostgresContainer)

func TestPostgreSQLRetriever(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		tableName string
		columns   map[string]string
		flagset   string
		assertErr assert.ErrorAssertionFunc
		want      string
	}{
		{
			name: "valid set of flag",
			files: []string{
				"sql/init.sql",
				"sql/insert_data.sql",
			},
			tableName: "go_feature_flag",
			assertErr: assert.NoError,
			want:      "response/valid.json",
		},
		{
			name: "invalid table name",
			files: []string{
				"sql/init.sql",
				"sql/insert_data.sql",
			},
			tableName: "invalid_table_name",
			assertErr: assert.Error,
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
			tableName: "go_feature_flag",
			assertErr: assert.NoError,
			flagset:   "team-A",
			want:      "response/valid_alternative_flagset.json",
		},
		{
			name: "valid set of flag with empty flagset",
			files: []string{
				"sql/init.sql",
				"sql/insert_data.sql",
			},
			tableName: "go_feature_flag",
			assertErr: assert.NoError,
			flagset:   "empty-flagset",
			want:      "response/empty-flagset.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connectionString := startPostgreSQLAndAddData(t, t.Name(), tt.files)
			defer stopPostgreSQL(t, t.Name())

			r := postgresqlretriever.Retriever{
				URI:     connectionString,
				Table:   tt.tableName,
				Columns: tt.columns,
			}
			assert.NoError(t, r.Init(context.TODO(), nil, &tt.flagset))
			defer func() {
				assert.NoError(t, r.Shutdown(context.TODO()))
			}()

			assert.Equal(t, r.Status(), retriever.RetrieverReady)
			got, err := r.Retrieve(context.TODO())
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

// TestRetrieverErrorHandling tests various error conditions in the PostgreSQL retriever
func TestRetrieverErrorHandling(t *testing.T) {
	t.Run("Init - Invalid connection URI", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   "invalid-connection-string",
			Table: "test_table",
		}

		err := r.Init(context.Background(), nil, nil)
		assert.Error(t, err)
		assert.Equal(t, retriever.RetrieverError, r.Status())
	})

	t.Run("Init - Connection timeout", func(t *testing.T) {
		// Use a non-routable IP address to simulate connection timeout
		r := postgresqlretriever.Retriever{
			URI:   "postgres://user:pass@192.0.2.1:5432/db", // RFC5737 test address
			Table: "test_table",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := r.Init(ctx, nil, nil)
		assert.Error(t, err)
		assert.Equal(t, retriever.RetrieverError, r.Status())
	})

	t.Run("Init - Ping failure", func(t *testing.T) {
		// This test uses a malformed URI that allows connection but fails on ping
		r := postgresqlretriever.Retriever{
			URI:   "postgres://nonexistentuser:wrongpass@localhost:9999/nonexistentdb",
			Table: "test_table",
		}

		err := r.Init(context.Background(), nil, nil)
		assert.Error(t, err)
		assert.Equal(t, retriever.RetrieverError, r.Status())
	})

	t.Run("Retrieve - nil connection", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   "postgres://user:pass@localhost:5432/db",
			Table: "test_table",
		}
		// Don't call Init() to ensure conn is nil

		_, err := r.Retrieve(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection is not initialized")
	})

	t.Run("Status - nil receiver", func(t *testing.T) {
		var r *postgresqlretriever.Retriever
		status := r.Status()
		assert.Equal(t, retriever.RetrieverNotReady, status)
	})

	t.Run("Status - empty status", func(t *testing.T) {
		r := &postgresqlretriever.Retriever{}
		status := r.Status()
		assert.Equal(t, retriever.RetrieverNotReady, status)
	})

	t.Run("Shutdown - nil connection", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   "postgres://user:pass@localhost:5432/db",
			Table: "test_table",
		}
		// Don't call Init() to ensure conn is nil

		err := r.Shutdown(context.Background())
		assert.NoError(t, err) // Should not error when connection is nil
	})
}

// TestRetrieverQueryErrors tests query-related error conditions using a real PostgreSQL container
func TestRetrieverQueryErrors(t *testing.T) {
	connectionString := startPostgreSQLAndAddData(t, t.Name(), []string{"sql/init.sql", "sql/insert_data.sql"})
	defer stopPostgreSQL(t, t.Name())

	t.Run("Query execution error - invalid table", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "nonexistent_table",
		}

		err := r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Shutdown(context.Background()))
		}()

		_, err = r.Retrieve(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute query")
	})

	t.Run("Query execution error - invalid columns", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "go_feature_flag",
			Columns: map[string]string{
				"flag_name": "nonexistent_column",
				"config":    "config",
			},
		}

		err := r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Shutdown(context.Background()))
		}()

		_, err = r.Retrieve(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute query")
	})

	t.Run("Context cancellation during query", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "go_feature_flag",
		}

		err := r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Shutdown(context.Background()))
		}()

		// Create a context that is immediately cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = r.Retrieve(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute query")
	})
}

// TestRetrieverDataHandlingErrors tests data processing error conditions
func TestRetrieverDataHandlingErrors(t *testing.T) {
	// Create a test container with some test data
	connectionString := startPostgreSQLAndAddData(t, t.Name(), []string{"sql/init.sql"})
	defer stopPostgreSQL(t, t.Name())

	// Insert some test data
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connectionString)
	assert.NoError(t, err)
	defer conn.Close(ctx)

	// Insert a row with malformed JSON that will cause unmarshaling issues
	// Using a JSON string that contains control characters that may cause issues
	_, err = conn.Exec(ctx, `
		INSERT INTO go_feature_flag (flag_name, flagset, config) 
		VALUES ('problematic_flag', 'default', '{"key": "value\nwith\nnewlines"}')
	`)
	assert.NoError(t, err)

	// Insert a row with valid JSON for comparison
	_, err = conn.Exec(ctx, `
		INSERT INTO go_feature_flag (flag_name, flagset, config) 
		VALUES ('valid_flag', 'default', '{"valid": "json"}')
	`)
	assert.NoError(t, err)

	t.Run("Data processing with special characters", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "go_feature_flag",
		}

		err := r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Shutdown(context.Background()))
		}()

		// This should handle the data gracefully
		result, err := r.Retrieve(context.Background())
		assert.NoError(t, err)

		// Should contain both flags since the JSON is valid
		assert.Contains(t, string(result), "valid_flag")
		assert.Contains(t, string(result), "problematic_flag")
	})

	t.Run("Context timeout during query execution", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "go_feature_flag",
		}

		err := r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Shutdown(context.Background()))
		}()

		// Create a context with a very short timeout to cause timeout errors
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
		defer cancel()

		// This should cause a timeout error during query execution
		_, err = r.Retrieve(ctx)
		// We expect either success (if query completes quickly) or timeout error
		if err != nil {
			assert.Contains(t, err.Error(), "failed to execute query")
		}
	})
}

// TestRetrieverEdgeCases tests additional edge cases and boundary conditions
func TestRetrieverEdgeCases(t *testing.T) {
	connectionString := startPostgreSQLAndAddData(t, t.Name(), []string{"sql/init.sql"})
	defer stopPostgreSQL(t, t.Name())

	t.Run("Empty result set", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "go_feature_flag", // Empty table
		}

		err := r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Shutdown(context.Background()))
		}()

		result, err := r.Retrieve(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "{}", string(result))
	})

	t.Run("Flagset filtering with no matches", func(t *testing.T) {
		// Insert some data first
		ctx := context.Background()
		conn, err := pgx.Connect(ctx, connectionString)
		assert.NoError(t, err)
		defer conn.Close(ctx)

		_, err = conn.Exec(ctx, `
			INSERT INTO go_feature_flag (flag_name, flagset, config) 
			VALUES ('test_flag', 'production', '{"enabled": true}')
		`)
		assert.NoError(t, err)

		// Query with a flagset that doesn't exist
		nonExistentFlagset := "non-existent-flagset"
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "go_feature_flag",
		}

		err = r.Init(context.Background(), nil, &nonExistentFlagset)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Shutdown(context.Background()))
		}()

		result, err := r.Retrieve(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "{}", string(result))
	})

	t.Run("Multiple Init calls", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "go_feature_flag",
		}

		// First init should succeed
		err := r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, retriever.RetrieverReady, r.Status())

		// Second init should also succeed (should reuse connection)
		err = r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, retriever.RetrieverReady, r.Status())

		defer func() {
			assert.NoError(t, r.Shutdown(context.Background()))
		}()
	})

	t.Run("Custom column names", func(t *testing.T) {
		// Create a table with different column names
		ctx := context.Background()
		conn, err := pgx.Connect(ctx, connectionString)
		assert.NoError(t, err)
		defer conn.Close(ctx)

		_, err = conn.Exec(ctx, `
			CREATE TABLE custom_flags (
				name TEXT PRIMARY KEY,
				environment TEXT,
				configuration JSONB
			)
		`)
		assert.NoError(t, err)

		_, err = conn.Exec(ctx, `
			INSERT INTO custom_flags (name, environment, configuration) 
			VALUES ('custom_flag', 'dev', '{"enabled": false}')
		`)
		assert.NoError(t, err)

		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "custom_flags",
			Columns: map[string]string{
				"flag_name": "name",
				"flagset":   "environment",
				"config":    "configuration",
			},
		}

		err = r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Shutdown(context.Background()))
		}()

		result, err := r.Retrieve(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, string(result), "custom_flag")
	})
}

// TestRetrieverShutdownErrors tests shutdown error conditions
func TestRetrieverShutdownErrors(t *testing.T) {
	connectionString := startPostgreSQLAndAddData(t, t.Name(), []string{"sql/init.sql"})
	defer stopPostgreSQL(t, t.Name())

	t.Run("Shutdown with active connection", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "go_feature_flag",
		}

		err := r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)

		// Shutdown should succeed
		err = r.Shutdown(context.Background())
		assert.NoError(t, err)

		// Second shutdown should also succeed (should not error on nil connection)
		err = r.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Shutdown with cancelled context", func(t *testing.T) {
		r := postgresqlretriever.Retriever{
			URI:   connectionString,
			Table: "go_feature_flag",
		}

		err := r.Init(context.Background(), nil, nil)
		assert.NoError(t, err)

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Shutdown with cancelled context should still succeed
		err = r.Shutdown(ctx)
		assert.NoError(t, err)
	})
}
