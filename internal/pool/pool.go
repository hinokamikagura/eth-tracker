package pool

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/alitto/pond"
	"github.com/grassrootseconomics/celo-tracker/internal/processor"
)

type (
	PoolOpts struct {
		Logg        *slog.Logger
		WorkerCount int
		Processor   *processor.Processor
	}

	Pool struct {
		logg       *slog.Logger
		workerPool *pond.WorkerPool
		processor  *processor.Processor
	}
)

const blocksBuffer = 100

func New(o PoolOpts) *Pool {
	return &Pool{
		logg: o.Logg,
		workerPool: pond.New(
			o.WorkerCount,
			blocksBuffer,
			pond.Strategy(pond.Balanced()),
			pond.PanicHandler(panicHandler(o.Logg)),
		),
		processor: o.Processor,
	}
}

func (p *Pool) Stop() {
	p.workerPool.StopAndWait()
}

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

func (p *Pool) ActiveWorkers() int {
	return p.workerPool.RunningWorkers()
}

func panicHandler(logg *slog.Logger) func(interface{}) {
	return func(panic interface{}) {
		logg.Error("block processor goroutine exited from a panic", "error", panic, "stack_trace", string(debug.Stack()))
	}
}
