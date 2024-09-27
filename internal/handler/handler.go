package handler

import "github.com/grassrootseconomics/eth-tracker/internal/cache"

type HandlerContainer struct {
	cache cache.Cache
}

func New(cacheProvider cache.Cache) *HandlerContainer {
	return &HandlerContainer{
		cache: cacheProvider,
	}
}
