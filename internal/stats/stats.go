// Package stats provides statistics collection and reporting for the eth-tracker service.
package stats

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/grassrootseconomics/eth-tracker/internal/cache"
	"github.com/grassrootseconomics/eth-tracker/internal/pool"
)

const (
	// statsPrinterInterval is the interval at which statistics are logged
	statsPrinterInterval = 15 * time.Second
)

type (
	// StatsOpts contains configuration options for creating a new Stats instance.
	StatsOpts struct {
		Cache cache.Cache  // Cache for size statistics
		Logg  *slog.Logger // Structured logger
		Pool  *pool.Pool   // Worker pool for queue statistics
	}

	// Stats collects and reports service statistics.
	Stats struct {
		cache       cache.Cache
		logg        *slog.Logger
		pool        *pool.Pool
		stopCh      chan struct{}
		latestBlock atomic.Uint64
	}
)

// New creates a new Stats instance.
func New(o StatsOpts) *Stats {
	return &Stats{
		cache:  o.Cache,
		logg:   o.Logg,
		pool:   o.Pool,
		stopCh: make(chan struct{}),
	}
}

// SetLatestBlock updates the latest processed block number.
func (s *Stats) SetLatestBlock(v uint64) {
	s.latestBlock.Store(v)
}

// GetLatestBlock returns the latest processed block number.
func (s *Stats) GetLatestBlock() uint64 {
	return s.latestBlock.Load()
}

// Stop stops the stats printer goroutine.
func (s *Stats) Stop() {
	close(s.stopCh)
	s.logg.Debug("stats stopped")
}

// APIStatsResponse returns current statistics as a map for API responses.
func (s *Stats) APIStatsResponse(ctx context.Context) (map[string]interface{}, error) {
	cacheSize, err := s.cache.Size(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"latestBlock":       s.GetLatestBlock(),
		"poolQueueSize":     s.pool.Size(),
		"poolActiveWorkers": s.pool.ActiveWorkers(),
		"cacheSize":         cacheSize,
	}, nil
}

// StartStatsPrinter starts a goroutine that periodically logs statistics.
func (s *Stats) StartStatsPrinter() {
	ticker := time.NewTicker(statsPrinterInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			s.logg.Debug("stats printer shutting down")
			return
		case <-ticker.C:
			cacheSize, err := s.cache.Size(context.Background())
			if err != nil {
				s.logg.Error("failed to fetch cache size", "error", err)
				continue
			}

			s.logg.Info("service statistics",
				"latest_block", s.GetLatestBlock(),
				"pool_queue_size", s.pool.Size(),
				"pool_active_workers", s.pool.ActiveWorkers(),
				"cache_size", cacheSize,
			)
		}
	}
}
