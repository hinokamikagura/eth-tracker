package backfill

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/grassrootseconomics/celo-tracker/db"
	"github.com/grassrootseconomics/celo-tracker/internal/pool"
)

type (
	BackfillOpts struct {
		BatchSize int
		DB        db.DB
		Logg      *slog.Logger
		Pool      *pool.Pool
	}

	Backfill struct {
		batchSize int
		db        db.DB
		logg      *slog.Logger
		pool      *pool.Pool
		stopCh    chan struct{}
		ticker    *time.Ticker
	}
)

const (
	idleCheckInterval = 60 * time.Second
	busyCheckInterval = 1 * time.Second
)

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

func (b *Backfill) Stop() {
	b.ticker.Stop()
	b.stopCh <- struct{}{}
}

func (b *Backfill) Start() {
	for {
		select {
		case <-b.stopCh:
			b.logg.Debug("backfill shutting down")
			return
		case <-b.ticker.C:
			if b.pool.Size() <= 1 {
				if err := b.Run(true); err != nil {
					b.logg.Error("backfill run error", "err", err)
				}
				b.logg.Debug("backfill successful run", "queue_size", b.pool.Size())
			} else {
				b.logg.Debug("skipping backfill tick", "queue_size", b.pool.Size())
			}
		}
	}

}

func (b *Backfill) Run(skipLatest bool) error {
	lower, err := b.db.GetLowerBound()
	if err != nil {
		return fmt.Errorf("verifier could not get lower bound from db: err %v", err)
	}
	upper, err := b.db.GetUpperBound()
	if err != nil {
		return fmt.Errorf("verifier could not get upper bound from db: err %v", err)
	}

	if skipLatest {
		upper--
	}

	missingBlocks, err := b.db.GetMissingValuesBitSet(lower, upper)
	if err != nil {
		return fmt.Errorf("verifier could not get missing values bitset: err %v", err)
	}
	missingBlocksCount := missingBlocks.Count()

	if missingBlocksCount > 0 {
		b.logg.Info("found missing blocks", "skip_latest", skipLatest, "missing_blocks_count", missingBlocksCount)

		buffer := make([]uint, b.batchSize)
		j := uint(0)
		pushedCount := 0
		j, buffer = missingBlocks.NextSetMany(j, buffer)
		for ; len(buffer) > 0; j, buffer = missingBlocks.NextSetMany(j, buffer) {
			for k := range buffer {
				if pushedCount >= b.batchSize {
					break
				}

				b.pool.Push(uint64(buffer[k]))
				b.logg.Debug("pushed block from backfill", "block", buffer[k])
				pushedCount++
			}
			j++
		}
	}

	if missingBlocksCount > uint(b.batchSize) {
		b.ticker.Reset(busyCheckInterval)
	} else {
		b.ticker.Reset(idleCheckInterval)
	}

	missingBlocks.ClearAll()
	missingBlocks = nil
	b.logg.Debug("backfill tick run complete")

	return nil
}
