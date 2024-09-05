package router

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celoutils/v3"
	"github.com/grassrootseconomics/w3-celo"
)

type tokenTransferHandler struct {
	cache cache.Cache
}

const transferEventName = "TOKEN_TRANSFER"

var (
	_ Handler = (*tokenTransferHandler)(nil)

	tokenTransferEvent   = w3.MustNewEvent("Transfer(address indexed _from, address indexed _to, uint256 _value)")
	tokenTransferSig     = w3.MustNewFunc("transfer(address, uint256)", "bool")
	tokenTransferFromSig = w3.MustNewFunc("transferFrom(address, address, uint256)", "bool")

	stables = map[string]bool{
		celoutils.CUSDContractMainnet: true,
		celoutils.CKESContractMainnet: true,
		celoutils.USDTContractMainnet: true,
		celoutils.USDCContractMainnet: true,
	}
)

func (h *tokenTransferHandler) Name() string {
	return transferEventName
}

func (h *tokenTransferHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var (
		from  common.Address
		to    common.Address
		value big.Int
	)

	if err := tokenTransferEvent.DecodeArgs(tx.Log, &from, &to, &value); err != nil {
		return err
	}

	proceed, err := h.checkStables(ctx, from.Hex(), to.Hex(), tx.Log.Address.Hex())
	if err != nil {
		return err
	}
	if !proceed {
		return nil
	}

	tokenTransferEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          transferEventName,
		Payload: map[string]any{
			"from":  from.Hex(),
			"to":    to.Hex(),
			"value": value.String(),
		},
	}

	return pubCB(ctx, tokenTransferEvent)
}

func (h *tokenTransferHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {
	tokenTransferEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          transferEventName,
	}

	switch tx.InputData[:8] {
	case "a9059cbb":
		var (
			to    common.Address
			value big.Int
		)

		if err := tokenTransferSig.DecodeArgs(w3.B(tx.InputData), &to, &value); err != nil {
			return err
		}

		proceed, err := h.checkStables(ctx, tx.From, to.Hex(), tx.ContractAddress)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		tokenTransferEvent.Payload = map[string]any{
			"from":  tx.From,
			"to":    to.Hex(),
			"value": value.String(),
		}

		return pubCB(ctx, tokenTransferEvent)
	case "23b872dd":
		var (
			from  common.Address
			to    common.Address
			value big.Int
		)

		if err := tokenTransferFromSig.DecodeArgs(w3.B(tx.InputData), &from, &to, &value); err != nil {
			return err
		}

		proceed, err := h.checkStables(ctx, from.Hex(), to.Hex(), tx.ContractAddress)
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

		return pubCB(ctx, tokenTransferEvent)
	}

	return nil
}

func (h *tokenTransferHandler) checkStables(ctx context.Context, from string, to string, contractAddress string) (bool, error) {
	_, ok := stables[contractAddress]
	if !ok {
		return true, nil
	}

	// TODO: Pipeline this check on Redis with a new method
	fromExists, err := h.cache.Exists(ctx, from)
	if err != nil {
		return false, err
	}

	toExists, err := h.cache.Exists(ctx, to)
	if err != nil {
		return false, err
	}

	return fromExists || toExists, nil
}
