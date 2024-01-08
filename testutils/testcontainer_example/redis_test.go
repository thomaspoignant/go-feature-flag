package testcontainer_example_test

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"os"
	"testing"
)

func TestWithRedis(t *testing.T) {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	ctx := context.Background()

	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("docker.io/redis:7"),
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
		// redis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	// Your code here ...
}
