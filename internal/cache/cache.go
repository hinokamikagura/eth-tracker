package cache

import (
	"context"
	"log/slog"
)

type (
	Cache interface {
		Add(context.Context, string) error
		Remove(context.Context, string) error
		Exists(context.Context, string) (bool, error)
		Size(context.Context) (int64, error)
	}

	CacheOpts struct {
		Logg      *slog.Logger
		RedisDSN  string
		CacheType string
	}
)

func New(o CacheOpts) (Cache, error) {
	var cache Cache

	switch o.CacheType {
	case "map":
		cache = NewMapCache()
	case "redis":
		redisCache, err := NewRedisCache(redisOpts{
			DSN: o.RedisDSN,
		})
		if err != nil {
			return nil, err
		}
		cache = redisCache
	default:
		cache = NewMapCache()
		o.Logg.Warn("invalid cache type, using default type (map)")
	}

	return cache, nil
}
