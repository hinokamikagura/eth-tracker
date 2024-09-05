package router

import (
	"context"
	"math/big"

	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type sealHandler struct{}

const sealEventName = "SEAL_STATE_CHANGE"

var (
	_ Handler = (*sealHandler)(nil)

	sealEvent = w3.MustNewEvent("SealStateChange(bool indexed _final, uint256 _sealState)")
	sealToSig = w3.MustNewFunc("seal(uint256)", "uint256")
)

func (h *sealHandler) Name() string {
	return sealEventName
}

func (h *sealHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var (
		final     bool
		sealState big.Int
	)

	if err := sealEvent.DecodeArgs(tx.Log, &final, &sealState); err != nil {
		return err
	}

	sealEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          sealEventName,
		Payload: map[string]any{
			"final":     final,
			"sealState": sealState.String(),
		},
	}

	return pubCB(ctx, sealEvent)
}

func (h *sealHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {
	var sealState big.Int

	if err := sealToSig.DecodeArgs(w3.B(tx.InputData), &sealState); err != nil {
		return err
	}

	sealEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          sealEventName,
		Payload: map[string]any{
			"sealState": sealState.String(),
		},
	}

	return pubCB(ctx, sealEvent)
}
