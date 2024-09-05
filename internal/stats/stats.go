package stats

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/pool"
)

type (
	StatsOpts struct {
		Cache cache.Cache
		Logg  *slog.Logger
		Pool  *pool.Pool
	}

	Stats struct {
		cache       cache.Cache
		logg        *slog.Logger
		pool        *pool.Pool
		stopCh      chan struct{}
		latestBlock atomic.Uint64
	}
)

const statsPrinterInterval = 15 * time.Second

func New(o StatsOpts) *Stats {
	return &Stats{
		cache:  o.Cache,
		logg:   o.Logg,
		pool:   o.Pool,
		stopCh: make(chan struct{}),
	}
}

func (s *Stats) SetLatestBlock(v uint64) {
	s.latestBlock.Store(v)
}

func (s *Stats) GetLatestBlock() uint64 {
	return s.latestBlock.Load()
}

func (s *Stats) Stop() {
	s.stopCh <- struct{}{}
}

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

func (s *Stats) StartStatsPrinter() {
	ticker := time.NewTicker(statsPrinterInterval)

	for {
		select {
		case <-s.stopCh:
			s.logg.Debug("stats shutting down")
			return
		case <-ticker.C:
			cacheSize, err := s.cache.Size(context.Background())
			if err != nil {
				s.logg.Error("stats printer could not fetch cache size", "error", err)
			}

			s.logg.Info("block stats",
				"latest_block", s.GetLatestBlock(),
				"pool_queue_size", s.pool.Size(),
				"pool_active_workers", s.pool.ActiveWorkers(),
				"cache_size", cacheSize,
			)
		}
	}
}
