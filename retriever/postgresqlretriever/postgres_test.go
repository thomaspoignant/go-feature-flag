package postgresqlretriever_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	rr "github.com/thomaspoignant/go-feature-flag/retriever/postgresqlretriever"
)

func TestGetPool_MultipleURIsAndReuse(t *testing.T) {
	ctx := context.Background()

	// Setup Container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_PASSWORD": "password"},
		// This waits until the log says the system is ready, preventing connection errors
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithStartupTimeout(10*time.Second),
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(10*time.Second),
		),
	}

	pg, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)
	defer pg.Terminate(ctx)

	endpoint, err := pg.Endpoint(ctx, "")
	assert.NoError(t, err)
	baseURI := "postgres://postgres:password@" + endpoint + "/postgres?sslmode=disable"

	// Different URIs
	uri1 := baseURI + "&application_name=A"
	uri2 := baseURI + "&application_name=B"

	// First Calls (RefCount = 1)
	pool1a, err := rr.GetPool(ctx, uri1)
	assert.NoError(t, err)
	assert.NotNil(t, pool1a)

	pool2a, err := rr.GetPool(ctx, uri2)
	assert.NoError(t, err)
	assert.NotNil(t, pool2a)

	// Verify distinct pools
	assert.NotEqual(t, pool1a, pool2a)

	// Reuse Logic (RefCount = 2 for URI1)
	pool1b, err := rr.GetPool(ctx, uri1)
	assert.NoError(t, err)
	assert.Equal(t, pool1a, pool1b, "Should return exact same pool instance")

	// Release Logic
	// URI1 RefCount: 2 -> 1
	rr.ReleasePool(ctx, uri1)

	// URI1 RefCount: 1 -> 0 (Closed & Removed)
	rr.ReleasePool(ctx, uri1)

	// Recreation Logic
	// URI1 should now create a NEW pool
	pool1c, err := rr.GetPool(ctx, uri1)
	assert.NoError(t, err)
	assert.NotEqual(t, pool1a, pool1c, "Should be a new pool instance after full release")

	// Cleanup new pool
	rr.ReleasePool(ctx, uri1)

	// URI2 Cleanup verification
	rr.ReleasePool(ctx, uri2) // RefCount -> 0

	pool2b, err := rr.GetPool(ctx, uri2)
	assert.NoError(t, err)
	assert.NotEqual(t, pool2a, pool2b, "URI2 should be recreated")

	rr.ReleasePool(ctx, uri2)
}
