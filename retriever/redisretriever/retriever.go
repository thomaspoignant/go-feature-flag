package redisretriever

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	redis "github.com/redis/go-redis/v9"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// Retriever is the interface to fetch the flags from Redis.
type Retriever struct {
	// Options to connect to Redis
	Options *redis.Options

	// Prefix is the prefix of the keys in Redis, it is used to filter
	// the keys to retrieve in redis. If empty, no prefix is used.
	// Your flag names will be returned without the prefix.
	Prefix string

	status retriever.Status
	client *redis.Client
}

// Init is initializing the retriever to start fetching the flags configuration.
func (r *Retriever) Init(ctx context.Context, _ *fflog.FFLogger) error {
	r.status = retriever.RetrieverNotReady
	client := redis.NewClient(r.Options)

	_, err := client.Ping(ctx).Result()
	if err != nil {
		r.status = retriever.RetrieverError
		return err
	}
	r.client = client
	r.status = retriever.RetrieverReady
	return nil
}

// Status is the function returning the internal state of the retriever.
func (r *Retriever) Status() retriever.Status {
	if r == nil || r.status == "" {
		return retriever.RetrieverNotReady
	}
	return r.status
}

// Shutdown gracefully shutdown the provider and set the status as not ready.
func (r *Retriever) Shutdown(_ context.Context) error {
	if r == nil {
		return nil
	}
	if r.client == nil {
		r.status = retriever.RetrieverNotReady
		return nil
	}
	err := r.client.Close()
	r.client = nil

	if err != nil && !errors.Is(err, redis.ErrClosed) {
		r.status = retriever.RetrieverError
		return err
	}
	r.status = retriever.RetrieverNotReady
	return nil
}

// Retrieve is the function in charge of fetching the flag configuration.
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	var flagsData = make(map[string]interface{})
	iter := r.client.Scan(ctx, 0, r.Prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		keyWithoutPrefix := strings.Replace(key, r.Prefix, "", 1)
		value, err := r.client.Get(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("error retrieving flag '%s': %v", key, err)
		}

		var flagData interface{}
		if err := json.Unmarshal([]byte(value), &flagData); err != nil {
			return nil, fmt.Errorf("error unmarshalling flag '%s': %v", key, err)
		}

		flagsData[keyWithoutPrefix] = flagData
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through Redis keys: %v", err)
	}

	content, err := json.Marshal(flagsData)
	if err != nil {
		return nil, fmt.Errorf("error marshalling flags data: %v", err)
	}
	return content, nil
}
