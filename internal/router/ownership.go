package router

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type ownershipHandler struct{}

const (
	ownershipEventName = "OWNERSHIP_TRANSFERRED"
)

var (
	_ Handler = (*ownershipHandler)(nil)

	ownershipEvent = w3.MustNewEvent("OwnershipTransferred(address indexed previousOwner, address indexed newOwner)")
	ownershipToSig = w3.MustNewFunc("transferOwnership(address)", "bool")
)

func (h *ownershipHandler) Name() string {
	return ownershipEventName
}

func (h *ownershipHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var (
		previousOwner common.Address
		newOwner      common.Address
	)

	if err := ownershipEvent.DecodeArgs(tx.Log, &previousOwner, &newOwner); err != nil {
		return err
	}

	ownershipEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          ownershipEventName,
		Payload: map[string]any{
			"previousOwner": previousOwner.Hex(),
			"newOwner":      newOwner.Hex(),
		},
	}

	return pubCB(ctx, ownershipEvent)
}

func (h *ownershipHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {

	var newOwner common.Address

	if err := ownershipToSig.DecodeArgs(w3.B(tx.InputData), &newOwner); err != nil {
		return err
	}

	ownershipEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          ownershipEventName,
		Payload: map[string]any{
			"previousOwner": tx.From,
			"newOwner":      newOwner.Hex(),
		},
	}

	return pubCB(ctx, ownershipEvent)
}
