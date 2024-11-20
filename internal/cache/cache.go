package cache

import (
	"context"
	"log/slog"

	"github.com/grassrootseconomics/eth-tracker/internal/chain"
)

type (
	Cache interface {
		Add(context.Context, string) error
		Remove(context.Context, string) error
		Exists(context.Context, string) (bool, error)
		ExistsNetwork(context.Context, string, ...string) (bool, error)
		Size(context.Context) (int64, error)
	}

	CacheOpts struct {
		RedisDSN   string
		CacheType  string
		Registries []string
		Watchlist  []string
		Blacklist  []string
		Chain      chain.Chain
		Logg       *slog.Logger
	}
)

func New(o CacheOpts) (Cache, error) {
	o.Logg.Info("initializing cache", "registries", o.Registries, "watchlist", o.Watchlist, "blacklist", o.Blacklist)
	var cache Cache

	switch o.CacheType {
	case "internal":
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

	if err := bootstrapCache(
		o.Chain,
		cache,
		o.Registries,
		o.Watchlist,
		o.Blacklist,
		o.Logg,
	); err != nil {
		return cache, err
	}

	return cache, nil
}
