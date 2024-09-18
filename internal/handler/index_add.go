package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celo-tracker/pkg/router"
	"github.com/grassrootseconomics/w3-celo"
)

const indexAddEventName = "INDEX_ADD"

var (
	indexAddEvent    = w3.MustNewEvent("AddressAdded(address _token)")
	indexAddSig      = w3.MustNewFunc("add(address)", "bool")
	indexRegisterSig = w3.MustNewFunc("register(address)", "bool")
)

func HandleIndexAddLog(hc *HandlerContainer) router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var address common.Address

		if err := indexAddEvent.DecodeArgs(lp.Log, &address); err != nil {
			return err
		}

		indexAddEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          indexAddEventName,
			Payload: map[string]any{
				"address": address.Hex(),
			},
		}

		if err := hc.cache.Add(ctx, address.Hex()); err != nil {
			return err
		}

		return c(ctx, indexAddEvent)
	}
}

func HandleIndexAddInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		indexAddEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          indexAddEventName,
		}

		switch idp.InputData[:8] {
		case "0a3b0a4f":
			var address common.Address

			indexAddEvent.Payload = map[string]any{
				"address": address.Hex(),
			}

			if err := indexAddSig.DecodeArgs(w3.B(idp.InputData), &address); err != nil {
				return err
			}

			return c(ctx, indexAddEvent)
		case "4420e486":
			var address common.Address

			indexAddEvent.Payload = map[string]any{
				"address": address.Hex(),
			}

			if err := indexRegisterSig.DecodeArgs(w3.B(idp.InputData), &address); err != nil {
				return err
			}

			return c(ctx, indexAddEvent)
		}

		return nil
	}
}
