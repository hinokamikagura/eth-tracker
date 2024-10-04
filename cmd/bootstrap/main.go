package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/eth-tracker/internal/cache"
	"github.com/grassrootseconomics/eth-tracker/internal/chain"
	"github.com/grassrootseconomics/eth-tracker/internal/util"
	"github.com/grassrootseconomics/ethutils"
	"github.com/knadh/koanf/v2"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

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
	if err := bootstrapCache(); err != nil {
		lo.Error("critical error bootstrapping cache", "error", err)
		os.Exit(1)
	}
}

func bootstrapCache() error {
	var (
		tokenRegistryGetter = w3.MustNewFunc("tokenRegistry()", "address")
		quoterGetter        = w3.MustNewFunc("quoter()", "address")
	)

	chain, err := chain.NewRPCFetcher(chain.EthRPCOpts{
		RPCEndpoint: ko.MustString("chain.rpc_endpoint"),
		ChainID:     ko.MustInt64("chain.chainid"),
	})
	if err != nil {
		lo.Error("could not initialize chain client", "error", err)
		os.Exit(1)
	}

	cache, err := cache.New(cache.CacheOpts{
		Logg:      lo,
		CacheType: ko.MustString("core.cache_type"),
		RedisDSN:  ko.MustString("redis.dsn"),
	})
	if err != nil {
		lo.Error("could not initialize cache", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	for _, registry := range ko.MustStrings("bootstrap.ge_registries") {
		registryMap, err := chain.Provider().RegistryMap(ctx, ethutils.HexToAddress(registry))
		if err != nil {
			lo.Error("could not fetch registry", "error", err)
			os.Exit(1)
		}

		for k, v := range registryMap {
			if v != ethutils.ZeroAddress {
				if err := cache.Add(ctx, v.Hex()); err != nil {
					return err
				}

				lo.Debug("cached registry entry", "type", k, "address", v.Hex())
			}
		}

		if accountIndex := registryMap[ethutils.AccountIndex]; accountIndex != ethutils.ZeroAddress {
			if err := cache.Add(ctx, accountIndex.Hex()); err != nil {
				return err
			}

			lo.Debug("cached account index", "address", accountIndex.Hex())

			accountIndexIter, err := chain.Provider().NewBatchIterator(ctx, accountIndex)
			if err != nil {
				return err
			}
			for {
				accountIndexBatch, err := accountIndexIter.Next(ctx)
				if err != nil {
					return err
				}
				if accountIndexBatch == nil {
					break
				}

				for _, address := range accountIndexBatch {
					if address != ethutils.ZeroAddress {
						if err := cache.Add(ctx, address.Hex()); err != nil {
							return err
						}

					}
				}
				lo.Debug("cached account index batch", "batch_size", len(accountIndexBatch))
			}
		}

		if tokenIndex := registryMap[ethutils.TokenIndex]; tokenIndex != ethutils.ZeroAddress {
			if err := cache.Add(ctx, tokenIndex.Hex()); err != nil {
				return err
			}

			lo.Debug("cached token index", "address", tokenIndex.Hex())

			tokenIndexIter, err := chain.Provider().NewBatchIterator(ctx, tokenIndex)
			if err != nil {
				return err
			}
			for {
				tokenIndexBatch, err := tokenIndexIter.Next(ctx)
				if err != nil {
					return err
				}
				if tokenIndexBatch == nil {
					break
				}

				for _, address := range tokenIndexBatch {
					if address != ethutils.ZeroAddress {
						if err := cache.Add(ctx, address.Hex()); err != nil {
							return err
						}

					}
				}
				lo.Debug("cached token index batch", "batch_size", len(tokenIndexBatch))
			}
		}

		if poolIndex := registryMap[ethutils.PoolIndex]; poolIndex != ethutils.ZeroAddress {
			if err := cache.Add(ctx, poolIndex.Hex()); err != nil {
				return err
			}

			lo.Debug("cached pool index", "address", poolIndex.Hex())

			poolIndexIter, err := chain.Provider().NewBatchIterator(ctx, poolIndex)
			if err != nil {
				return err
			}
			for {
				poolIndexBatch, err := poolIndexIter.Next(ctx)
				if err != nil {
					return err
				}
				if poolIndexBatch == nil {
					break
				}
				for _, address := range poolIndexBatch {
					if address != ethutils.ZeroAddress {
						if err := cache.Add(ctx, address.Hex()); err != nil {
							return err
						}

						var poolTokenIndex, priceQuoter common.Address
						err := chain.Provider().Client.CallCtx(
							ctx,
							eth.CallFunc(address, tokenRegistryGetter).Returns(&poolTokenIndex),
							eth.CallFunc(address, quoterGetter).Returns(&priceQuoter),
						)
						if err != nil {
							return err
						}
						if priceQuoter != ethutils.ZeroAddress {
							if err := cache.Add(ctx, priceQuoter.Hex()); err != nil {
								return err
							}

							lo.Debug("cached pool index quoter", "pool", poolIndex.Hex(), "address", priceQuoter.Hex())
						}
						if poolTokenIndex != ethutils.ZeroAddress {
							if err := cache.Add(ctx, poolTokenIndex.Hex()); err != nil {
								return err
							}

							lo.Debug("cached pool index token index", "pool", poolIndex.Hex(), "address", poolTokenIndex.Hex())

							poolTokenIndexIter, err := chain.Provider().NewBatchIterator(ctx, poolTokenIndex)
							if err != nil {
								return err
							}
							for {
								poolTokenIndexBatch, err := poolTokenIndexIter.Next(ctx)
								if err != nil {
									return err
								}
								if poolTokenIndexBatch == nil {
									break
								}
								for _, address := range poolTokenIndexBatch {
									if address != ethutils.ZeroAddress {
										if err := cache.Add(ctx, address.Hex()); err != nil {
											return err
										}

									}
								}
								lo.Debug("cached pool token index batch", "batch_size", len(poolTokenIndexBatch))
							}
						}
					}
				}
				lo.Debug("cached pool index batch", "batch_size", len(poolIndexBatch))
			}
		}

		for _, address := range ko.MustStrings("bootstrap.watchlist") {
			if err := cache.Add(ctx, address); err != nil {
				return err
			}
		}
		for _, address := range ko.MustStrings("bootstrap.blacklist") {
			if err := cache.Remove(ctx, address); err != nil {
				return err
			}
		}
		if err := cache.Remove(ctx, ethutils.ZeroAddress.Hex()); err != nil {
			return err
		}
		cacheSize, err := cache.Size(ctx)
		if err != nil {
			return err
		}
		lo.Info("registry bootstrap complete", "registry", registry, "current_cache_size", cacheSize)
	}

	return nil
}
