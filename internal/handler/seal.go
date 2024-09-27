package handler

import (
	"context"
	"math/big"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
	"github.com/lmittmann/w3"
)

const sealEventName = "SEAL_STATE_CHANGE"

var (
	sealEvent = w3.MustNewEvent("SealStateChange(bool indexed _final, uint256 _sealState)")
	sealToSig = w3.MustNewFunc("seal(uint256)", "uint256")
)

func HandleSealStateChangeLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			final     bool
			sealState big.Int
		)

		if err := sealEvent.DecodeArgs(lp.Log, &final, &sealState); err != nil {
			return err
		}

		sealEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          sealEventName,
			Payload: map[string]any{
				"final":     final,
				"sealState": sealState.String(),
			},
		}

		return c(ctx, sealEvent)
	}
}

func HandleSealStateChangeInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var sealState big.Int

		if err := sealToSig.DecodeArgs(w3.B(idp.InputData), &sealState); err != nil {
			return err
		}

		sealEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          sealEventName,
			Payload: map[string]any{
				"sealState": sealState.String(),
			},
		}

		return c(ctx, sealEvent)
	}
}
