package processor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/db"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celo-tracker/pkg/router"
)

type (
	ProcessorOpts struct {
		Cache  cache.Cache
		Chain  chain.Chain
		DB     db.DB
		Router *router.Router
		Logg   *slog.Logger
	}

	Processor struct {
		cache  cache.Cache
		chain  chain.Chain
		db     db.DB
		router *router.Router
		logg   *slog.Logger
	}
)

func NewProcessor(o ProcessorOpts) *Processor {
	return &Processor{
		cache:  o.Cache,
		chain:  o.Chain,
		db:     o.DB,
		router: o.Router,
		logg:   o.Logg,
	}
}

func (p *Processor) ProcessBlock(ctx context.Context, blockNumber uint64) error {
	block, err := p.chain.GetBlock(ctx, blockNumber)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("block %d error: %v", blockNumber, err)
	}

	receipts, err := p.chain.GetReceipts(ctx, block.Number())
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("receipts fetch error: block %d: %v", blockNumber, err)
	}

	for _, receipt := range receipts {
		if receipt.Status > 0 {
			for _, log := range receipt.Logs {
				exists, err := p.cache.Exists(ctx, log.Address.Hex())
				if err != nil {
					return err
				}
				if exists {
					if err := p.router.ProcessLog(
						ctx,
						router.LogPayload{
							Log:       log,
							Timestamp: block.Time(),
						},
					); err != nil && !errors.Is(err, context.Canceled) {
						return fmt.Errorf("route success transaction error: tx %s: %v", receipt.TxHash.Hex(), err)
					}
				}
			}
		} else {
			tx, err := p.chain.GetTransaction(ctx, receipt.TxHash)
			if err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("get transaction error: tx %s: %v", receipt.TxHash.Hex(), err)
			}

			if tx.To() != nil {
				exists, err := p.cache.Exists(ctx, tx.To().Hex())
				if err != nil {
					return err
				}
				if exists {
					from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
					if err != nil {
						return fmt.Errorf("transaction decode error: tx %s: %v", receipt.TxHash.Hex(), err)
					}

					if err := p.router.ProcessInputData(
						ctx,
						router.InputDataPayload{
							From:            from.Hex(),
							InputData:       common.Bytes2Hex(tx.Data()),
							Block:           blockNumber,
							ContractAddress: tx.To().Hex(),
							Timestamp:       block.Time(),
							TxHash:          receipt.TxHash.Hex(),
						},
					); err != nil && !errors.Is(err, context.Canceled) {
						return fmt.Errorf("route revert transaction error: tx %s: %v", receipt.TxHash.Hex(), err)
					}
				}
			}
		}
	}

	if err := p.db.SetValue(blockNumber); err != nil {
		return err
	}
	p.logg.Debug("successfully processed block", "block", blockNumber)

	return nil
}
