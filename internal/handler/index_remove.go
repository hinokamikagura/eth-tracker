package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celo-tracker/pkg/router"
	"github.com/grassrootseconomics/w3-celo"
)

const indexRemoveEventName = "INDEX_REMOVE"

var (
	indexRemoveEvent = w3.MustNewEvent("AddressRemoved(address _token)")
	indexRemoveSig   = w3.MustNewFunc("remove(address)", "bool")
)

func HandleIndexRemoveLog(hc *HandlerContainer) router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var address common.Address

		if err := indexRemoveEvent.DecodeArgs(lp.Log, &address); err != nil {
			return err
		}

		indexRemoveEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          indexRemoveEventName,
			Payload: map[string]any{
				"address": address.Hex(),
			},
		}

		if err := hc.cache.Remove(ctx, address.Hex()); err != nil {
			return err
		}

		return c(ctx, indexRemoveEvent)
	}
}

func HandleIndexRemoveInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var address common.Address

		if err := indexRemoveSig.DecodeArgs(w3.B(idp.InputData), &address); err != nil {
			return err
		}

		indexRemoveEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          indexRemoveEventName,
			Payload: map[string]any{
				"address": address.Hex(),
			},
		}

		return c(ctx, indexRemoveEvent)
	}
}
