package cache

import (
	"context"

	"github.com/puzpuzpuz/xsync/v3"
)

type mapCache struct {
	xmap *xsync.Map
}

func NewMapCache() Cache {
	return &mapCache{
		xmap: xsync.NewMap(),
	}
}

func (c *mapCache) Add(_ context.Context, key string) error {
	c.xmap.Store(key, true)
	return nil
}

func (c *mapCache) Remove(_ context.Context, key string) error {
	c.xmap.Delete(key)
	return nil
}

func (c *mapCache) Exists(_ context.Context, key string) (bool, error) {
	_, ok := c.xmap.Load(key)
	return ok, nil
}

func (c *mapCache) Size(_ context.Context) (int64, error) {
	return int64(c.xmap.Size()), nil
}
