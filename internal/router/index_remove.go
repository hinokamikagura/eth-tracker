package router

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type indexRemoveHandler struct {
	cache cache.Cache
}

const indexRemoveEventName = "INDEX_REMOVE"

var (
	_ Handler = (*indexRemoveHandler)(nil)

	indexRemoveEvent = w3.MustNewEvent("AddressRemoved(address _token)")
	indexRemoveSig   = w3.MustNewFunc("remove(address)", "bool")
)

func (h *indexRemoveHandler) Name() string {
	return indexRemoveEventName
}

func (h *indexRemoveHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var address common.Address

	if err := indexRemoveEvent.DecodeArgs(tx.Log, &address); err != nil {
		return err
	}

	indexRemoveEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          indexRemoveEventName,
		Payload: map[string]any{
			"address": address.Hex(),
		},
	}

	if err := h.cache.Remove(ctx, address.Hex()); err != nil {
		return err
	}

	return pubCB(ctx, indexRemoveEvent)

}

func (h *indexRemoveHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {
	var address common.Address

	if err := indexRemoveSig.DecodeArgs(w3.B(tx.InputData), &address); err != nil {
		return err
	}

	indexRemoveEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          indexRemoveEventName,
		Payload: map[string]any{
			"address": address.Hex(),
		},
	}

	return pubCB(ctx, indexRemoveEvent)
}
