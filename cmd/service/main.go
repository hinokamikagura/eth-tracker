// Package main provides the entry point for the eth-tracker service.
// eth-tracker is a high-performance EVM blockchain event tracker that monitors
// transactions and events, filters them based on configured addresses, and
// publishes filtered events to NATS JetStream for downstream processing.
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

	"github.com/grassrootseconomics/eth-tracker/db"
	"github.com/grassrootseconomics/eth-tracker/internal/api"
	"github.com/grassrootseconomics/eth-tracker/internal/backfill"
	"github.com/grassrootseconomics/eth-tracker/internal/cache"
	"github.com/grassrootseconomics/eth-tracker/internal/chain"
	"github.com/grassrootseconomics/eth-tracker/internal/pool"
	"github.com/grassrootseconomics/eth-tracker/internal/processor"
	"github.com/grassrootseconomics/eth-tracker/internal/pub"
	"github.com/grassrootseconomics/eth-tracker/internal/stats"
	"github.com/grassrootseconomics/eth-tracker/internal/syncer"
	"github.com/grassrootseconomics/eth-tracker/internal/util"
	"github.com/knadh/koanf/v2"
)

const (
	// defaultGracefulShutdownPeriod defines the maximum time allowed for graceful shutdown
	// before forcefully terminating the application.
	defaultGracefulShutdownPeriod = time.Second * 30

	// defaultWorkerPoolMultiplier is the multiplier used to calculate default worker pool size
	// based on CPU count when pool_size is not explicitly configured.
	defaultWorkerPoolMultiplier = 3
)

var (
	// build is set during compilation via -ldflags "-X main.build=<version>"
	build = "dev"

	// confFlag holds the path to the configuration file
	confFlag string

	// lo is the global structured logger instance
	lo *slog.Logger

	// ko is the global configuration instance
	ko *koanf.Koanf
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Path to configuration file (TOML format)")
	flag.Parse()

	lo = util.InitLogger()
	ko = util.InitConfig(lo, confFlag)
}

