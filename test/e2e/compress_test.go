//go:build e2ecompress
// +build e2ecompress

package e2e

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/fiwallets/zkevm-bridge-service/bridgectrl"
	"github.com/fiwallets/zkevm-bridge-service/db"
	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/fiwallets/zkevm-bridge-service/server"
	"github.com/fiwallets/zkevm-bridge-service/test/operations"
	"github.com/fiwallets/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

const (
	defaultInterval = 10 * time.Second
	defaultDeadline = 600 * time.Second
)

func multiDepositFromL1(ctx context.Context, opsman *operations.Manager, destAddr common.Address, t *testing.T) {
	amount := new(big.Int).SetUint64(250000000000000000)
	tokenAddr := common.Address{} // This means is eth
	var destNetwork uint32 = 1
	// L1 Deposit to R1
	err := opsman.SendMultipleL1Deposit(ctx, tokenAddr, amount, destNetwork, &destAddr, 30)
	require.NoError(t, err)

	// L1 Deposit to R2
	destNetwork = 2
	err = opsman.SendMultipleL1Deposit(ctx, tokenAddr, amount, destNetwork, &destAddr, 30)
	require.NoError(t, err)

	deposits, err := opsman.GetBridgeInfoByDestAddr(ctx, &destAddr)
	require.NoError(t, err)
	time.Sleep(5 * time.Second) // Delay to give time to the synchronizer to read all events
	// Check a L2 claim tx
	err = opsman.CustomCheckClaim(ctx, deposits[0], defaultInterval, defaultDeadline)
	require.NoError(t, err)
}

func TestClaimCompressor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	opsCfg := &operations.Config{
		L1NetworkURL: "http://localhost:8545",
		L2NetworkURL: "http://localhost:8123",
		L2NetworkID:  1,
		Storage: db.Config{
			Database: "postgres",
			Name:     "test_db",
			User:     "test_user",
			Password: "test_password",
			Host:     "localhost",
			Port:     "5435",
			MaxConns: 10,
		},
		BT: bridgectrl.Config{
			Store:  "postgres",
			Height: uint8(32),
		},
		BS: server.Config{
			GRPCPort:         "9090",
			HTTPPort:         "8080",
			CacheSize:        100000,
			DefaultPageLimit: 25,
			MaxPageLimit:     100,
			BridgeVersion:    "v1",
			DB: db.Config{
				Database: "postgres",
				Name:     "test_db",
				User:     "test_user",
				Password: "test_password",
				Host:     "localhost",
				Port:     "5435",
				MaxConns: 10,
			},
		},
	}

	err := os.Setenv("ZKEVM_BRIDGE_CLAIMTXMANAGER_GROUPINGCLAIMS_ENABLED", "true")
	require.NoError(t, err)
	require.NoError(t, operations.StartBridge3())
	opsman, err := operations.NewManager(ctx, opsCfg)
	require.NoError(t, err)
	const st time.Duration = 20 // wait until the syncing is finished
	time.Sleep(st * time.Second)

	t.Run("Test claim compressor", func(t *testing.T) {
		log.Info("ZKEVM_BRIDGE_CLAIMTXMANAGER_GROUPINGCLAIMS_ENABLED: ", os.Getenv("ZKEVM_BRIDGE_CLAIMTXMANAGER_GROUPINGCLAIMS_ENABLED"))
		destAddr := common.HexToAddress("0xc949254d682d8c9ad5682521675b8f43b102aec4")
		multiDepositFromL1(ctx, opsman, destAddr, t)
		// Check number claim events
		numberClaims, err := opsman.GetNumberClaims(ctx, destAddr.String())
		require.NoError(t, err)
		require.Equal(t, 60, numberClaims)
		// Check L2 balance
		balance, err := opsman.CheckAccountBalance(ctx, "l2", &destAddr)
		require.NoError(t, err)
		require.Equal(t, "7500000000000000435", balance.String())
		maxGroupID, err := opsman.GetLatestMonitoredTxGroupID(ctx)
		require.NoError(t, err)
		require.Equal(t, uint64(4), maxGroupID)
	})
}
