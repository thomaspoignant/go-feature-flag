//go:build docker
// +build docker

package redisretriever_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	testcontainerRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	ret "github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/redisretriever"
	"golang.org/x/net/context"
)

var redisContainerList = make(map[string]*testcontainerRedis.RedisContainer)

func Test_Redis_Retrieve(t *testing.T) {

	tests := []struct {
		name        string
		prefix      string
		flagsToLoad []string

		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:        "flags with prefix",
			flagsToLoad: []string{"flag1.json", "flag2.json"},
			prefix:      "goff:",
			wantErr:     assert.NoError,
		},
		{
			name:        "flags without prefix",
			flagsToLoad: []string{"flag1.json", "flag2.json"},
			prefix:      "",
			wantErr:     assert.NoError,
		},
		{
			name:        "flags with invalid json",
			flagsToLoad: []string{"flag1.json", " flag2.json", "invalid-json.json"},
			prefix:      "",
			wantErr:     assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := startRedisAndAddData(t, t.Name(), tt.flagsToLoad, tt.prefix)
			defer stopRedis(t, t.Name())

			retriever := redisretriever.Retriever{
				Options: options,
				Prefix:  tt.prefix,
			}

			err := retriever.Init(context.Background(), nil)
			assert.NoError(t, err)
			defer func() {
				err := retriever.Shutdown(context.TODO())
				assert.NoError(t, err)
			}()
			got, err := retriever.Retrieve(context.Background())
			tt.wantErr(t, err)

			if err == nil {
				flagStr := "{"
				for index, file := range tt.flagsToLoad {
					flagName, content := readFile(t, file)
					flagStr += `"` + flagName + `":` + string(content)
					if index < len(tt.flagsToLoad)-1 {
						flagStr += ","
					}
				}
				flagStr += "}"
				assert.JSONEq(t, string(got), flagStr)
			}

		})
	}
}

func startRedisAndAddData(t *testing.T, testName string, files []string, prefix string) *redis.Options {
	ctx := context.TODO()

	redisContainer, err := testcontainerRedis.RunContainer(ctx,
		testcontainers.WithImage("docker.io/redis:7"),
		testcontainerRedis.WithSnapshotting(10, 1),
		testcontainerRedis.WithLogLevel(testcontainerRedis.LogLevelVerbose),
		testcontainerRedis.WithConfigFile(filepath.Join("testdata", "redis.conf")),
	)
	assert.NoError(t, err)

	address, err := redisContainer.Endpoint(ctx, "")
	assert.NoError(t, err)

	options := &redis.Options{
		Addr: address,
	}
	rdb := redis.NewClient(options)
	defer func() { _ = rdb.Close() }()

	for _, file := range files {
		flagName, content := readFile(t, file)
		_, errAdd := rdb.Set(ctx, prefix+flagName, string(content), 0).Result()
		assert.NoError(t, errAdd)

	}
	redisContainerList[testName] = redisContainer
	return options
}

func stopRedis(t *testing.T, testName string) {
	redisContainer := redisContainerList[testName]
	if redisContainer != nil {
		if err := redisContainer.Terminate(context.TODO()); err != nil {
			assert.NoError(t, err)
		}
	}
}

func Test_Redis_Status(t *testing.T) {
	t.Run("should return NotReady for nil receiver", func(t *testing.T) {
		var retriever *redisretriever.Retriever
		assert.Equal(t, ret.RetrieverNotReady, retriever.Status())
	})

	t.Run("should return NotReady for uninitialized retriever", func(t *testing.T) {
		retriever := &redisretriever.Retriever{}
		assert.Equal(t, ret.RetrieverNotReady, retriever.Status())
	})
}

func Test_Redis_Shutdown(t *testing.T) {
	options := startRedisAndAddData(t, t.Name(), []string{"flag1.json"}, "")
	defer stopRedis(t, t.Name())

	t.Run("should close connection successfully", func(t *testing.T) {
		retriever := redisretriever.Retriever{
			Options: options,
		}

		err := retriever.Init(context.Background(), nil)
		assert.NoError(t, err)

		err = retriever.Shutdown(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, ret.RetrieverNotReady, retriever.Status())
	})

	t.Run("should succeed when called before init", func(t *testing.T) {
		retriever := &redisretriever.Retriever{
			Options: options,
		}

		err := retriever.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("allow multiple calls idempotently", func(t *testing.T) {
		retriever := redisretriever.Retriever{
			Options: options,
		}

		err := retriever.Init(context.Background(), nil)
		assert.NoError(t, err)

		err = retriever.Shutdown(context.Background())
		assert.NoError(t, err)

		err = retriever.Shutdown(context.Background())
		assert.NoError(t, err)
	})
}

func readFile(t *testing.T, file string) (string, []byte) {
	content, err := os.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		return "", nil
	}
	assert.NoError(t, err)
	flagName := file[:len(file)-5]
	return flagName, content
}
