// Package cache provides address filtering functionality for the eth-tracker service.
// It maintains a cache of addresses to track, populated from registries, watchlists,
// and blacklists during bootstrap.
package cache

import (
	"context"
	"log/slog"

	"github.com/grassrootseconomics/eth-tracker/internal/chain"
)

const (
	// CacheTypeInternal is the internal in-memory cache implementation
	CacheTypeInternal = "internal"
	// CacheTypeRedis is the Redis-backed cache implementation (not yet implemented)
	CacheTypeRedis = "redis"
)

type (
	// Cache defines the interface for address caching and filtering.
	Cache interface {
		// Add adds an address to the cache for tracking.
		Add(context.Context, string) error

		// Remove removes an address from the cache.
		Remove(context.Context, string) error

		// Exists checks if an address exists in the cache.
		Exists(context.Context, string) (bool, error)

		// ExistsNetwork checks if a token address exists and any of the provided addresses exist.
		// Used for network-based filtering (e.g., token transfers within a network).
		ExistsNetwork(context.Context, string, ...string) (bool, error)

		// Size returns the current number of addresses in the cache.
		Size(context.Context) (int64, error)
	}

	// CacheOpts contains configuration options for creating a new Cache instance.
	CacheOpts struct {
		RedisDSN   string       // Redis DSN (if using Redis cache type)
		CacheType  string       // Cache implementation type ("internal" or "redis")
		Registries []string     // Registry contract addresses to bootstrap from
		Watchlist  []string     // Additional addresses to track
		Blacklist  []string     // Addresses to exclude from tracking
		Chain      chain.Chain  // Chain client for fetching registry data
		Logg       *slog.Logger // Structured logger
	}
)

// New creates a new Cache instance based on the provided options.
// It bootstraps the cache with addresses from registries, watchlists, and blacklists.
func New(o CacheOpts) (Cache, error) {
	o.Logg.Info("initializing cache",
		"cache_type", o.CacheType,
		"registry_count", len(o.Registries),
		"watchlist_count", len(o.Watchlist),
		"blacklist_count", len(o.Blacklist),
	)

	var cache Cache

	switch o.CacheType {
	case CacheTypeInternal:
		cache = NewMapCache()
	case CacheTypeRedis:
		// TODO: Implement Redis cache
		o.Logg.Warn("Redis cache not yet implemented, falling back to internal cache")
		cache = NewMapCache()
	default:
		o.Logg.Warn("unknown cache type, using default internal cache", "cache_type", o.CacheType)
		cache = NewMapCache()
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
