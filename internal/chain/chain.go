// Package chain provides blockchain data fetching capabilities for the eth-tracker service.
// It abstracts the underlying RPC client implementation and provides a clean interface
// for fetching blocks, transactions, receipts, and other chain data.
package chain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/grassrootseconomics/ethutils"
)

// Chain defines the interface for blockchain data access.
// Implementations should handle connection management, retries, and error handling.
type Chain interface {
	// GetBlocks fetches multiple blocks by their numbers in a single batch request.
	// Returns blocks in the same order as the input block numbers.
	GetBlocks(context.Context, []uint64) ([]*types.Block, error)

	// GetBlock fetches a single block by its number.
	GetBlock(context.Context, uint64) (*types.Block, error)

	// GetLatestBlock returns the latest block number from the chain.
	GetLatestBlock(context.Context) (uint64, error)

	// GetTransaction fetches a transaction by its hash.
	GetTransaction(context.Context, common.Hash) (*types.Transaction, error)

	// GetReceipts fetches all transaction receipts for a given block.
	GetReceipts(context.Context, *big.Int) (types.Receipts, error)

	// Provider returns the underlying ethutils provider instance.
	// This is exposed temporarily until we fully abstract away from ethutils.
	Provider() *ethutils.Provider
}