// main initializes and starts all service components including:
// - Chain RPC client for blockchain data fetching
// - Database for block tracking and state persistence
// - Cache for address filtering
// - NATS JetStream publisher for event distribution
// - Event router for processing blockchain events
// - Block processor for transaction and log processing
// - Worker pool for concurrent block processing
// - Stats collector for monitoring
// - Real-time chain syncer for new blocks
// - Backfill processor for missed blocks
// - HTTP API server for metrics and health checks
func main() {
	lo.Info("starting eth-tracker service", "build", build, "version", build)

	var wg sync.WaitGroup
	ctx, stop := notifyShutdown()

	chain, err := chain.NewRPCFetcher(chain.EthRPCOpts{
		RPCEndpoint: ko.MustString("chain.rpc_endpoint"),
		ChainID:     ko.MustInt64("chain.chainid"),
	})
	if err != nil {
		lo.Error("could not initialize chain client", "error", err)
		os.Exit(1)
	}
	lo.Debug("loaded rpc fetcher")

	db, err := db.New(db.DBOpts{
		Logg:   lo,
		DBType: ko.MustString("core.db_type"),
	})
	if err != nil {
		lo.Error("could not initialize blocks db", "error", err)
		os.Exit(1)
	}
	lo.Debug("loaded blocks db")

	cacheOpts := cache.CacheOpts{
		Chain:      chain,
		Registries: ko.MustStrings("bootstrap.ge_registry"),
		Watchlist:  ko.Strings("bootstrap.watchlist"),
		Blacklist:  ko.Strings("bootstrap.blacklist"),
		CacheType:  ko.MustString("core.cache_type"),
		Logg:       lo,
	}
	if ko.MustString("core.cache_type") == "redis" {
		cacheOpts.RedisDSN = ko.MustString("redis.dsn")
	}
	cache, err := cache.New(cacheOpts)
	if err != nil {
		lo.Error("could not initialize cache", "error", err)
		os.Exit(1)
	}
	lo.Debug("loaded and boostrapped cache")

	jetStreamPub, err := pub.NewJetStreamPub(pub.JetStreamOpts{
		Endpoint:        ko.MustString("jetstream.endpoint"),
		PersistDuration: time.Duration(ko.MustInt("jetstream.persist_duration_hrs")) * time.Hour,
		Logg:            lo,
	})
	if err != nil {
		lo.Error("could not initialize jetstream pub", "error", err)
		os.Exit(1)
	}
	lo.Debug("loaded jetstream publisher")

	router := bootstrapEventRouter(cache, jetStreamPub.Send)
	lo.Debug("bootstrapped event router")

	blockProcessor := processor.NewProcessor(processor.ProcessorOpts{
		Cache:  cache,
		Chain:  chain,
		DB:     db,
		Router: router,
		Logg:   lo,
	})
	lo.Debug("bootstrapped processor")

	poolSize := ko.Int("core.pool_size")
	if poolSize <= 0 {
		poolSize = runtime.NumCPU() * defaultWorkerPoolMultiplier
		lo.Info("using default worker pool size", "cpu_count", runtime.NumCPU(), "pool_size", poolSize)
	}
	poolOpts := pool.PoolOpts{
		Logg:        lo,
		WorkerCount: poolSize,
		Processor:   blockProcessor,
	}
	workerPool := pool.New(poolOpts)
	lo.Debug("bootstrapped worker pool")

	stats := stats.New(stats.StatsOpts{
		Cache: cache,
		Logg:  lo,
		Pool:  workerPool,
	})
	lo.Debug("bootstrapped stats provider")

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
	lo.Debug("bootstrapped realtime syncer")

	backfill := backfill.New(backfill.BackfillOpts{
		BatchSize: ko.MustInt("core.batch_size"),
		DB:        db,
		Logg:      lo,
		Pool:      workerPool,
	})
	lo.Debug("bootstrapped backfiller")

	apiServer := &http.Server{
		Addr:    ko.MustString("api.address"),
		Handler: api.New(),
	}
	lo.Debug("bootstrapped API server")
	lo.Debug("starting routines")

	// Start real-time chain syncer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		chainSyncer.Start()
		lo.Info("chain syncer started")
	}()

	// Start backfill processor goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := backfill.Run(false); err != nil {
			lo.Error("backfiller initial run error", "error", err)
		} else {
			lo.Info("completed initial backfill run")
		}
		backfill.Start()
		lo.Info("periodic backfiller started")
	}()

	// Start HTTP API server goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		apiAddr := ko.MustString("api.address")
		lo.Info("starting API server", "address", apiAddr)
		if err := apiServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lo.Error("API server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	lo.Info("shutdown signal received, initiating graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)
	defer cancel()

	// Perform graceful shutdown in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("stopping service components")
		chainSyncer.Stop()
		backfill.Stop()
		workerPool.Stop()
		jetStreamPub.Close()
		if err := db.Cleanup(); err != nil {
			lo.Error("database cleanup error", "error", err)
		}
		if err := db.Close(); err != nil {
			lo.Error("database close error", "error", err)
		}
		if err := apiServer.Shutdown(shutdownCtx); err != nil {
			lo.Error("API server shutdown error", "error", err)
		}
		lo.Info("graceful shutdown complete")
	}()

	// Wait for shutdown completion or timeout
	shutdownDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		stop()
		lo.Info("service stopped successfully")
		os.Exit(0)
	case <-shutdownCtx.Done():
		if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
			stop()
			lo.Error("graceful shutdown timeout exceeded, forcing exit")
			os.Exit(1)
		}
	}
}

// notifyShutdown creates a context that is cancelled when the application receives
// a shutdown signal (SIGINT, SIGTERM, or interrupt).
func notifyShutdown() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
}
