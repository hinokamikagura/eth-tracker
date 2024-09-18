package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celo-tracker/pkg/router"
	"github.com/grassrootseconomics/w3-celo"
)

const burnEventName = "TOKEN_BURN"

var (
	tokenBurnEvent = w3.MustNewEvent("Burn(address indexed _tokenBurner, uint256 _value)")
	tokenBurnToSig = w3.MustNewFunc("burn(uint256)", "bool")
)

func HandleTokenBurnLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			tokenBurner common.Address
			value       big.Int
		)

		if err := tokenBurnEvent.DecodeArgs(lp.Log, &tokenBurner, &value); err != nil {
			return err
		}

		tokenBurnEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          burnEventName,
			Payload: map[string]any{
				"tokenBurner": tokenBurner.Hex(),
				"value":       value.String(),
			},
		}

		return c(ctx, tokenBurnEvent)
	}
}

func HandleTokenBurnInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var value big.Int

		if err := tokenBurnToSig.DecodeArgs(w3.B(idp.InputData), &value); err != nil {
			return err
		}

		tokenBurnEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          burnEventName,
			Payload: map[string]any{
				"tokenBurner": idp.From,
				"value":       value.String(),
			},
		}

		return c(ctx, tokenBurnEvent)
	}
}
