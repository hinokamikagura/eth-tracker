package router

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
)

type (
	Callback func(context.Context, event.Event) error

	LogPayload struct {
		Log       *types.Log
		Timestamp uint64
	}

	InputDataPayload struct {
		From            string
		InputData       string
		Block           uint64
		ContractAddress string
		Timestamp       uint64
		TxHash          string
	}

	LogHandlerFunc       func(context.Context, LogPayload, Callback) error
	InputDataHandlerFunc func(context.Context, InputDataPayload, Callback) error

	LogRouteEntry struct {
		Signature   common.Hash
		HandlerFunc LogHandlerFunc
	}

	InputDataEntry struct {
		Signature   string
		HandlerFunc InputDataHandlerFunc
	}

	Router struct {
		callbackFn        Callback
		logHandlers       map[common.Hash]LogRouteEntry
		inputDataHandlers map[string]InputDataEntry
	}
)

func New(callbackFn Callback) *Router {
	return &Router{
		callbackFn:        callbackFn,
		logHandlers:       make(map[common.Hash]LogRouteEntry),
		inputDataHandlers: make(map[string]InputDataEntry),
	}
}

func (r *Router) RegisterLogRoute(signature common.Hash, handlerFunc LogHandlerFunc) {
	r.logHandlers[signature] = LogRouteEntry{
		Signature:   signature,
		HandlerFunc: handlerFunc,
	}
}

func (r *Router) RegisterInputDataRoute(signature string, handlerFunc InputDataHandlerFunc) {
	r.inputDataHandlers[signature] = InputDataEntry{
		Signature:   signature,
		HandlerFunc: handlerFunc,
	}
}

func (r *Router) ProcessLog(ctx context.Context, payload LogPayload) error {
	handler, ok := r.logHandlers[payload.Log.Topics[0]]
	if ok {
		return handler.HandlerFunc(ctx, payload, r.callbackFn)
	}

	return nil
}

func (r *Router) ProcessInputData(ctx context.Context, payload InputDataPayload) error {
	if len(payload.InputData) < 8 {
		return nil
	}

	handler, ok := r.inputDataHandlers[payload.InputData[:8]]
	if ok {
		return handler.HandlerFunc(ctx, payload, r.callbackFn)
	}

	return nil
}
