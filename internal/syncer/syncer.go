// Package syncer provides real-time blockchain synchronization functionality.
// It subscribes to new block headers via WebSocket and queues them for processing.
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

const (
	// defaultStartBlock indicates to start from the latest block
	defaultStartBlock = 0
)

type (
	// SyncerOpts contains configuration options for creating a new Syncer.
	SyncerOpts struct {
		DB                db.DB        // Database for block state management
		Chain             chain.Chain  // Chain client for fetching latest block
		Logg              *slog.Logger // Structured logger
		Pool              *pool.Pool   // Worker pool for block processing
		Stats             *stats.Stats // Statistics collector
		StartBlock        int64        // Starting block number (0 = latest)
		WebSocketEndpoint string       // WebSocket RPC endpoint for real-time subscriptions
	}

	// Syncer manages real-time blockchain synchronization via WebSocket subscriptions.
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

// New creates a new Syncer instance and initializes block bounds.
// It sets the lower bound from the database or start block config, and upper bound to latest.
func New(o SyncerOpts) (*Syncer, error) {
	ctx := context.Background()

	// Fetch latest block from chain
	latestBlock, err := o.Chain.GetLatestBlock(ctx)
	if err != nil {
		return nil, err
	}

	// Initialize lower bound if not set
	lowerBound, err := o.DB.GetLowerBound()
	if err != nil {
		return nil, err
	}

	if lowerBound == 0 {
		if o.StartBlock > defaultStartBlock {
			lowerBound = uint64(o.StartBlock)
		} else {
			lowerBound = latestBlock
		}

		if err := o.DB.SetLowerBound(lowerBound); err != nil {
			return nil, err
		}
		o.Logg.Info("initialized lower bound", "block", lowerBound)
	}

	// Set upper bound to latest block
	if err := o.DB.SetUpperBound(latestBlock); err != nil {
		return nil, err
	}
	o.Stats.SetLatestBlock(latestBlock)

	// Connect to WebSocket endpoint
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
