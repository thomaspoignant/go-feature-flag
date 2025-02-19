package postgresqlretriever_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // Needed for the SQL container driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_PostgreSQLRetriever_Retrieve_JSON(t *testing.T) {
	ctx := context.Background()

	dbName := "flags_db"
	dbUser := "root"
	dbPassword := "example"

	tests := []struct {
		name    string
		want    []byte
		data    string
		wantErr bool
	}{
		{
			name:    "Returns well formed flag definition document",
			data:    testutils.PostgresFindResultString,
			want:    []byte(testutils.PostgresQueryResult),
			wantErr: false,
		},
		{
			name:    "One of the Flag definition document does not have 'flag' key/value (ignore this document)",
			data:    testutils.PostgresMissingFlagKey,
			want:    []byte(testutils.PostgresMissingFlagKeyResult),
			wantErr: false,
		},
		{
			name:    "Flag definition document 'flag' key does not have 'string' value (ignore this document)",
			data:    testutils.PostgresFindResultFlagNoStr,
			want:    []byte(testutils.PostgresFlagKeyNotStringResult),
			wantErr: false,
		},
		{
			name:    "No flags found on DB",
			want:    []byte("{}"),
			wantErr: true,
		},
	}

	// Start the postgres container
	ctr, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Run initialization query to create the table and the column
	_, _, err = ctr.Exec(ctx, []string{"psql", "-U", dbUser, "-d", dbName, "-c", "CREATE TABLE flags (id SERIAL PRIMARY KEY,flag JSONB)"})
	require.NoError(t, err)

	// Create snapshot of the database, which is then restored before each test
	err = ctr.Snapshot(ctx)
	require.NoError(t, err)

	dbURL, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	for _, item := range tests {
		// Restore the database state to its snapshot
		err = ctr.Restore(ctx)
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), dbURL)
		require.NoError(t, err)
		defer conn.Close(context.Background())

		if item.data != "" {
			// Insert data
			var documents []bson.M
			err = json.Unmarshal([]byte(item.data), &documents)
			require.NoError(t, err)

			for _, doc := range documents {
				_, err = conn.Exec(ctx, "INSERT INTO flags(flag) VALUES ($1)", doc)
				require.NoError(t, err)
			}
		}

		// Initialize Retriever
		mdb := postgresqlretriever.Retriever{
			URI:    dbURL,
			Table:  "flags",
			Column: "flag",
			Type:   "json",
		}

		assert.Equal(t, retriever.RetrieverNotReady, mdb.Status())
		err = mdb.Init(context.TODO(), &fflog.FFLogger{})
		assert.NoError(t, err)
		defer func() { _ = mdb.Shutdown(context.TODO()) }()
		assert.Equal(t, retriever.RetrieverReady, mdb.Status())

		got, err := mdb.Retrieve(context.Background())
		if item.want == nil {
			assert.Nil(t, got)
		} else {
			modifiedGot, err := removeIDFromJSON(string(got))
			require.NoError(t, err)
			assert.JSONEq(t, string(item.want), modifiedGot)
		}

		require.NoError(t, err)
	}
}

func Test_PostgreSQLRetriever_Retrieve_Relational(t *testing.T) {
	ctx := context.Background()

	dbName := "flags_db"
	dbUser := "root"
	dbPassword := "example"

	tests := []struct {
		name    string
		want    []byte
		data    string
		wantErr bool
	}{
		{
			name:    "Returns well formed flag definition document",
			data:    testutils.PostgresQueryProperFlagsRelational,
			want:    []byte(testutils.PostgresQueryProperFlagsRelationalResult),
			wantErr: false,
		},
	}

	// Start the postgres container
	ctr, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithInitScripts("create_relational_postgres_db.sql"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	dbURL, _ := ctr.ConnectionString(ctx)

	// Create snapshot of the database, which is then restored before each test
	err = ctr.Snapshot(ctx)
	require.NoError(t, err)

	require.NoError(t, err)

	for _, item := range tests {
		// Restore the database state to its snapshot
		err = ctr.Restore(ctx)
		require.NoError(t, err)

		conn, err := pgx.Connect(ctx, dbURL)
		require.NoError(t, err)
		defer conn.Close(ctx)

		if item.data != "" {
			// Insert data
			_, err = conn.Exec(ctx, item.data)
			require.NoError(t, err)
		}

		// Initialize Retriever
		mdb := postgresqlretriever.Retriever{
			URI:  dbURL,
			Type: "relational",
		}

		assert.Equal(t, retriever.RetrieverNotReady, mdb.Status())
		err = mdb.Init(context.TODO(), &fflog.FFLogger{})
		assert.NoError(t, err)
		defer func() { _ = mdb.Shutdown(context.TODO()) }()
		assert.Equal(t, retriever.RetrieverReady, mdb.Status())

		got, err := mdb.Retrieve(context.Background())
		if item.want == nil {
			assert.Nil(t, got)
		} else {
			modifiedGot, err := removeIDFromJSON(string(got))
			require.NoError(t, err)
			assert.JSONEq(t, string(item.want), modifiedGot)
		}

		require.NoError(t, err)
	}
}

func Test_PostgreSQLRetriever_InvalidURI(t *testing.T) {
	mdb := postgresqlretriever.Retriever{
		URI:    "invalidURI",
		Table:  "xxx",
		Column: "xxx",
	}
	assert.Equal(t, retriever.RetrieverNotReady, mdb.Status())
	err := mdb.Init(context.TODO(), &fflog.FFLogger{})
	assert.Error(t, err)
	assert.Equal(t, retriever.RetrieverError, mdb.Status())
}

func removeIDFromJSON(jsonStr string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	removeIDFields(data)

	modifiedJSON, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(modifiedJSON), nil
}

func removeIDFields(data interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		delete(v, "_id")
		for _, value := range v {
			removeIDFields(value)
		}
	case []interface{}:
		for _, item := range v {
			removeIDFields(item)
		}
	}
}
