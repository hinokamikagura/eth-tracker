package router

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type indexAddHandler struct {
	cache cache.Cache
}

const indexAddEventName = "INDEX_ADD"

var (
	_ Handler = (*indexAddHandler)(nil)

	indexAddEvent    = w3.MustNewEvent("AddressAdded(address _token)")
	indexAddSig      = w3.MustNewFunc("add(address)", "bool")
	indexRegisterSig = w3.MustNewFunc("register(address)", "bool")
)

func (h *indexAddHandler) Name() string {
	return indexAddEventName
}

func (h *indexAddHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var address common.Address

	if err := indexAddEvent.DecodeArgs(tx.Log, &address); err != nil {
		return err
	}

	indexAddEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          indexAddEventName,
		Payload: map[string]any{
			"address": address.Hex(),
		},
	}

	if err := h.cache.Add(ctx, address.Hex()); err != nil {
		return err
	}

	return pubCB(ctx, indexAddEvent)
}

func (h *indexAddHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {
	indexAddEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          indexAddEventName,
	}

	switch tx.InputData[:8] {
	case "0a3b0a4f":
		var address common.Address

		indexAddEvent.Payload = map[string]any{
			"address": address.Hex(),
		}

		if err := indexAddSig.DecodeArgs(w3.B(tx.InputData), &address); err != nil {
			return err
		}

		return pubCB(ctx, indexAddEvent)
	case "4420e486":
		var address common.Address

		indexAddEvent.Payload = map[string]any{
			"address": address.Hex(),
		}

		if err := indexRegisterSig.DecodeArgs(w3.B(tx.InputData), &address); err != nil {
			return err
		}

		return pubCB(ctx, indexAddEvent)
	}

	return nil
}
