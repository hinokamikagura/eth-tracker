package router

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type tokenMintHandler struct{}

const mintEventName = "TOKEN_MINT"

var (
	_ Handler = (*tokenMintHandler)(nil)

	tokenMintEvent = w3.MustNewEvent("Mint(address indexed _tokenMinter, address indexed _beneficiary, uint256 _value)")
	tokenMintToSig = w3.MustNewFunc("mintTo(address, uint256)", "bool")
)

func (h *tokenMintHandler) Name() string {
	return mintEventName
}

func (h *tokenMintHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var (
		tokenMinter common.Address
		to          common.Address
		value       big.Int
	)

	if err := tokenMintEvent.DecodeArgs(tx.Log, &tokenMinter, &to, &value); err != nil {
		return err
	}

	tokenMintEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          mintEventName,
		Payload: map[string]any{
			"tokenMinter": tokenMinter.Hex(),
			"to":          to.Hex(),
			"value":       value.String(),
		},
	}

	return pubCB(ctx, tokenMintEvent)
}

func (h *tokenMintHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {

	var (
		to    common.Address
		value big.Int
	)

	if err := tokenMintToSig.DecodeArgs(w3.B(tx.InputData), &to, &value); err != nil {
		return err
	}

	tokenMintEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          mintEventName,
		Payload: map[string]any{
			"tokenMinter": tx.From,
			"to":          to.Hex(),
			"value":       value.String(),
		},
	}

	return pubCB(ctx, tokenMintEvent)
}
