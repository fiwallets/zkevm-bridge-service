package main

import (
	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/0xPolygonHermez/zkevm-node/etherman/smartcontracts/polygonzkevmglobalexitroot"
	"github.com/fiwallets/go-ethereum/accounts/abi/bind"
	"github.com/fiwallets/go-ethereum/common"
	"github.com/fiwallets/go-ethereum/ethclient"
)

const (
	gerManAddr = "0xa40d5f56745a118d0906a34e69aec8c0db1cb8fa"

	nodeURL = "http://localhost:8124"
)

func main() {
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		log.Fatal("error conecting to the node. Error: ", err)
	}
	g, err := polygonzkevmglobalexitroot.NewPolygonzkevmglobalexitroot(common.HexToAddress(gerManAddr), client)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	rollupExitRoot, err := g.LastRollupExitRoot(&bind.CallOpts{})
	if err != nil {
		log.Fatal("Error: ", err)
	}
	// ger, err := g.GlobalExitRootMap(&bind.CallOpts{})
	// if err != nil {
	// 	log.Fatal("Error: ", err)
	// }
	// log.Info("ger! ", common.BytesToAddress(ger[:]))
	log.Info("rollupExitRoot! ", common.BytesToHash(rollupExitRoot[:]))
}
