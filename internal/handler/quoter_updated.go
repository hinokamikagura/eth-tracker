package handler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
	"github.com/lmittmann/w3"
)

const quoterUpdatedEventName = "QUOTER_UPDATED"

var (
	quoterUpdatedEvent = w3.MustNewEvent("QuoterUpdated(address indexed newQuoter)")
	quoterUpdatedSig   = w3.MustNewFunc("setQuoter(address)", "")
)

func HandleQuoterUpdatedLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var newQuoter common.Address

		if err := quoterUpdatedEvent.DecodeArgs(lp.Log, &newQuoter); err != nil {
			return err
		}

		quoterUpdatedEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          quoterUpdatedEventName,
			Payload: map[string]any{
				"newQuoter": newQuoter.Hex(),
			},
		}

		return c(ctx, quoterUpdatedEvent)
	}
}

func HandleQuoterUpdatedInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var newQuoter common.Address

		if err := quoterUpdatedSig.DecodeArgs(w3.B(idp.InputData), &newQuoter); err != nil {
			return err
		}

		quoterUpdatedEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          quoterUpdatedEventName,
			Payload: map[string]any{
				"newQuoter": newQuoter.Hex(),
			},
		}

		return c(ctx, quoterUpdatedEvent)
	}
}
