package chain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testRPCEndpoint = "https://alfajores-forno.celo-testnet.org"
	testChainID     = 44787
)

func setupEthRPC() (Chain, error) {
	opts := EthRPCOpts{
		RPCEndpoint: testRPCEndpoint,
		ChainID:     testChainID,
	}
	return NewRPCFetcher(opts)
}

func TestRPC_GetBlocks(t *testing.T) {
	rpcFetcher, err := setupEthRPC()
	require.NoError(t, err)

	blockNumbers := []uint64{
		19_600_000,
		23_000_000,
		26_385_000,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blocks, err := rpcFetcher.GetBlocks(ctx, blockNumbers)
	require.NoError(t, err)
	t.Logf("blocks %+v\n", blocks)
}

func TestRPC_GetBlock(t *testing.T) {
	rpcFetcher, err := setupEthRPC()
	require.NoError(t, err)

	var blockNumber uint64 = 26_386_000
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	block, err := rpcFetcher.GetBlock(ctx, blockNumber)
	require.NoError(t, err)
	t.Logf("block %+v\n", block)
}

func TestRPC_GetLatestBlock(t *testing.T) {
	rpcFetcher, err := setupEthRPC()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	block, err := rpcFetcher.GetLatestBlock(ctx)
	require.NoError(t, err)
	t.Logf("block %+v\n", block)
}

func TestRPC_GetTransaction(t *testing.T) {
	rpcFetcher, err := setupEthRPC()
	require.NoError(t, err)

	var blockNumber uint64 = 26_387_000
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	block, err := rpcFetcher.GetBlock(ctx, blockNumber)
	require.NoError(t, err)
	require.NotNil(t, block)

	transaction, err := rpcFetcher.GetTransaction(ctx, block.Transactions()[0].Hash())
	require.NoError(t, err)
	t.Logf("transaction %+v\n", transaction)
}

func TestRPC_GetReceipts(t *testing.T) {
	rpcFetcher, err := setupEthRPC()
	require.NoError(t, err)

	var blockNumber uint64 = 26_388_000
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	block, err := rpcFetcher.GetBlock(ctx, blockNumber)
	require.NoError(t, err)
	require.NotNil(t, block)

	receipts, err := rpcFetcher.GetReceipts(ctx, block.Number())
	require.NoError(t, err)
	t.Logf("receipts %+v\n", receipts)
}
