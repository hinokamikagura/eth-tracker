package router

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	PubCallback func(context.Context, event.Event) error

	SuccessTx struct {
		Log       *types.Log
		Timestamp uint64
	}

	RevertTx struct {
		From            string
		InputData       string
		Block           uint64
		ContractAddress string
		Timestamp       uint64
		TxHash          string
	}

	Handler interface {
		Name() string
		SuccessTx(context.Context, SuccessTx, PubCallback) error
		RevertTx(context.Context, RevertTx, PubCallback) error
	}

	RouterOpts struct {
		Pub   pub.Pub
		Cache cache.Cache
	}

	Router struct {
		pub               pub.Pub
		cache             cache.Cache
		logHandlers       map[common.Hash]Handler
		inputDataHandlers map[string]Handler
	}
)

func New(o RouterOpts) *Router {
	var (
		indexAddHandler      *indexAddHandler      = &indexAddHandler{cache: o.Cache}
		indexRemoveHandler   *indexRemoveHandler   = &indexRemoveHandler{cache: o.Cache}
		tokenTransferHandler *tokenTransferHandler = &tokenTransferHandler{cache: o.Cache}
		faucetGiveHandler    *faucetGiveHandler    = &faucetGiveHandler{}
		ownershipHandler     *ownershipHandler     = &ownershipHandler{}
		poolDepositHandler   *poolDepositHandler   = &poolDepositHandler{}
		poolSwapHandler      *poolSwapHandler      = &poolSwapHandler{}
		quoterPriceHandler   *quoterPriceHandler   = &quoterPriceHandler{}
		sealHandler          *sealHandler          = &sealHandler{}
		tokenBurnHandler     *tokenBurnHandler     = &tokenBurnHandler{}
		tokenMintHandler     *tokenMintHandler     = &tokenMintHandler{}
	)

	logHandlers := map[common.Hash]Handler{
		w3.H("0x26162814817e23ec5035d6a2edc6c422da2da2119e27cfca6be65cc2dc55ca4c"): faucetGiveHandler,
		w3.H("0xa226db3f664042183ee0281230bba26cbf7b5057e50aee7f25a175ff45ce4d7f"): indexAddHandler,
		w3.H("0x24a12366c02e13fe4a9e03d86a8952e85bb74a456c16e4a18b6d8295700b74bb"): indexRemoveHandler,
		w3.H("0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0"): ownershipHandler,
		w3.H("0x5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f62"): poolDepositHandler,
		w3.H("0xd6d34547c69c5ee3d2667625c188acf1006abb93e0ee7cf03925c67cf7760413"): poolSwapHandler,
		w3.H("0xdb9ce1a76955721ca61ac50cd1b87f9ab8620325c8619a62192c2dc7871d56b1"): quoterPriceHandler,
		w3.H("0x6b7e2e653f93b645d4ed7292d6429f96637084363e477c8aaea1a43ed13c284e"): sealHandler,
		w3.H("0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5"): tokenBurnHandler,
		w3.H("0xab8530f87dc9b59234c4623bf917212bb2536d647574c8e7e5da92c2ede0c9f8"): tokenMintHandler,
		w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"): tokenTransferHandler,
	}

	inputDataHandlers := map[string]Handler{
		"63e4bff4": faucetGiveHandler,
		"de82efb4": faucetGiveHandler,
		"0a3b0a4f": indexAddHandler,
		"4420e486": indexAddHandler,
		"29092d0e": indexRemoveHandler,
		"f2fde38b": ownershipHandler,
		"47e7ef24": poolDepositHandler,
		"d9caed12": poolSwapHandler,
		"ebc59dff": quoterPriceHandler,
		"86fe212d": sealHandler,
		"42966c68": tokenBurnHandler,
		"449a52f8": tokenMintHandler,
		"a9059cbb": tokenTransferHandler,
		"23b872dd": tokenTransferHandler,
	}

	return &Router{
		pub:               o.Pub,
		logHandlers:       logHandlers,
		inputDataHandlers: inputDataHandlers,
	}
}

func (r *Router) RouteSuccessTx(ctx context.Context, msg SuccessTx) error {
	handler, ok := r.logHandlers[msg.Log.Topics[0]]
	if ok {
		return handler.SuccessTx(ctx, msg, r.pub.Send)
	}

	return nil
}

func (r *Router) RouteRevertTx(ctx context.Context, msg RevertTx) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	handler, ok := r.inputDataHandlers[msg.InputData[:8]]
	if ok {
		return handler.RevertTx(ctx, msg, r.pub.Send)
	}

	return nil
}
