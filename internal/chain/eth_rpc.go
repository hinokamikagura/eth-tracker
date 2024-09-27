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

type (
	EthRPCOpts struct {
		RPCEndpoint string
		ChainID     int64
	}

	EthRPC struct {
		provider *ethutils.Provider
	}
)

func NewRPCFetcher(o EthRPCOpts) (Chain, error) {
	customRPCClient, err := lowTimeoutRPCClient(o.RPCEndpoint)
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

func lowTimeoutRPCClient(rpcEndpoint string) (*w3.Client, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	rpcClient, err := rpc.DialOptions(context.Background(), rpcEndpoint, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return w3.NewClient(rpcClient), nil
}

func (c *EthRPC) GetBlocks(ctx context.Context, blockNumbers []uint64) ([]*types.Block, error) {
	blocksCount := len(blockNumbers)
	calls := make([]w3types.RPCCaller, blocksCount)
	blocks := make([]*types.Block, blocksCount)

	for i, v := range blockNumbers {
		calls[i] = eth.BlockByNumber(new(big.Int).SetUint64(v)).Returns(&blocks[i])
	}

	if err := c.provider.Client.CallCtx(ctx, calls...); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (c *EthRPC) GetBlock(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	var block *types.Block
	blockCall := eth.BlockByNumber(new(big.Int).SetUint64(blockNumber)).Returns(&block)

	if err := c.provider.Client.CallCtx(ctx, blockCall); err != nil {
		return nil, err
	}

	return block, nil
}

func (c *EthRPC) GetLatestBlock(ctx context.Context) (uint64, error) {
	var latestBlock *big.Int
	latestBlockCall := eth.BlockNumber().Returns(&latestBlock)

	if err := c.provider.Client.CallCtx(ctx, latestBlockCall); err != nil {
		return 0, err
	}

	return latestBlock.Uint64(), nil
}

func (c *EthRPC) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, error) {
	var transaction *types.Transaction
	if err := c.provider.Client.CallCtx(ctx, eth.Tx(txHash).Returns(&transaction)); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (c *EthRPC) GetReceipts(ctx context.Context, blockNumber *big.Int) (types.Receipts, error) {
	var receipts types.Receipts

	if err := c.provider.Client.CallCtx(ctx, eth.BlockReceipts(blockNumber).Returns(&receipts)); err != nil {
		return nil, err
	}

	return receipts, nil
}

func (c *EthRPC) Provider() *ethutils.Provider {
	return c.provider
}
