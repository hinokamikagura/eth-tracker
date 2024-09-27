package handler

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
	"github.com/lmittmann/w3"
)

const mintEventName = "TOKEN_MINT"

var (
	tokenMintEvent = w3.MustNewEvent("Mint(address indexed _tokenMinter, address indexed _beneficiary, uint256 _value)")
	tokenMintToSig = w3.MustNewFunc("mintTo(address, uint256)", "bool")
)

func HandleTokenMintLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			tokenMinter common.Address
			to          common.Address
			value       big.Int
		)

		if err := tokenMintEvent.DecodeArgs(lp.Log, &tokenMinter, &to, &value); err != nil {
			return err
		}

		tokenMintEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          mintEventName,
			Payload: map[string]any{
				"tokenMinter": tokenMinter.Hex(),
				"to":          to.Hex(),
				"value":       value.String(),
			},
		}

		return c(ctx, tokenMintEvent)
	}
}

func HandleTokenMintInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var (
			to    common.Address
			value big.Int
		)

		if err := tokenMintToSig.DecodeArgs(w3.B(idp.InputData), &to, &value); err != nil {
			return err
		}

		tokenMintEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          mintEventName,
			Payload: map[string]any{
				"tokenMinter": idp.From,
				"to":          to.Hex(),
				"value":       value.String(),
			},
		}

		return c(ctx, tokenMintEvent)
	}
}
