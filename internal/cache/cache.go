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

	// geSmartContracts, err := o.Chain.Provider().GetGESmartContracts(
	// 	context.Background(),
	// 	o.Registries,
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("cache could not bootstrap GE smart contracts: err %v", err)
	// }

	// for k, v := range geSmartContracts {
	// 	cache.Add(k, v)
	// }
	// for _, address := range o.Watchlist {
	// 	cache.Add(address, false)
	// }
	// for _, address := range o.Blacklist {
	// 	cache.Remove(address)
	// }
	// o.Logg.Info("cache bootstrap complete", "cached_addresses", cache.Size())

	return cache, nil
}
