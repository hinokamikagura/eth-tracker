package chain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/grassrootseconomics/ethutils"
)

type Chain interface {
	GetBlocks(context.Context, []uint64) ([]*types.Block, error)
	GetBlock(context.Context, uint64) (*types.Block, error)
	GetLatestBlock(context.Context) (uint64, error)
	GetTransaction(context.Context, common.Hash) (*types.Transaction, error)
	GetReceipts(context.Context, *big.Int) (types.Receipts, error)
	// Expose provider until we eject from celoutils
	Provider() *ethutils.Provider
}
