package queue

//
// import (
// 	"context"
// 	"log/slog"

// 	"github.com/alitto/pond"
// 	"github.com/grassrootseconomics/celo-tracker/internal/processor"
// )

// type (
// 	QueueOpts struct {
// 		QueueSize int
// 		Logg      *slog.Logger
// 		Processor *processor.Processor
// 		Pool      *pond.WorkerPool
// 	}

// 	Queue struct {
// 		logg        *slog.Logger
// 		processChan chan uint64
// 		stopSignal  chan interface{}
// 		processor   *processor.Processor
// 		pool        *pond.WorkerPool
// 	}
// )

// func New(o QueueOpts) *Queue {
// 	return &Queue{
// 		logg:        o.Logg,
// 		processChan: make(chan uint64, o.QueueSize),
// 		stopSignal:  make(chan interface{}),
// 		processor:   o.Processor,
// 		pool:        o.Pool,
// 	}
// }

// func (q *Queue) Stop() {
// 	q.stopSignal <- struct{}{}
// }

// func (q *Queue) Process() {
// 	for {
// 		select {
// 		case <-q.stopSignal:
// 			q.logg.Info("shutdown signal received stopping queue processing")
// 			return
// 		case block, ok := <-q.processChan:
// 			if !ok {
// 				return
// 			}
// 			q.pool.Submit(func() {
// 				err := q.processor.ProcessBlock(context.Background(), block)
// 				if err != nil {
// 					q.logg.Error("block processor error", "block_number", block, "error", err)
// 				}
// 			})
// 		}
// 	}
// }

// func (q *Queue) Push(block uint64) {
// 	q.processChan <- block
// }

// func (q *Queue) Size() int {
// 	return len(q.processChan)
// }

// func (q *Queue) WaitingSize() uint64 {
// 	return q.pool.WaitingTasks()
// }
