package main

import (
	"context"
	"math/big"

	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/fiwallets/zkevm-bridge-service/utils"
	"github.com/fiwallets/go-ethereum/common"
)

const (
	l1BridgeAddr = "0xFe12ABaa190Ef0c8638Ee0ba9F828BF41368Ca0E"

	l1AccHexAddress    = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	l1AccHexPrivateKey = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	l1NetworkURL       = "http://localhost:8545"

	funds              = 90000000000000000 // nolint
	destNetwork uint32 = 1
)

var tokenAddr = common.Address{}

func main() {
	ctx := context.Background()
	client, err := utils.NewClient(ctx, l1NetworkURL, common.HexToAddress(l1BridgeAddr))
	if err != nil {
		log.Fatal("Error: ", err)
	}
	auth, err := client.GetSigner(ctx, l1AccHexPrivateKey)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	amount := big.NewInt(funds)
	emptyAddr := common.Address{}
	if tokenAddr == emptyAddr {
		auth.Value = amount
	}
	destAddr := common.HexToAddress(l1AccHexAddress)
	log.Info("Sending bridge tx...")
	err = client.SendBridgeAsset(ctx, tokenAddr, amount, destNetwork, &destAddr, []byte{}, auth)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	log.Info("Success!")
}
