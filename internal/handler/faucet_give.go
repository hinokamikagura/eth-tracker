package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celo-tracker/pkg/router"
	"github.com/grassrootseconomics/celoutils/v3"
	"github.com/grassrootseconomics/w3-celo"
)

const faucetGiveEventName = "FAUCET_GIVE"

var (
	faucetGiveEvent = w3.MustNewEvent("Give(address indexed _recipient, address indexed _token, uint256 _amount)")
	faucetGiveToSig = w3.MustNewFunc("giveTo(address)", "uint256")
	faucetGimmeSig  = w3.MustNewFunc("gimme()", "uint256")
)

func HandleFaucetGiveLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			recipient common.Address
			token     common.Address
			amount    big.Int
		)

		if err := faucetGiveEvent.DecodeArgs(lp.Log, &recipient, &token, &amount); err != nil {
			return err
		}

		faucetGiveEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          faucetGiveEventName,
			Payload: map[string]any{
				"recipient": recipient.Hex(),
				"token":     token.Hex(),
				"amount":    amount.String(),
			},
		}

		return c(ctx, faucetGiveEvent)
	}
}

func HandleFaucetGiveInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		faucetGiveEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          faucetGiveEventName,
		}

		switch idp.InputData[:8] {
		case "63e4bff4":
			var to common.Address

			if err := faucetGiveToSig.DecodeArgs(w3.B(idp.InputData), &to); err != nil {
				return err
			}

			faucetGiveEvent.Payload = map[string]any{
				"recipient": to.Hex(),
				"token":     celoutils.ZeroAddress,
				"amount":    "0",
			}

			return c(ctx, faucetGiveEvent)
		case "de82efb4":
			faucetGiveEvent.Payload = map[string]any{
				"recipient": celoutils.ZeroAddress,
				"token":     celoutils.ZeroAddress,
				"amount":    "0",
			}

			return c(ctx, faucetGiveEvent)
		}

		return nil
	}
}
