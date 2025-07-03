package main

import (
	"github.com/grassrootseconomics/eth-tracker/internal/cache"
	"github.com/grassrootseconomics/eth-tracker/internal/handler"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
	"github.com/lmittmann/w3"
)

func bootstrapEventRouter(cacheProvider cache.Cache, pubCB router.Callback) *router.Router {
	handlerContainer := handler.New(cacheProvider)
	router := router.New(pubCB)

	router.RegisterContractCreationHandler(handler.HandleContractCreation(handlerContainer))

	router.RegisterLogRoute(w3.H("0x26162814817e23ec5035d6a2edc6c422da2da2119e27cfca6be65cc2dc55ca4c"), handler.HandleFaucetGiveLog())
	router.RegisterLogRoute(w3.H("0xa226db3f664042183ee0281230bba26cbf7b5057e50aee7f25a175ff45ce4d7f"), handler.HandleIndexAddLog(handlerContainer))
	router.RegisterLogRoute(w3.H("0x24a12366c02e13fe4a9e03d86a8952e85bb74a456c16e4a18b6d8295700b74bb"), handler.HandleIndexRemoveLog(handlerContainer))
	router.RegisterLogRoute(w3.H("0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0"), handler.HandleOwnershipLog())
	router.RegisterLogRoute(w3.H("0x5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f62"), handler.HandlePoolDepositLog())
	router.RegisterLogRoute(w3.H("0xd6d34547c69c5ee3d2667625c188acf1006abb93e0ee7cf03925c67cf7760413"), handler.HandlePoolSwapLog())
	router.RegisterLogRoute(w3.H("0xdb9ce1a76955721ca61ac50cd1b87f9ab8620325c8619a62192c2dc7871d56b1"), handler.HandleQuoterPriceUpdateLog())
	router.RegisterLogRoute(w3.H("0x6b7e2e653f93b645d4ed7292d6429f96637084363e477c8aaea1a43ed13c284e"), handler.HandleSealStateChangeLog())
	router.RegisterLogRoute(w3.H("0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5"), handler.HandleTokenBurnLog())
	router.RegisterLogRoute(w3.H("0xab8530f87dc9b59234c4623bf917212bb2536d647574c8e7e5da92c2ede0c9f8"), handler.HandleTokenMintLog())
	router.RegisterLogRoute(w3.H("0x894e56e1dac400b4475c83d8af0f0aa44de17c62764bd82f6e768a504e242461"), handler.HandleCustodialRegistrationLog())
	router.RegisterLogRoute(w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"), handler.HandleTokenTransferLog(handlerContainer))
	router.RegisterLogRoute(w3.H("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"), handler.HandleTokenApproveLog(handlerContainer))
	router.RegisterLogRoute(w3.H("0x5f7542858008eeb041631f30e6109ae94b83a58e9a58261dd2c42c508850f939"), handler.HandleTokenTransferFromLog(handlerContainer))
	router.RegisterLogRoute(w3.H("0x06526a30af2ff868c2686df12e95844d8ae300416bbec5d5ccc2d2f4afdb17a0"), handler.HandleQuoterUpdatedLog())

	router.RegisterInputDataRoute("63e4bff4", handler.HandleFaucetGiveInputData())
	router.RegisterInputDataRoute("de82efb4", handler.HandleFaucetGiveInputData())
	router.RegisterInputDataRoute("0a3b0a4f", handler.HandleIndexAddInputData())
	router.RegisterInputDataRoute("4420e486", handler.HandleIndexAddInputData())
	router.RegisterInputDataRoute("29092d0e", handler.HandleIndexRemoveInputData())
	router.RegisterInputDataRoute("f2fde38b", handler.HandleOwnershipInputData())
	router.RegisterInputDataRoute("47e7ef24", handler.HandlePoolDepositInputData())
	router.RegisterInputDataRoute("d9caed12", handler.HandlePoolSwapInputData())
	router.RegisterInputDataRoute("ebc59dff", handler.HandleQuoterPriceUpdateInputdata())
	router.RegisterInputDataRoute("86fe212d", handler.HandleSealStateChangeInputData())
	router.RegisterInputDataRoute("42966c68", handler.HandleTokenBurnInputData())
	router.RegisterInputDataRoute("449a52f8", handler.HandleTokenMintInputData())
	router.RegisterInputDataRoute("4420e486", handler.HandleCustodialRegistrationInputData())
	router.RegisterInputDataRoute("a9059cbb", handler.HandleTokenTransferInputData(handlerContainer))
	router.RegisterInputDataRoute("23b872dd", handler.HandleTokenTransferInputData(handlerContainer))
	router.RegisterInputDataRoute("095ea7b3", handler.HandleTokenApproveInputData(handlerContainer))
	router.RegisterInputDataRoute("f912c64b", handler.HandleQuoterUpdatedInputData())

	return router
}
