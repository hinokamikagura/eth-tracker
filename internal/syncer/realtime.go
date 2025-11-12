package syncer

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

const (
	// resubscribeInterval is the delay before attempting to resubscribe after a connection failure
	resubscribeInterval = 2 * time.Second
	// newHeadersBufferSize is the buffer size for the new headers channel
	newHeadersBufferSize = 1
)

// BlockQueueFn is a function type for queuing blocks for processing
type BlockQueueFn func(uint64) error

// Stop stops the real-time subscription and closes the WebSocket connection.
func (s *Syncer) Stop() {
	if s.realtimeSub != nil {
		s.realtimeSub.Unsubscribe()
		s.logg.Info("realtime subscription stopped")
	}
}

// Start begins subscribing to new block headers and processing them in real-time.
// It automatically resubscribes on connection failures.
func (s *Syncer) Start() {
	s.realtimeSub = event.ResubscribeErr(resubscribeInterval, s.resubscribeFn())
	s.logg.Info("realtime syncer started")
}

// receiveRealtimeBlocks subscribes to new block headers and processes them.
func (s *Syncer) receiveRealtimeBlocks(ctx context.Context, fn BlockQueueFn) (ethereum.Subscription, error) {
	newHeadersReceiver := make(chan *types.Header, newHeadersBufferSize)
	sub, err := s.ethClient.SubscribeNewHead(ctx, newHeadersReceiver)
	if err != nil {
		return nil, err
	}

	s.logg.Info("realtime syncer connected to WebSocket endpoint")

	return event.NewSubscription(func(quit <-chan struct{}) error {
		eventsCtx, eventsCancel := context.WithCancel(context.Background())
		defer eventsCancel()

		// Handle shutdown signal
		go func() {
			select {
			case <-quit:
				s.logg.Info("realtime syncer stopping")
				eventsCancel()
			case <-eventsCtx.Done():
				return
			}
		}()

		for {
			select {
			case header := <-newHeadersReceiver:
				blockNumber := header.Number.Uint64()
				if err := fn(blockNumber); err != nil {
					s.logg.Error("failed to queue realtime block",
						"block_number", blockNumber,
						"error", err,
					)
				} else {
					s.logg.Debug("queued new block", "block_number", blockNumber)
				}
			case <-eventsCtx.Done():
				s.logg.Info("realtime syncer shutting down")
				return nil
			case err := <-sub.Err():
				if err != nil {
					s.logg.Error("subscription error", "error", err)
				}
				return err
			}
		}
	}), nil
}

// queueRealtimeBlock queues a new block for processing and updates state.
func (s *Syncer) queueRealtimeBlock(blockNumber uint64) error {
	s.pool.Push(blockNumber)
	s.stats.SetLatestBlock(blockNumber)

	if err := s.db.SetUpperBound(blockNumber); err != nil {
		return err
	}

	return nil
}

// resubscribeFn returns a function that handles resubscription on connection failures.
func (s *Syncer) resubscribeFn() event.ResubscribeErrFunc {
	return func(ctx context.Context, err error) (event.Subscription, error) {
		if err != nil {
			s.logg.Warn("resubscribing after connection failure", "error", err)
		}
		return s.receiveRealtimeBlocks(ctx, s.queueRealtimeBlock)
	}
}
