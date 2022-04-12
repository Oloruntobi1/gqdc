package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNil = errors.New("nil")
)

type Cacher interface {
	Get(context context.Context, key string) (string, error)
	Set(context context.Context, key string, value string, duration time.Duration) error
	Delete(context context.Context, key string) error
}

func GetCurrentCache(cacheSystem string) Cacher {
	switch {
	case cacheSystem == "redis":
		redis := NewRedisCacheClient()
		return redis
	default:
		return nil
	}
}
