// Package processor provides block processing functionality for the eth-tracker service.
// It handles fetching blocks, receipts, and transactions, then routes relevant events
// through the event router based on cached address filters.
package processor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/grassrootseconomics/eth-tracker/db"
	"github.com/grassrootseconomics/eth-tracker/internal/cache"
	"github.com/grassrootseconomics/eth-tracker/internal/chain"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
)

const (
	// txStatusSuccess indicates a successful transaction (status = 1)
	txStatusSuccess = 1
	// txStatusFailed indicates a failed/reverted transaction (status = 0)
	txStatusFailed = 0
	// minInputDataLength is the minimum length required for function selector matching (4 bytes = 8 hex chars)
	minInputDataLength = 8
)

type (
	// ProcessorOpts contains configuration options for creating a new Processor.
	ProcessorOpts struct {
		Cache  cache.Cache    // Cache for address filtering
		Chain  chain.Chain    // Chain client for blockchain data
		DB     db.DB          // Database for block state persistence
		Router *router.Router // Event router for processing events
		Logg   *slog.Logger   // Structured logger
	}

	// Processor handles block processing, transaction filtering, and event routing.
	Processor struct {
		cache  cache.Cache
		chain  chain.Chain
		db     db.DB
		router *router.Router
		logg   *slog.Logger
	}
)

// NewProcessor creates a new Processor instance with the provided options.
func NewProcessor(o ProcessorOpts) *Processor {
	return &Processor{
		cache:  o.Cache,
		chain:  o.Chain,
		db:     o.DB,
		router: o.Router,
		logg:   o.Logg,
	}
}

// ProcessBlock fetches and processes a single block, extracting relevant events
// and routing them through the event router. It handles both successful and
// reverted transactions, including contract creations and log events.
//
// The function:
//   - Fetches the block and its receipts
//   - Processes successful transactions (logs and contract creations)
//   - Processes failed transactions (reverted contract creations and input data)
//   - Filters events based on cached addresses
//   - Persists block processing state to the database
func (p *Processor) ProcessBlock(ctx context.Context, blockNumber uint64) error {
	// Fetch block data
	block, err := p.chain.GetBlock(ctx, blockNumber)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("failed to fetch block %d: %w", blockNumber, err)
	}

	// Fetch transaction receipts
	receipts, err := p.chain.GetReceipts(ctx, block.Number())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("failed to fetch receipts for block %d: %w", blockNumber, err)
	}

	blockTimestamp := block.Time()

	// Process each receipt
	for _, receipt := range receipts {
		if err := p.processReceipt(ctx, receipt, blockNumber, blockTimestamp); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			return fmt.Errorf("failed to process receipt %s in block %d: %w", receipt.TxHash.Hex(), blockNumber, err)
		}
	}

	// Mark block as processed
	if err := p.db.SetValue(blockNumber); err != nil {
		return fmt.Errorf("failed to persist block %d: %w", blockNumber, err)
	}

	p.logg.Debug("successfully processed block", "block", blockNumber, "tx_count", len(receipts))
	return nil
}

// processReceipt handles processing of a single transaction receipt.
func (p *Processor) processReceipt(ctx context.Context, receipt *types.Receipt, blockNumber, timestamp uint64) error {
	switch receipt.Status {
	case txStatusSuccess:
		return p.processSuccessfulReceipt(ctx, receipt, blockNumber, timestamp)
	case txStatusFailed:
		return p.processFailedReceipt(ctx, receipt, blockNumber, timestamp)
	default:
		// Unknown status, skip
		return nil
	}
}

// processSuccessfulReceipt processes a successful transaction receipt.
func (p *Processor) processSuccessfulReceipt(ctx context.Context, receipt *types.Receipt, blockNumber, timestamp uint64) error {
	// Process event logs
	for _, log := range receipt.Logs {
		exists, err := p.cache.Exists(ctx, log.Address.Hex())
		if err != nil {
			return fmt.Errorf("cache lookup failed for log address %s: %w", log.Address.Hex(), err)
		}
		if !exists {
			continue
		}

		if err := p.router.ProcessLog(ctx, router.LogPayload{
			Log:       log,
			Timestamp: timestamp,
		}); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			return fmt.Errorf("failed to process log for tx %s: %w", receipt.TxHash.Hex(), err)
		}
	}

	// Process contract creation if applicable
	if receipt.ContractAddress != (common.Address{}) {
		return p.processContractCreation(ctx, receipt, blockNumber, timestamp, true)
	}

	return nil
}

// processFailedReceipt processes a failed/reverted transaction receipt.
func (p *Processor) processFailedReceipt(ctx context.Context, receipt *types.Receipt, blockNumber, timestamp uint64) error {
	tx, err := p.chain.GetTransaction(ctx, receipt.TxHash)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("failed to fetch transaction %s: %w", receipt.TxHash.Hex(), err)
	}

	// Handle contract creation attempts (even if reverted)
	if tx.To() == nil {
		return p.processContractCreation(ctx, receipt, blockNumber, timestamp, false)
	}

	// Handle regular transaction with input data
	exists, err := p.cache.Exists(ctx, tx.To().Hex())
	if err != nil {
		return fmt.Errorf("cache lookup failed for contract address %s: %w", tx.To().Hex(), err)
	}
	if !exists {
		return nil
	}

	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return fmt.Errorf("failed to decode transaction sender for tx %s: %w", receipt.TxHash.Hex(), err)
	}

	if err := p.router.ProcessInputData(ctx, router.InputDataPayload{
		From:            from.Hex(),
		InputData:       common.Bytes2Hex(tx.Data()),
		Block:           blockNumber,
		ContractAddress: tx.To().Hex(),
		Timestamp:       timestamp,
		TxHash:          receipt.TxHash.Hex(),
	}); err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("failed to process input data for tx %s: %w", receipt.TxHash.Hex(), err)
	}

	return nil
}

// processContractCreation processes contract creation events (both successful and reverted).
func (p *Processor) processContractCreation(ctx context.Context, receipt *types.Receipt, blockNumber, timestamp uint64, success bool) error {
	tx, err := p.chain.GetTransaction(ctx, receipt.TxHash)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("failed to fetch transaction %s: %w", receipt.TxHash.Hex(), err)
	}

	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return fmt.Errorf("failed to decode transaction sender for tx %s: %w", receipt.TxHash.Hex(), err)
	}

	exists, err := p.cache.Exists(ctx, from.Hex())
	if err != nil {
		return fmt.Errorf("cache lookup failed for sender address %s: %w", from.Hex(), err)
	}
	if !exists {
		return nil
	}

	if err := p.router.ProcessContractCreation(ctx, router.ContractCreationPayload{
		From:            from.Hex(),
		Block:           blockNumber,
		ContractAddress: receipt.ContractAddress.Hex(),
		Timestamp:       timestamp,
		TxHash:          receipt.TxHash.Hex(),
		Success:         success,
	}); err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("failed to process contract creation for tx %s: %w", receipt.TxHash.Hex(), err)
	}

	return nil
}
