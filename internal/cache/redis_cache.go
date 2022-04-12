package cache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

var _ Cacher = (*RedisCache)(nil)

func NewRedisCacheClient() *RedisCache {
	var (
		cache *redis.Client
	)
	var redisDSN string
	if os.Getenv("PLATFORM") == "docker" {
		redisDSN = fmt.Sprintf(
			"%v:%v",
			os.Getenv("REDIS_CONTAINER_NAME"),
			os.Getenv("REDIS_PORT"),
		)
	} else {
		redisDSN = fmt.Sprintf(
			"%v:%v",
			os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PORT"),
		)
	}
	cache = redis.NewClient(&redis.Options{
		Addr: redisDSN,
	})
	_, err := cache.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return &RedisCache{
		client: cache,
	}
}

// implement the Cacher interface
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return result, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value string, duration time.Duration) error {
	return c.client.Set(ctx, key, value, duration).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
