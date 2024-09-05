package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/grassrootseconomics/celo-tracker/db"
	"github.com/grassrootseconomics/celo-tracker/internal/api"
	"github.com/grassrootseconomics/celo-tracker/internal/backfill"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celo-tracker/internal/pool"
	"github.com/grassrootseconomics/celo-tracker/internal/processor"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/internal/router"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
	"github.com/grassrootseconomics/celo-tracker/internal/syncer"
	"github.com/grassrootseconomics/celo-tracker/internal/util"
	"github.com/knadh/koanf/v2"
)

const defaultGracefulShutdownPeriod = time.Second * 30

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

	lo.Info("starting celo tracker", "build", build)
}

func main() {
	var wg sync.WaitGroup
	ctx, stop := notifyShutdown()

	chain, err := chain.NewRPCFetcher(chain.CeloRPCOpts{
		RPCEndpoint: ko.MustString("chain.rpc_endpoint"),
		ChainID:     ko.MustInt64("chain.chainid"),
	})
	if err != nil {
		lo.Error("could not initialize chain client", "error", err)
		os.Exit(1)
	}

	db, err := db.New(db.DBOpts{
		Logg:   lo,
		DBType: ko.MustString("core.db_type"),
	})
	if err != nil {
		lo.Error("could not initialize blocks db", "error", err)
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

	jetStreamPub, err := pub.NewJetStreamPub(pub.JetStreamOpts{
		Endpoint:        ko.MustString("jetstream.endpoint"),
		PersistDuration: time.Duration(ko.MustInt("jetstream.persist_duration_hrs")) * time.Hour,
		DedupDuration:   time.Duration(ko.MustInt("jetstream.dedup_duration_hrs")) * time.Hour,
		Logg:            lo,
	})
	if err != nil {
		lo.Error("could not initialize jetstream pub", "error", err)
		os.Exit(1)
	}

	router := router.New(router.RouterOpts{
		Pub:   jetStreamPub,
		Cache: cache,
	})

	blockProcessor := processor.NewProcessor(processor.ProcessorOpts{
		Cache:  cache,
		Chain:  chain,
		DB:     db,
		Router: router,
		Logg:   lo,
	})

	poolOpts := pool.PoolOpts{
		Logg:        lo,
		WorkerCount: ko.Int("core.pool_size"),
		BatchSize:   ko.MustInt("core.batch_size"),
		Processor:   blockProcessor,
	}
	if ko.Int("core.pool_size") <= 0 {
		poolOpts.WorkerCount = runtime.NumCPU() * 3
	}
	workerPool := pool.New(poolOpts)

	stats := stats.New(stats.StatsOpts{
		Cache: cache,
		Logg:  lo,
		Pool:  workerPool,
	})

	chainSyncer, err := syncer.New(syncer.SyncerOpts{
		DB:                db,
		Chain:             chain,
		Logg:              lo,
		Pool:              workerPool,
		Stats:             stats,
		StartBlock:        ko.Int64("chain.start_block"),
		WebSocketEndpoint: ko.MustString("chain.ws_endpoint"),
	})
	if err != nil {
		lo.Error("could not initialize chain syncer", "error", err)
		os.Exit(1)
	}

	backfill := backfill.New(backfill.BackfillOpts{
		BatchSize: ko.MustInt("core.batch_size"),
		DB:        db,
		Logg:      lo,
		Pool:      workerPool,
	})

	apiServer := &http.Server{
		Addr:    ko.MustString("api.address"),
		Handler: api.New(),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		chainSyncer.Start()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := backfill.Run(false); err != nil {
			lo.Error("backfiller initial run error", "error", err)
		}
		backfill.Start()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("metrics and stats server starting", "address", ko.MustString("api.address"))
		if err := apiServer.ListenAndServe(); err != http.ErrServerClosed {
			lo.Error("failed to start API server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	lo.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		chainSyncer.Stop()
		backfill.Stop()
		workerPool.Stop()
		jetStreamPub.Close()
		db.Cleanup()
		db.Close()
		apiServer.Shutdown(shutdownCtx)
		lo.Info("graceful shutdown routine complete")
	}()

	go func() {
		wg.Wait()
		stop()
		cancel()
		os.Exit(0)
	}()

	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		stop()
		cancel()
		lo.Error("graceful shutdown period exceeded, forcefully shutting down")
	}
	os.Exit(1)
}

func notifyShutdown() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
}
