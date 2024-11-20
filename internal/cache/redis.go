package cache

import (
	"context"

	"github.com/redis/rueidis"
)

type (
	redisOpts struct {
		DSN string
	}

	redisCache struct {
		client rueidis.Client
	}
)

func NewRedisCache(o redisOpts) (Cache, error) {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{o.DSN},
	})
	if err != nil {
		return nil, err
	}

	return &redisCache{
		client: client,
	}, nil
}

func (c *redisCache) Add(ctx context.Context, key string) error {
	// Without NX it will overwrite any existing KEY
	cmd := c.client.B().Set().Key(key).Value("true").Build()
	return c.client.Do(ctx, cmd).Error()
}

func (c *redisCache) Remove(ctx context.Context, key string) error {
	cmd := c.client.B().Del().Key(key).Build()
	return c.client.Do(ctx, cmd).Error()
}

func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	cmd := c.client.B().Exists().Key(key).Build()
	res, err := c.client.Do(ctx, cmd).AsBool()
	if err != nil {
		return false, err
	}

	return res, nil
}

func (c *redisCache) ExistsNetwork(ctx context.Context, token string, addresses ...string) (bool, error) {
	tokenCmd := c.client.B().Exists().Key(token).Build()
	tokenRes, err := c.client.Do(ctx, tokenCmd).AsBool()
	if err != nil {
		return false, err
	} else if !tokenRes {
		return false, nil
	}

	cmd := c.client.B().Exists().Key(addresses...).Build()
	res, err := c.client.Do(ctx, cmd).AsBool()
	if err != nil {
		return false, err
	}

	return res, nil
}

func (c *redisCache) Size(ctx context.Context) (int64, error) {
	cmd := c.client.B().Dbsize().Build()
	return c.client.Do(ctx, cmd).AsInt64()
}
