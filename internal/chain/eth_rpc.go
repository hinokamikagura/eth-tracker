package chain

import (
	"context"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/grassrootseconomics/ethutils"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
)

const (
	// defaultRPCClientTimeout is the default HTTP client timeout for RPC requests.
	defaultRPCClientTimeout = 10 * time.Second
)

type (
	// EthRPCOpts contains configuration options for creating a new EthRPC client.
	EthRPCOpts struct {
		RPCEndpoint string // RPC endpoint URL (HTTP)
		ChainID     int64  // Chain ID for transaction signing
	}

	// EthRPC implements the Chain interface using Ethereum RPC calls.
	// It uses the w3 library for efficient batch RPC requests.
	EthRPC struct {
		provider *ethutils.Provider
	}
)

// NewRPCFetcher creates a new Chain implementation using HTTP RPC.
// It configures a low-timeout HTTP client for fast failure detection.
func NewRPCFetcher(o EthRPCOpts) (Chain, error) {
	customRPCClient, err := newRPCClient(o.RPCEndpoint)
	if err != nil {
		return nil, err
	}

	chainProvider := ethutils.NewProvider(
		o.RPCEndpoint,
		o.ChainID,
		ethutils.WithClient(customRPCClient),
	)

	return &EthRPC{
		provider: chainProvider,
	}, nil
}

// newRPCClient creates a new w3 RPC client with a configured HTTP client.
func newRPCClient(rpcEndpoint string) (*w3.Client, error) {
	httpClient := &http.Client{
		Timeout: defaultRPCClientTimeout,
	}

	rpcClient, err := rpc.DialOptions(context.Background(), rpcEndpoint, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return w3.NewClient(rpcClient), nil
}

// GetBlocks fetches multiple blocks in a single batch RPC call for efficiency.
func (c *EthRPC) GetBlocks(ctx context.Context, blockNumbers []uint64) ([]*types.Block, error) {
	if len(blockNumbers) == 0 {
		return nil, nil
	}

	blocksCount := len(blockNumbers)
	calls := make([]w3types.RPCCaller, blocksCount)
	blocks := make([]*types.Block, blocksCount)

	for i, blockNum := range blockNumbers {
		calls[i] = eth.BlockByNumber(new(big.Int).SetUint64(blockNum)).Returns(&blocks[i])
	}

	if err := c.provider.Client.CallCtx(ctx, calls...); err != nil {
		return nil, err
	}

	return blocks, nil
}

// GetBlock fetches a single block by its number.
func (c *EthRPC) GetBlock(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	var block *types.Block
	blockCall := eth.BlockByNumber(new(big.Int).SetUint64(blockNumber)).Returns(&block)

	if err := c.provider.Client.CallCtx(ctx, blockCall); err != nil {
		return nil, err
	}

	return block, nil
}

// GetLatestBlock returns the latest block number from the chain.
func (c *EthRPC) GetLatestBlock(ctx context.Context) (uint64, error) {
	var latestBlock *big.Int
	latestBlockCall := eth.BlockNumber().Returns(&latestBlock)

	if err := c.provider.Client.CallCtx(ctx, latestBlockCall); err != nil {
		return 0, err
	}

	return latestBlock.Uint64(), nil
}

// GetTransaction fetches a transaction by its hash.
func (c *EthRPC) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, error) {
	var transaction *types.Transaction
	if err := c.provider.Client.CallCtx(ctx, eth.Tx(txHash).Returns(&transaction)); err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetReceipts fetches all transaction receipts for a given block.
func (c *EthRPC) GetReceipts(ctx context.Context, blockNumber *big.Int) (types.Receipts, error) {
	var receipts types.Receipts

	if err := c.provider.Client.CallCtx(ctx, eth.BlockReceipts(blockNumber).Returns(&receipts)); err != nil {
		return nil, err
	}

	return receipts, nil
}

// Provider returns the underlying ethutils provider instance.
func (c *EthRPC) Provider() *ethutils.Provider {
	return c.provider
}
