// Package backfill provides functionality for processing missed blocks.
// It periodically checks for gaps in processed blocks and queues them for reprocessing.
package backfill

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/grassrootseconomics/eth-tracker/db"
	"github.com/grassrootseconomics/eth-tracker/internal/pool"
)

const (
	// idleCheckInterval is the interval between backfill checks when the queue is idle
	idleCheckInterval = 60 * time.Second
	// busyCheckInterval is the interval between backfill checks when there are many missing blocks
	busyCheckInterval = 250 * time.Millisecond
	// minQueueSizeForIdleCheck is the maximum queue size to consider the system idle
	minQueueSizeForIdleCheck = 1
)

type (
	// BackfillOpts contains configuration options for creating a new Backfill.
	BackfillOpts struct {
		BatchSize int          // Maximum number of blocks to queue per backfill run
		DB        db.DB        // Database for block state management
		Logg      *slog.Logger // Structured logger
		Pool      *pool.Pool   // Worker pool for block processing
	}

	// Backfill manages periodic backfilling of missed blocks.
	Backfill struct {
		batchSize int
		db        db.DB
		logg      *slog.Logger
		pool      *pool.Pool
		stopCh    chan struct{}
		ticker    *time.Ticker
	}
)

// New creates a new Backfill instance with the provided options.
func New(o BackfillOpts) *Backfill {
	return &Backfill{
		batchSize: o.BatchSize,
		db:        o.DB,
		logg:      o.Logg,
		pool:      o.Pool,
		stopCh:    make(chan struct{}),
		ticker:    time.NewTicker(idleCheckInterval),
	}
}

// Stop stops the backfill ticker and signals shutdown.
func (b *Backfill) Stop() {
	b.ticker.Stop()
	close(b.stopCh)
	b.logg.Info("backfill stopped")
}

// Start begins periodic backfill processing.
// It checks for missing blocks at intervals based on queue load.
func (b *Backfill) Start() {
	b.logg.Info("backfill started", "batch_size", b.batchSize)

	for {
		select {
		case <-b.stopCh:
			b.logg.Debug("backfill shutting down")
			return
		case <-b.ticker.C:
			queueSize := b.pool.Size()
			if queueSize <= minQueueSizeForIdleCheck {
				if err := b.Run(true); err != nil {
					b.logg.Error("backfill run failed", "error", err)
				} else {
					b.logg.Debug("backfill run completed", "queue_size", queueSize)
				}
			} else {
				b.logg.Debug("skipping backfill tick due to busy queue", "queue_size", queueSize)
			}
		}
	}
}

// Run performs a single backfill operation, finding and queuing missing blocks.
// If skipLatest is true, the latest block is excluded from the range check.
func (b *Backfill) Run(skipLatest bool) error {
	lower, err := b.db.GetLowerBound()
	if err != nil {
		return fmt.Errorf("failed to get lower bound: %w", err)
	}

	upper, err := b.db.GetUpperBound()
	if err != nil {
		return fmt.Errorf("failed to get upper bound: %w", err)
	}

	if skipLatest && upper > lower {
		upper--
	}

	if upper < lower {
		return nil
	}

	missingBlocks, err := b.db.GetMissingValuesBitSet(lower, upper)
	if err != nil {
		return fmt.Errorf("failed to get missing blocks bitset: %w", err)
	}

	missingBlocksCount := missingBlocks.Count()
	if missingBlocksCount == 0 {
		missingBlocks.ClearAll()
		return nil
	}

	b.logg.Info("found missing blocks",
		"skip_latest", skipLatest,
		"missing_count", missingBlocksCount,
		"range", fmt.Sprintf("%d-%d", lower, upper),
	)

	// Process missing blocks in batches
	buffer := make([]uint, b.batchSize)
	j := uint(0)
	pushedCount := 0

	j, buffer = missingBlocks.NextSetMany(j, buffer)
	for len(buffer) > 0 && pushedCount < b.batchSize {
		for _, blockIdx := range buffer {
			if pushedCount >= b.batchSize {
				break
			}

			blockNumber := uint64(blockIdx)
			b.pool.Push(blockNumber)
			b.logg.Debug("queued missing block", "block", blockNumber)
			pushedCount++
		}

		if pushedCount < b.batchSize {
			j++
			j, buffer = missingBlocks.NextSetMany(j, buffer)
		} else {
			break
		}
	}

	// Adjust ticker interval based on remaining missing blocks
	if missingBlocksCount > uint(b.batchSize) {
		b.ticker.Reset(busyCheckInterval)
		b.logg.Debug("switched to busy check interval", "remaining", missingBlocksCount-uint(pushedCount))
	} else {
		b.ticker.Reset(idleCheckInterval)
	}

	missingBlocks.ClearAll()
	b.logg.Debug("backfill run complete", "queued", pushedCount, "remaining", missingBlocksCount-uint(pushedCount))

	return nil
}
