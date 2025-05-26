package pool

import (
	"context"
	"log/slog"

	"github.com/alitto/pond/v2"
	"github.com/grassrootseconomics/eth-tracker/internal/processor"
)

type (
	PoolOpts struct {
		Logg        *slog.Logger
		WorkerCount int
		Processor   *processor.Processor
	}

	Pool struct {
		logg       *slog.Logger
		workerPool pond.Pool
		processor  *processor.Processor
	}
)

func New(o PoolOpts) *Pool {
	return &Pool{
		logg: o.Logg,
		workerPool: pond.NewPool(
			o.WorkerCount,
		),
		processor: o.Processor,
	}
}

func (p *Pool) Stop() {
	p.workerPool.StopAndWait()
}

// non-blocking
func (p *Pool) Push(block uint64) {
	p.workerPool.Submit(func() {
		err := p.processor.ProcessBlock(context.Background(), block)
		if err != nil {
			p.logg.Error("block processor error", "block_number", block, "error", err)
		}
	})
}

func (p *Pool) Size() uint64 {
	return p.workerPool.WaitingTasks()
}

func (p *Pool) ActiveWorkers() int64 {
	return p.workerPool.RunningWorkers()
}
