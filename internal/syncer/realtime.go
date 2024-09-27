package syncer

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

type BlockQueueFn func(uint64) error

const resubscribeInterval = 2 * time.Second

func (s *Syncer) Stop() {
	if s.realtimeSub != nil {
		s.realtimeSub.Unsubscribe()
	}
}

func (s *Syncer) Start() {
	s.realtimeSub = event.ResubscribeErr(resubscribeInterval, s.resubscribeFn())
}

func (s *Syncer) receiveRealtimeBlocks(ctx context.Context, fn BlockQueueFn) (ethereum.Subscription, error) {
	newHeadersReceiver := make(chan *types.Header, 1)
	sub, err := s.ethClient.SubscribeNewHead(ctx, newHeadersReceiver)
	s.logg.Info("realtime syncer connected to ws endpoint")
	if err != nil {
		return nil, err
	}

	return event.NewSubscription(func(quit <-chan struct{}) error {
		eventsCtx, eventsCancel := context.WithCancel(context.Background())
		defer eventsCancel()

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
				if err := fn(header.Number.Uint64()); err != nil {
					s.logg.Error("realtime block queuer error", "error", err)
				}
			case <-eventsCtx.Done():
				s.logg.Info("realtime syncer shutting down")
				return nil
			case err := <-sub.Err():
				return err
			}
		}
	}), nil
}

func (s *Syncer) queueRealtimeBlock(blockNumber uint64) error {
	s.pool.Push(blockNumber)
	s.stats.SetLatestBlock(blockNumber)
	if err := s.db.SetUpperBound(blockNumber); err != nil {
		return err
	}
	return nil
}

func (s *Syncer) resubscribeFn() event.ResubscribeErrFunc {
	return func(ctx context.Context, err error) (event.Subscription, error) {
		if err != nil {
			s.logg.Error("resubscribing after failed subscription", "error", err)
		}
		return s.receiveRealtimeBlocks(ctx, s.queueRealtimeBlock)
	}
}
