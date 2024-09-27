package handler

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
	"github.com/lmittmann/w3"
)

const poolDepositEventName = "POOL_DEPOSIT"

var (
	poolDepositEvent = w3.MustNewEvent("Deposit(address indexed initiator, address indexed tokenIn, uint256 amountIn)")
	poolDepositSig   = w3.MustNewFunc("deposit(address, uint256)", "")
)

func HandlePoolDepositLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			initiator common.Address
			tokenIn   common.Address
			amountIn  big.Int
		)

		if err := poolDepositEvent.DecodeArgs(
			lp.Log,
			&initiator,
			&tokenIn,
			&amountIn,
		); err != nil {
			return err
		}

		poolDepositEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          poolDepositEventName,
			Payload: map[string]any{
				"initiator": initiator.Hex(),
				"tokenIn":   tokenIn.Hex(),
				"amountIn":  amountIn.String(),
			},
		}

		return c(ctx, poolDepositEvent)
	}
}

func HandlePoolDepositInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var (
			tokenIn  common.Address
			amountIn big.Int
		)

		if err := poolDepositSig.DecodeArgs(w3.B(idp.InputData), &tokenIn, &amountIn); err != nil {
			return err
		}

		poolDepositEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          poolDepositEventName,
			Payload: map[string]any{
				"initiator": idp.From,
				"tokenIn":   tokenIn.Hex(),
				"amountIn":  amountIn.String(),
			},
		}

		return c(ctx, poolDepositEvent)
	}
}
