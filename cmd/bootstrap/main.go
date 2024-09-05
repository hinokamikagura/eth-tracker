package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celo-tracker/internal/util"
	"github.com/grassrootseconomics/celoutils/v3"
	"github.com/grassrootseconomics/w3-celo"
	"github.com/grassrootseconomics/w3-celo/module/eth"
	"github.com/knadh/koanf/v2"
)

const cacheType = "redis"

var (
	build = "dev"

	confFlag string

	lo *slog.Logger
	ko *koanf.Koanf
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.Parse()

	lo = util.InitLogger()
	ko = util.InitConfig(lo, confFlag)

	lo.Info("starting GE redis cache bootstrapper", "build", build)
}

func main() {
	var (
		tokenRegistryGetter = w3.MustNewFunc("tokenRegistry()", "address")
		quoterGetter        = w3.MustNewFunc("quoter()", "address")
	)

	chain, err := chain.NewRPCFetcher(chain.CeloRPCOpts{
		RPCEndpoint: ko.MustString("chain.rpc_endpoint"),
		ChainID:     ko.MustInt64("chain.chainid"),
	})
	if err != nil {
		lo.Error("could not initialize chain client", "error", err)
		os.Exit(1)
	}

	cache, err := cache.New(cache.CacheOpts{
		Logg:      lo,
		CacheType: cacheType,
		RedisDSN:  ko.MustString("redis.dsn"),
	})
	if err != nil {
		lo.Error("could not initialize cache", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	for _, registry := range ko.MustStrings("bootstrap.ge_registries") {
		registryMap, err := chain.Provider().RegistryMap(ctx, celoutils.HexToAddress(registry))
		if err != nil {
			lo.Error("could not fetch registry", "error", err)
			os.Exit(1)
		}

		if tokenIndex := registryMap[celoutils.TokenIndex]; tokenIndex != celoutils.ZeroAddress {
			tokenIndexIter, err := chain.Provider().NewBatchIterator(ctx, tokenIndex)
			if err != nil {
				lo.Error("could not create token index iter", "error", err)
				os.Exit(1)
			}

			for {
				batch, err := tokenIndexIter.Next(ctx)
				if err != nil {
					lo.Error("error fetching next token index batch", "error", err)
					os.Exit(1)
				}
				if batch == nil {
					break
				}
				lo.Debug("index batch", "index", tokenIndex.Hex(), "size", len(batch))
				for _, address := range batch {
					if err := cache.Add(ctx, address.Hex()); err != nil {
						lo.Error("redis error setting key", "error", err)
						os.Exit(1)
					}
				}
			}
		}

		if poolIndex := registryMap[celoutils.PoolIndex]; poolIndex != celoutils.ZeroAddress {
			poolIndexIter, err := chain.Provider().NewBatchIterator(ctx, poolIndex)
			if err != nil {
				lo.Error("cache could create pool index iter", "error", err)
				os.Exit(1)
			}

			for {
				batch, err := poolIndexIter.Next(ctx)
				if err != nil {
					lo.Error("error fetching next pool index batch", "error", err)
					os.Exit(1)
				}
				if batch == nil {
					break
				}
				lo.Debug("index batch", "index", poolIndex.Hex(), "size", len(batch))
				for _, address := range batch {
					if err := cache.Add(ctx, address.Hex()); err != nil {
						lo.Error("redis error setting key", "error", err, "address", address.Hex())
						os.Exit(1)
					}

					var poolTokenIndex, priceQuoter common.Address
					err := chain.Provider().Client.CallCtx(
						ctx,
						eth.CallFunc(address, tokenRegistryGetter).Returns(&poolTokenIndex),
						eth.CallFunc(address, quoterGetter).Returns(&priceQuoter),
					)
					if err != nil {
						lo.Error("error fetching pool token index and/or quoter", "error", err)
						os.Exit(1)
					}
					if priceQuoter != celoutils.ZeroAddress {
						if err := cache.Add(ctx, priceQuoter.Hex()); err != nil {
							lo.Error("redis error setting key", "error", err)
							os.Exit(1)
						}
					}
					if poolTokenIndex != celoutils.ZeroAddress {
						if err := cache.Add(ctx, poolTokenIndex.Hex()); err != nil {
							lo.Error("redis error setting key", "error", err)
							os.Exit(1)
						}

						poolTokenIndexIter, err := chain.Provider().NewBatchIterator(ctx, poolTokenIndex)
						if err != nil {
							lo.Error("error creating pool token index iter", "error", err)
							os.Exit(1)
						}

						for {
							batch, err := poolTokenIndexIter.Next(ctx)
							if err != nil {
								lo.Error("error fetching next pool token index batch", "error", err)
								os.Exit(1)
							}
							if batch == nil {
								break
							}
							lo.Debug("index batch", "index", poolTokenIndex.Hex(), "size", len(batch))
							for _, address := range batch {
								if address != celoutils.ZeroAddress {
									if err := cache.Add(ctx, address.Hex()); err != nil {
										lo.Error("redis error setting key", "error", err)
										os.Exit(1)
									}
								}
							}
						}
					}
				}
			}
		}

		if accountsIndex := registryMap[celoutils.AccountIndex]; accountsIndex != celoutils.ZeroAddress {
			accountsIndexIter, err := chain.Provider().NewBatchIterator(ctx, accountsIndex)
			if err != nil {
				lo.Error("could not create accounts index iter", "error", err)
				os.Exit(1)
			}

			for {
				batch, err := accountsIndexIter.Next(ctx)
				if err != nil {
					lo.Error("error fetching next accounts index batch", "error", err)
					os.Exit(1)
				}
				if batch == nil {
					break
				}
				lo.Debug("index batch", "index", accountsIndex.Hex(), "size", len(batch))
				for _, address := range batch {
					if err := cache.Add(ctx, address.Hex()); err != nil {
						lo.Error("redis error setting key", "error", err)
						os.Exit(1)
					}
				}
			}
		}

		for _, v := range registryMap {
			if err := cache.Add(ctx, v.Hex()); err != nil {
				lo.Error("redis error setting key", "error", err)
				os.Exit(1)
			}
		}
	}

	for _, address := range ko.MustStrings("bootstrap.watchlist") {
		if err := cache.Add(ctx, address); err != nil {
			lo.Error("redis error setting key", "error", err)
			os.Exit(1)
		}
	}

	for _, address := range ko.MustStrings("bootstrap.blacklist") {
		if err := cache.Remove(ctx, address); err != nil {
			lo.Error("redis error deleting key", "error", err)
			os.Exit(1)
		}
	}

	if err := cache.Remove(ctx, celoutils.ZeroAddress.Hex()); err != nil {
		lo.Error("redis error deleting key", "error", err)
		os.Exit(1)
	}

	cacheSize, err := cache.Size(ctx)
	if err != nil {
		lo.Warn("error fetching cache size")
	}

	lo.Info("redis cache bootstrap complete", "addresses_cached", cacheSize)
}
