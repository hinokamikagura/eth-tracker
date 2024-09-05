package router

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celoutils/v3"
	"github.com/grassrootseconomics/w3-celo"
)

type faucetGiveHandler struct{}

const faucetGiveEventName = "FAUCET_GIVE"

var (
	_ Handler = (*faucetGiveHandler)(nil)

	faucetGiveEvent = w3.MustNewEvent("Give(address indexed _recipient, address indexed _token, uint256 _amount)")
	faucetGiveToSig = w3.MustNewFunc("giveTo(address)", "uint256")
	faucetGimmeSig  = w3.MustNewFunc("gimme()", "uint256")
)

func (h *faucetGiveHandler) Name() string {
	return faucetGiveEventName
}

func (h *faucetGiveHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var (
		recipient common.Address
		token     common.Address
		amount    big.Int
	)

	if err := faucetGiveEvent.DecodeArgs(tx.Log, &recipient, &token, &amount); err != nil {
		return err
	}

	faucetGiveEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          faucetGiveEventName,
		Payload: map[string]any{
			"recipient": recipient.Hex(),
			"token":     token.Hex(),
			"amount":    amount.String(),
		},
	}

	return pubCB(ctx, faucetGiveEvent)
}

func (h *faucetGiveHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {
	faucetGiveEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          faucetGiveEventName,
	}

	switch tx.InputData[:8] {
	case "63e4bff4":
		var to common.Address

		if err := faucetGiveToSig.DecodeArgs(w3.B(tx.InputData), &to); err != nil {
			return err
		}

		faucetGiveEvent.Payload = map[string]any{
			"recipient": to.Hex(),
			"token":     celoutils.ZeroAddress,
			"amount":    "0",
		}

		return pubCB(ctx, faucetGiveEvent)
	case "de82efb4":
		faucetGiveEvent.Payload = map[string]any{
			"recipient": celoutils.ZeroAddress,
			"token":     celoutils.ZeroAddress,
			"amount":    "0",
		}

		return pubCB(ctx, faucetGiveEvent)
	}

	return nil
}
