package router

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type quoterPriceHandler struct{}

const quoterPriceEventName = "QUOTER_PRICE_INDEX_UPDATED"

var (
	_ Handler = (*quoterPriceHandler)(nil)

	quoterPriceEvent = w3.MustNewEvent("PriceIndexUpdated(address _tokenAddress, uint256 _exchangeRate)")
	quoterPriceToSig = w3.MustNewFunc("setPriceIndexValue(address, uint256)", "uint256")
)

func (h *quoterPriceHandler) Name() string {
	return quoterPriceEventName
}

func (h *quoterPriceHandler) SuccessTx(ctx context.Context, tx SuccessTx, pubCB PubCallback) error {
	var (
		token        common.Address
		exchangeRate big.Int
	)

	if err := quoterPriceEvent.DecodeArgs(tx.Log, &token, &exchangeRate); err != nil {
		return err
	}

	quoterPriceEvent := event.Event{
		Index:           tx.Log.Index,
		Block:           tx.Log.BlockNumber,
		ContractAddress: tx.Log.Address.Hex(),
		Success:         true,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.Log.TxHash.Hex(),
		TxType:          quoterPriceEventName,
		Payload: map[string]any{
			"token":        token.Hex(),
			"exchangeRate": exchangeRate.String(),
		},
	}

	return pubCB(ctx, quoterPriceEvent)
}

func (h *quoterPriceHandler) RevertTx(ctx context.Context, tx RevertTx, pubCB PubCallback) error {
	var (
		token        common.Address
		exchangeRate big.Int
	)

	if err := quoterPriceToSig.DecodeArgs(w3.B(tx.InputData), &token, &exchangeRate); err != nil {
		return err
	}

	quoterPriceEvent := event.Event{
		Block:           tx.Block,
		ContractAddress: tx.ContractAddress,
		Success:         false,
		Timestamp:       tx.Timestamp,
		TxHash:          tx.TxHash,
		TxType:          quoterPriceEventName,
		Payload: map[string]any{
			"token":        token.Hex(),
			"exchangeRate": exchangeRate.String(),
		},
	}

	return pubCB(ctx, quoterPriceEvent)
}
