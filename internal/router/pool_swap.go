package router

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type poolSwapHandler struct{}

const poolSwapEventName = "POOL_SWAP"

var (
	_ Handler = (*poolSwapHandler)(nil)

	poolSwapEvent = w3.MustNewEvent("Swap(address indexed initiator, address indexed tokenIn, address tokenOut, uint256 amountIn, uint256 amountOut, uint256 fee)")
	poolSwapSig   = w3.MustNewFunc("withdraw(address, address, uint256)", "")
)

func (h *poolSwapHandler) Name() string {
	return poolSwapEventName
}

func (h *poolSwapHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var (
		initiator common.Address
		tokenIn   common.Address
		tokenOut  common.Address
		amountIn  big.Int
		amountOut big.Int
		fee       big.Int
	)

	if err := poolSwapEvent.DecodeArgs(
		tx.Log,
		&initiator,
		&tokenIn,
		&tokenOut,
		&amountIn,
		&amountOut,
		&fee,
	); err != nil {
		return err
	}

	poolSwapEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          poolSwapEventName,
		Payload: map[string]any{
			"initiator": initiator.Hex(),
			"tokenIn":   tokenIn.Hex(),
			"tokenOut":  tokenOut.Hex(),
			"amountIn":  amountIn.String(),
			"amountOut": amountOut.String(),
			"fee":       fee.String(),
		},
	}

	return pubCB(ctx, poolSwapEvent)
}

func (h *poolSwapHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {
	var (
		tokenOut common.Address
		tokenIn  common.Address
		amountIn big.Int
	)

	if err := poolSwapSig.DecodeArgs(w3.B(tx.InputData), &tokenOut, &tokenIn, &amountIn); err != nil {
		return err
	}

	poolSwapEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          poolSwapEventName,
		Payload: map[string]any{
			"initiator": tx.From,
			"tokenIn":   tokenIn.Hex(),
			"tokenOut":  tokenOut.Hex(),
			"amountIn":  amountIn.String(),
			"amountOut": "0",
			"fee":       "0",
		},
	}

	return pubCB(ctx, poolSwapEvent)
}
