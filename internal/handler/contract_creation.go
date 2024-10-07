package handler

import (
	"context"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
)

const contractCreationEventName = "CONTRACT_CREATION"

func HandleContractCreation() router.ContractCreationHandlerFunc {
	return func(ctx context.Context, ccp router.ContractCreationPayload, c router.Callback) error {
		contractCreationEvent := event.Event{
			Block:           ccp.Block,
			ContractAddress: ccp.ContractAddress,
			Success:         ccp.Success,
			Timestamp:       ccp.Timestamp,
			TxHash:          ccp.TxHash,
			TxType:          contractCreationEventName,
			Payload: map[string]any{
				"from": ccp.From,
			},
		}

		return c(ctx, contractCreationEvent)
	}
}
