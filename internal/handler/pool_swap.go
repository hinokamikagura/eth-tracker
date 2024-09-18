package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celo-tracker/pkg/router"
	"github.com/grassrootseconomics/w3-celo"
)

const poolSwapEventName = "POOL_SWAP"

var (
	poolSwapEvent = w3.MustNewEvent("Swap(address indexed initiator, address indexed tokenIn, address tokenOut, uint256 amountIn, uint256 amountOut, uint256 fee)")
	poolSwapSig   = w3.MustNewFunc("withdraw(address, address, uint256)", "")
)

func HandlePoolSwapLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			initiator common.Address
			tokenIn   common.Address
			tokenOut  common.Address
			amountIn  big.Int
			amountOut big.Int
			fee       big.Int
		)

		if err := poolSwapEvent.DecodeArgs(
			lp.Log,
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
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
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

		return c(ctx, poolSwapEvent)
	}
}

func HandlePoolSwapInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var (
			tokenOut common.Address
			tokenIn  common.Address
			amountIn big.Int
		)

		if err := poolSwapSig.DecodeArgs(w3.B(idp.InputData), &tokenOut, &tokenIn, &amountIn); err != nil {
			return err
		}

		poolSwapEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          poolSwapEventName,
			Payload: map[string]any{
				"initiator": idp.From,
				"tokenIn":   tokenIn.Hex(),
				"tokenOut":  tokenOut.Hex(),
				"amountIn":  amountIn.String(),
				"amountOut": "0",
				"fee":       "0",
			},
		}

		return c(ctx, poolSwapEvent)
	}
}
