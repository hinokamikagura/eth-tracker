package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celo-tracker/pkg/router"
	"github.com/grassrootseconomics/w3-celo"
)

const quoterPriceEventName = "QUOTER_PRICE_INDEX_UPDATED"

var (
	quoterPriceEvent = w3.MustNewEvent("PriceIndexUpdated(address _tokenAddress, uint256 _exchangeRate)")
	quoterPriceToSig = w3.MustNewFunc("setPriceIndexValue(address, uint256)", "uint256")
)

func HandleQuoterPriceUpdateLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			token        common.Address
			exchangeRate big.Int
		)

		if err := quoterPriceEvent.DecodeArgs(lp.Log, &token, &exchangeRate); err != nil {
			return err
		}

		quoterPriceEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          quoterPriceEventName,
			Payload: map[string]any{
				"token":        token.Hex(),
				"exchangeRate": exchangeRate.String(),
			},
		}

		return c(ctx, quoterPriceEvent)
	}
}

func HandleQuoterPriceUpdateInputdata() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var (
			token        common.Address
			exchangeRate big.Int
		)

		if err := quoterPriceToSig.DecodeArgs(w3.B(idp.InputData), &token, &exchangeRate); err != nil {
			return err
		}

		quoterPriceEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          quoterPriceEventName,
			Payload: map[string]any{
				"token":        token.Hex(),
				"exchangeRate": exchangeRate.String(),
			},
		}

		return c(ctx, quoterPriceEvent)
	}
}
