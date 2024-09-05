package router

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type poolDepositHandler struct{}

const poolDepositEventName = "POOL_DEPOSIT"

var (
	_ Handler = (*poolDepositHandler)(nil)

	poolDepositEvent = w3.MustNewEvent("Deposit(address indexed initiator, address indexed tokenIn, uint256 amountIn)")
	poolDepositSig   = w3.MustNewFunc("deposit(address, uint256)", "")
)

func (h *poolDepositHandler) Name() string {
	return poolDepositEventName
}

func (h *poolDepositHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var (
		initiator common.Address
		tokenIn   common.Address
		amountIn  big.Int
	)

	if err := poolDepositEvent.DecodeArgs(
		tx.Log,
		&initiator,
		&tokenIn,
		&amountIn,
	); err != nil {
		return err
	}

	poolDepositEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          poolDepositEventName,
		Payload: map[string]any{
			"initiator": initiator.Hex(),
			"tokenIn":   tokenIn.Hex(),
			"amountIn":  amountIn.String(),
		},
	}

	return pubCB(ctx, poolDepositEvent)
}

func (h *poolDepositHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {
	var (
		tokenIn  common.Address
		amountIn big.Int
	)

	if err := poolDepositSig.DecodeArgs(w3.B(tx.InputData), &tokenIn, &amountIn); err != nil {
		return err
	}

	poolDepositEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          poolDepositEventName,
		Payload: map[string]any{
			"initiator": tx.From,
			"tokenIn":   tokenIn.Hex(),
			"amountIn":  amountIn.String(),
		},
	}

	return pubCB(ctx, poolDepositEvent)
}
