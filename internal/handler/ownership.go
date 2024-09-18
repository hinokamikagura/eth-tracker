package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celo-tracker/pkg/router"
	"github.com/grassrootseconomics/w3-celo"
)

const ownershipEventName = "OWNERSHIP_TRANSFERRED"

var (
	ownershipEvent = w3.MustNewEvent("OwnershipTransferred(address indexed previousOwner, address indexed newOwner)")
	ownershipToSig = w3.MustNewFunc("transferOwnership(address)", "bool")
)

func HandleOwnershipLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			previousOwner common.Address
			newOwner      common.Address
		)

		if err := ownershipEvent.DecodeArgs(lp.Log, &previousOwner, &newOwner); err != nil {
			return err
		}

		ownershipEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          ownershipEventName,
			Payload: map[string]any{
				"previousOwner": previousOwner.Hex(),
				"newOwner":      newOwner.Hex(),
			},
		}

		return c(ctx, ownershipEvent)
	}
}

func HandleOwnershipInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var newOwner common.Address

		if err := ownershipToSig.DecodeArgs(w3.B(idp.InputData), &newOwner); err != nil {
			return err
		}

		ownershipEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          ownershipEventName,
			Payload: map[string]any{
				"previousOwner": idp.From,
				"newOwner":      newOwner.Hex(),
			},
		}

		return c(ctx, ownershipEvent)
	}
}
