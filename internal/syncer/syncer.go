package syncer

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/grassrootseconomics/eth-tracker/db"
	"github.com/grassrootseconomics/eth-tracker/internal/chain"
	"github.com/grassrootseconomics/eth-tracker/internal/pool"
	"github.com/grassrootseconomics/eth-tracker/internal/stats"
)

type (
	SyncerOpts struct {
		DB                db.DB
		Chain             chain.Chain
		Logg              *slog.Logger
		Pool              *pool.Pool
		Stats             *stats.Stats
		StartBlock        int64
		WebSocketEndpoint string
	}

	Syncer struct {
		db          db.DB
		ethClient   *ethclient.Client
		logg        *slog.Logger
		realtimeSub ethereum.Subscription
		pool        *pool.Pool
		stats       *stats.Stats
		stopCh      chan struct{}
	}
)

func New(o SyncerOpts) (*Syncer, error) {
	latestBlock, err := o.Chain.GetLatestBlock(context.Background())
	if err != nil {
		return nil, err
	}

	lowerBound, err := o.DB.GetLowerBound()
	if err != nil {
		return nil, err
	}
	if lowerBound == 0 {
		if o.StartBlock > 0 {
			if err := o.DB.SetLowerBound(uint64(o.StartBlock)); err != nil {
				return nil, err
			}
		} else {
			if err := o.DB.SetLowerBound(latestBlock); err != nil {
				return nil, err
			}
		}
	}

	if err := o.DB.SetUpperBound(latestBlock); err != nil {
		return nil, err
	}
	o.Stats.SetLatestBlock(latestBlock)

	ethClient, err := ethclient.Dial(o.WebSocketEndpoint)
	if err != nil {
		return nil, err
	}

	return &Syncer{
		db:        o.DB,
		ethClient: ethClient,
		logg:      o.Logg,
		pool:      o.Pool,
		stats:     o.Stats,
		stopCh:    make(chan struct{}),
	}, nil
}
