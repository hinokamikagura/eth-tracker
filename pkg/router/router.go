// Package router provides event routing functionality for blockchain events.
// It routes different types of events (logs, input data, contract creations) to
// their respective handlers based on event signatures and function selectors.
package router

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
)

const (
	// functionSelectorLength is the length of a function selector in hex characters (4 bytes = 8 hex chars)
	functionSelectorLength = 8
)

type (
	// Callback is the function type called after an event is processed by a handler.
	Callback func(context.Context, event.Event) error

	// LogPayload contains data for processing event logs.
	LogPayload struct {
		Log       *types.Log // The event log from the transaction receipt
		Timestamp uint64     // Block timestamp
	}

	// InputDataPayload contains data for processing transaction input data.
	InputDataPayload struct {
		From            string // Transaction sender address
		InputData       string // Transaction input data (hex encoded)
		Block           uint64 // Block number
		ContractAddress string // Contract address that received the transaction
		Timestamp       uint64 // Block timestamp
		TxHash          string // Transaction hash
	}

	// ContractCreationPayload contains data for processing contract creation events.
	ContractCreationPayload struct {
		From            string // Transaction sender address
		ContractAddress string // Address of the created contract
		Block           uint64 // Block number
		Timestamp       uint64 // Block timestamp
		TxHash          string // Transaction hash
		Success         bool   // Whether the contract creation was successful
	}

	// Handler function types for different event types
	LogHandlerFunc              func(context.Context, LogPayload, Callback) error
	InputDataHandlerFunc        func(context.Context, InputDataPayload, Callback) error
	ContractCreationHandlerFunc func(context.Context, ContractCreationPayload, Callback) error

	// Route entries for different event types
	LogRouteEntry struct {
		Signature   common.Hash    // Event signature (first topic)
		HandlerFunc LogHandlerFunc // Handler function for this event
	}

	InputDataEntry struct {
		Signature   string               // Function selector (first 4 bytes)
		HandlerFunc InputDataHandlerFunc // Handler function for this function
	}

	// Router routes blockchain events to their respective handlers.
	Router struct {
		callbackFn              Callback
		logHandlers             map[common.Hash]LogRouteEntry
		inputDataHandlers       map[string]InputDataEntry
		contractCreationHandler ContractCreationHandlerFunc
	}
)

// New creates a new Router instance with the provided callback function.
// The callback is invoked after each event is processed by its handler.
func New(callbackFn Callback) *Router {
	return &Router{
		callbackFn:              callbackFn,
		logHandlers:             make(map[common.Hash]LogRouteEntry),
		inputDataHandlers:       make(map[string]InputDataEntry),
		contractCreationHandler: nil,
	}
}

// RegisterLogRoute registers a handler for a specific event log signature.
// The signature should be the keccak256 hash of the event signature (first topic).
func (r *Router) RegisterLogRoute(signature common.Hash, handlerFunc LogHandlerFunc) {
	r.logHandlers[signature] = LogRouteEntry{
		Signature:   signature,
		HandlerFunc: handlerFunc,
	}
}

// RegisterInputDataRoute registers a handler for a specific function selector.
// The signature should be the first 4 bytes (8 hex characters) of the function selector.
func (r *Router) RegisterInputDataRoute(signature string, handlerFunc InputDataHandlerFunc) {
	r.inputDataHandlers[signature] = InputDataEntry{
		Signature:   signature,
		HandlerFunc: handlerFunc,
	}
}

// RegisterContractCreationHandler registers a handler for contract creation events.
func (r *Router) RegisterContractCreationHandler(handlerFunc ContractCreationHandlerFunc) {
	r.contractCreationHandler = handlerFunc
}

// ProcessLog routes an event log to its registered handler based on the first topic (event signature).
// Returns nil if no handler is registered for the event signature.
func (r *Router) ProcessLog(ctx context.Context, payload LogPayload) error {
	if len(payload.Log.Topics) == 0 {
		return nil
	}

	handler, ok := r.logHandlers[payload.Log.Topics[0]]
	if !ok {
		return nil
	}

	return handler.HandlerFunc(ctx, payload, r.callbackFn)
}

// ProcessInputData routes transaction input data to its registered handler based on the function selector.
// Returns nil if the input data is too short or no handler is registered for the function selector.
func (r *Router) ProcessInputData(ctx context.Context, payload InputDataPayload) error {
	if len(payload.InputData) < functionSelectorLength {
		return nil
	}

	selector := payload.InputData[:functionSelectorLength]
	handler, ok := r.inputDataHandlers[selector]
	if !ok {
		return nil
	}

	return handler.HandlerFunc(ctx, payload, r.callbackFn)
}

// ProcessContractCreation routes a contract creation event to its registered handler.
// Returns an error if no handler is registered.
func (r *Router) ProcessContractCreation(ctx context.Context, payload ContractCreationPayload) error {
	if r.contractCreationHandler == nil {
		return nil
	}

	return r.contractCreationHandler(ctx, payload, r.callbackFn)
}
