// Package pool provides a worker pool for concurrent block processing.
// It manages a pool of goroutines that process blocks in parallel, improving
// throughput for high-volume blockchain event tracking.
package pool

import (
	"context"
	"log/slog"

	"github.com/alitto/pond/v2"
	"github.com/grassrootseconomics/eth-tracker/internal/processor"
)

type (
	// PoolOpts contains configuration options for creating a new Pool.
	PoolOpts struct {
		Logg        *slog.Logger         // Structured logger
		WorkerCount int                  // Number of worker goroutines
		Processor   *processor.Processor // Block processor instance
	}

	// Pool manages a worker pool for concurrent block processing.
	Pool struct {
		logg       *slog.Logger
		workerPool pond.Pool
		processor  *processor.Processor
	}
)

// New creates a new Pool instance with the specified number of workers.
func New(o PoolOpts) *Pool {
	return &Pool{
		logg: o.Logg,
		workerPool: pond.NewPool(
			o.WorkerCount,
		),
		processor: o.Processor,
	}
}

// Stop gracefully stops the worker pool, waiting for all in-flight tasks to complete.
func (p *Pool) Stop() {
	p.workerPool.StopAndWait()
}

// Push submits a block for processing asynchronously (non-blocking).
// The block will be processed by an available worker from the pool.
func (p *Pool) Push(block uint64) {
	p.workerPool.Submit(func() {
		ctx := context.Background()
		if err := p.processor.ProcessBlock(ctx, block); err != nil {
			p.logg.Error("block processing failed",
				"block_number", block,
				"error", err,
			)
		}
	})
}

// Size returns the number of tasks currently waiting in the queue.
func (p *Pool) Size() uint64 {
	return p.workerPool.WaitingTasks()
}

// ActiveWorkers returns the number of workers currently processing tasks.
func (p *Pool) ActiveWorkers() int64 {
	return p.workerPool.RunningWorkers()
}
