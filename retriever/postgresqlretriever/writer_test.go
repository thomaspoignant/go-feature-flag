//go:build docker

package postgresqlretriever_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever"
)

const validFlagDefinition = `{"variations":{"enabled":true,"disabled":false},"defaultRule":{"variation":"disabled"}}`
const updatedFlagDefinition = `{"variations":{"enabled":true,"disabled":false},"defaultRule":{"variation":"enabled"}}`

func newWriterRetriever(t *testing.T, connectionString string, flagset string) *postgresqlretriever.Retriever {
	r := &postgresqlretriever.Retriever{
		URI:   connectionString,
		Table: "go_feature_flag",
	}
	assert.NoError(t, r.Init(context.Background(), nil, &flagset))
	return r
}

func TestWriter_FlagsetNotConfigured(t *testing.T) {
	connectionString := startPostgreSQLAndAddData(t, t.Name(), []string{"sql/init.sql"})
	defer stopPostgreSQL(t, t.Name())

	r := newWriterRetriever(t, connectionString, "")
	defer func() { assert.NoError(t, r.Shutdown(context.Background())) }()

	_, err := r.CreateFlag(context.Background(), "my-flag", []byte(validFlagDefinition))
	assert.ErrorIs(t, err, retriever.ErrFlagsetNotConfigured)

	_, _, err = r.UpsertFlag(context.Background(), "my-flag", []byte(validFlagDefinition), nil)
	assert.ErrorIs(t, err, retriever.ErrFlagsetNotConfigured)

	err = r.DeleteFlag(context.Background(), "my-flag", nil)
	assert.ErrorIs(t, err, retriever.ErrFlagsetNotConfigured)
}

func TestWriter_CreateGetDelete(t *testing.T) {
	connectionString := startPostgreSQLAndAddData(t, t.Name(), []string{"sql/init.sql"})
	defer stopPostgreSQL(t, t.Name())

	r := newWriterRetriever(t, connectionString, "team-A")
	defer func() { assert.NoError(t, r.Shutdown(context.Background())) }()
	ctx := context.Background()

	assert.Equal(t, "postgresql:go_feature_flag", r.Source())

	_, _, err := r.GetFlag(ctx, "my-flag")
	assert.ErrorIs(t, err, retriever.ErrFlagNotFound)

	etag, err := r.CreateFlag(ctx, "my-flag", []byte(validFlagDefinition))
	assert.NoError(t, err)
	assert.NotEmpty(t, etag)

	_, err = r.CreateFlag(ctx, "my-flag", []byte(validFlagDefinition))
	assert.ErrorIs(t, err, retriever.ErrFlagAlreadyExists)

	definition, getETag, err := r.GetFlag(ctx, "my-flag")
	assert.NoError(t, err)
	assert.JSONEq(t, validFlagDefinition, string(definition))
	assert.Equal(t, etag, getETag)

	assert.NoError(t, r.DeleteFlag(ctx, "my-flag", &etag))
	_, _, err = r.GetFlag(ctx, "my-flag")
	assert.ErrorIs(t, err, retriever.ErrFlagNotFound)

	assert.ErrorIs(t, r.DeleteFlag(ctx, "my-flag", nil), retriever.ErrFlagNotFound)
}

func TestWriter_UpsertCreatesAndReplaces(t *testing.T) {
	connectionString := startPostgreSQLAndAddData(t, t.Name(), []string{"sql/init.sql"})
	defer stopPostgreSQL(t, t.Name())

	r := newWriterRetriever(t, connectionString, "team-A")
	defer func() { assert.NoError(t, r.Shutdown(context.Background())) }()
	ctx := context.Background()

	created, etag, err := r.UpsertFlag(ctx, "my-flag", []byte(validFlagDefinition), nil)
	assert.NoError(t, err)
	assert.True(t, created)
	assert.NotEmpty(t, etag)

	created, etag2, err := r.UpsertFlag(ctx, "my-flag", []byte(updatedFlagDefinition), &etag)
	assert.NoError(t, err)
	assert.False(t, created)
	assert.NotEqual(t, etag, etag2)

	definition, _, err := r.GetFlag(ctx, "my-flag")
	assert.NoError(t, err)
	assert.JSONEq(t, updatedFlagDefinition, string(definition))
}

func TestWriter_ETagMismatch(t *testing.T) {
	connectionString := startPostgreSQLAndAddData(t, t.Name(), []string{"sql/init.sql"})
	defer stopPostgreSQL(t, t.Name())

	r := newWriterRetriever(t, connectionString, "team-A")
	defer func() { assert.NoError(t, r.Shutdown(context.Background())) }()
	ctx := context.Background()

	_, etag, err := r.UpsertFlag(ctx, "my-flag", []byte(validFlagDefinition), nil)
	assert.NoError(t, err)

	staleETag := `"stale-etag-value"`
	_, _, err = r.UpsertFlag(ctx, "my-flag", []byte(updatedFlagDefinition), &staleETag)
	assert.ErrorIs(t, err, retriever.ErrETagMismatch)

	err = r.DeleteFlag(ctx, "my-flag", &staleETag)
	assert.ErrorIs(t, err, retriever.ErrETagMismatch)

	// the correct etag still works after the failed attempts above
	_, _, err = r.UpsertFlag(ctx, "my-flag", []byte(updatedFlagDefinition), &etag)
	assert.NoError(t, err)
}

// TestWriter_ConcurrentUpdate proves the row-level locking (SELECT ... FOR UPDATE) makes the
// optimistic-concurrency check race-free: two goroutines racing an update with the same stale
// ETag must result in exactly one success and one ErrETagMismatch.
func TestWriter_ConcurrentUpdate(t *testing.T) {
	connectionString := startPostgreSQLAndAddData(t, t.Name(), []string{"sql/init.sql"})
	defer stopPostgreSQL(t, t.Name())

	r := newWriterRetriever(t, connectionString, "team-A")
	defer func() { assert.NoError(t, r.Shutdown(context.Background())) }()
	ctx := context.Background()

	_, etag, err := r.UpsertFlag(ctx, "my-flag", []byte(validFlagDefinition), nil)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _, errs[0] = r.UpsertFlag(ctx, "my-flag", []byte(updatedFlagDefinition), &etag)
	}()
	go func() {
		defer wg.Done()
		_, _, errs[1] = r.UpsertFlag(ctx, "my-flag", []byte(updatedFlagDefinition), &etag)
	}()
	wg.Wait()

	successCount := 0
	mismatchCount := 0
	for _, e := range errs {
		switch {
		case e == nil:
			successCount++
		case e == retriever.ErrETagMismatch:
			mismatchCount++
		}
	}
	assert.Equal(t, 1, successCount, "exactly one concurrent update should succeed")
	assert.Equal(t, 1, mismatchCount, "exactly one concurrent update should see a stale etag")
}
