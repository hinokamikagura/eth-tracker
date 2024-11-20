package handler

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
	"github.com/lmittmann/w3"
)

const transferEventName = "TOKEN_TRANSFER"

var (
	tokenTransferEvent   = w3.MustNewEvent("Transfer(address indexed _from, address indexed _to, uint256 _value)")
	tokenTransferSig     = w3.MustNewFunc("transfer(address, uint256)", "bool")
	tokenTransferFromSig = w3.MustNewFunc("transferFrom(address, address, uint256)", "bool")
)

func HandleTokenTransferLog(hc *HandlerContainer) router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			from  common.Address
			to    common.Address
			value big.Int
		)

		if err := tokenTransferEvent.DecodeArgs(lp.Log, &from, &to, &value); err != nil {
			return err
		}

		proceed, err := hc.checkWithinNetwork(ctx, lp.Log.Address.Hex(), from.Hex(), to.Hex())
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		tokenTransferEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          transferEventName,
			Payload: map[string]any{
				"from":  from.Hex(),
				"to":    to.Hex(),
				"value": value.String(),
			},
		}

		return c(ctx, tokenTransferEvent)
	}
}

func HandleTokenTransferInputData(hc *HandlerContainer) router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		tokenTransferEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          transferEventName,
		}

		switch idp.InputData[:8] {
		case "a9059cbb":
			var (
				to    common.Address
				value big.Int
			)

			if err := tokenTransferSig.DecodeArgs(w3.B(idp.InputData), &to, &value); err != nil {
				return err
			}

			proceed, err := hc.checkWithinNetwork(ctx, idp.ContractAddress, idp.From, to.Hex())
			if err != nil {
				return err
			}
			if !proceed {
				return nil
			}

			tokenTransferEvent.Payload = map[string]any{
				"from":  idp.From,
				"to":    to.Hex(),
				"value": value.String(),
			}

			return c(ctx, tokenTransferEvent)
		case "23b872dd":
			var (
				from  common.Address
				to    common.Address
				value big.Int
			)

			if err := tokenTransferFromSig.DecodeArgs(w3.B(idp.InputData), &from, &to, &value); err != nil {
				return err
			}

			proceed, err := hc.checkWithinNetwork(ctx, idp.ContractAddress, from.Hex(), to.Hex())
			if err != nil {
				return err
			}
			if !proceed {
				return nil
			}

			tokenTransferEvent.Payload = map[string]any{
				"from":  from.Hex(),
				"to":    to.Hex(),
				"value": value.String(),
			}

			return c(ctx, tokenTransferEvent)
		}

		return nil
	}
}
