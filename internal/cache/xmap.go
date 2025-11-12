package cache

import (
	"context"

	"github.com/puzpuzpuz/xsync/v3"
)

// mapCache is an in-memory cache implementation using xsync.MapOf for thread-safe operations.
// It provides O(1) average-case lookup, insertion, and deletion operations.
type mapCache struct {
	xmap *xsync.MapOf[string, bool]
}

// NewMapCache creates a new in-memory cache instance.
func NewMapCache() Cache {
	return &mapCache{
		xmap: xsync.NewMapOf[string, bool](),
	}
}

// Add adds an address to the cache.
func (c *mapCache) Add(_ context.Context, key string) error {
	c.xmap.Store(key, true)
	return nil
}

// Remove removes an address from the cache.
func (c *mapCache) Remove(_ context.Context, key string) error {
	c.xmap.Delete(key)
	return nil
}

// Exists checks if an address exists in the cache.
func (c *mapCache) Exists(_ context.Context, key string) (bool, error) {
	_, ok := c.xmap.Load(key)
	return ok, nil
}

// ExistsNetwork checks if a token address exists and any of the provided addresses exist.
// This is used for filtering transactions within a token network (e.g., transfers between
// addresses that are both tracked for a specific token).
func (c *mapCache) ExistsNetwork(_ context.Context, token string, addresses ...string) (bool, error) {
	// First check if the token itself is tracked
	_, tokenExists := c.xmap.Load(token)
	if !tokenExists {
		return false, nil
	}

	// Then check if any of the provided addresses are tracked
	for _, addr := range addresses {
		if _, exists := c.xmap.Load(addr); exists {
			return true, nil
		}
	}

	return false, nil
}

// Size returns the current number of addresses in the cache.
func (c *mapCache) Size(_ context.Context) (int64, error) {
	return int64(c.xmap.Size()), nil
}
