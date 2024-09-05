package router

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type tokenBurnHandler struct{}

const burnEventName = "TOKEN_BURN"

var (
	_ Handler = (*tokenBurnHandler)(nil)

	tokenBurnEvent = w3.MustNewEvent("Burn(address indexed _tokenBurner, uint256 _value)")
	tokenBurnToSig = w3.MustNewFunc("Burn(uint256)", "bool")
)

func (h *tokenBurnHandler) Name() string {
	return burnEventName
}

func (h *tokenBurnHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var (
		tokenBurner common.Address
		value       big.Int
	)

	if err := tokenBurnEvent.DecodeArgs(tx.Log, &tokenBurner, &value); err != nil {
		return err
	}

	tokenBurnEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          burnEventName,
		Payload: map[string]any{
			"tokenBurner": tokenBurner.Hex(),
			"value":       value.String(),
		},
	}

	return pubCB(ctx, tokenBurnEvent)
}

func (h *tokenBurnHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {
	var value big.Int

	if err := tokenBurnToSig.DecodeArgs(w3.B(tx.InputData), &value); err != nil {
		return err
	}

	tokenBurnEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          burnEventName,
		Payload: map[string]any{
			"tokenBurner": tx.From,
			"value":       value.String(),
		},
	}

	return pubCB(ctx, tokenBurnEvent)
}
