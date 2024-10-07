package handler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
	"github.com/lmittmann/w3"
)

const custodialRegistrationEventName = "CUSTODIAL_REGISTRATION"

var (
	custodialRegistrationEvent = w3.MustNewEvent("NewRegistration(address indexed subject)")
	custodialRegistrationSig   = w3.MustNewFunc("register(address)", "")
)

func HandleCustodialRegistrationLog() router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var account common.Address

		if err := custodialRegistrationEvent.DecodeArgs(lp.Log, &account); err != nil {
			return err
		}

		custodialRegistrationEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          custodialRegistrationEventName,
			Payload: map[string]any{
				"account": account.Hex(),
			},
		}

		return c(ctx, custodialRegistrationEvent)
	}
}

func HandleCustodialRegistrationInputData() router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var account common.Address

		if err := custodialRegistrationSig.DecodeArgs(w3.B(idp.InputData), &account); err != nil {
			return err
		}

		custodialRegistrationEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          custodialRegistrationEventName,
			Payload: map[string]any{
				"account": account.Hex(),
			},
		}

		return c(ctx, custodialRegistrationEvent)
	}
}
