//go:build edge
// +build edge

package e2e

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/fiwallets/zkevm-bridge-service/bridgectrl"
	"github.com/fiwallets/zkevm-bridge-service/db"
	"github.com/fiwallets/zkevm-bridge-service/server"
	"github.com/fiwallets/zkevm-bridge-service/test/operations"
	"github.com/fiwallets/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func depositFromL1(ctx context.Context, opsman *operations.Manager, t *testing.T) {
	amount := new(big.Int).SetUint64(250000000000000000)
	tokenAddr := common.Address{} // This means is eth
	destAddr := common.HexToAddress("0xc949254d682d8c9ad5682521675b8f43b102aec4")
	var destNetwork uint32 = 1
	// L1 Deposit
	err := opsman.SendL1Deposit(ctx, tokenAddr, amount, destNetwork, &destAddr)
	require.NoError(t, err)

	deposits, err := opsman.GetBridgeInfoByDestAddr(ctx, &destAddr)
	require.NoError(t, err)
	// Check a L2 claim tx
	err = opsman.CheckClaim(ctx, deposits[0])
	require.NoError(t, err)
}

func depositFromL2(ctx context.Context, opsman *operations.Manager, t *testing.T) {
	// Send L2 Deposit to withdraw the some funds
	var destNetwork uint32 = 0
	amount := new(big.Int).SetUint64(100000000000000000)
	tokenAddr := common.Address{} // This means is eth
	destAddr := common.HexToAddress("0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC")
	err := opsman.SendL2Deposit(ctx, tokenAddr, amount, destNetwork, &destAddr, operations.L2)
	require.NoError(t, err)

	// Get Bridge Info By DestAddr
	deposits, err := opsman.GetBridgeInfoByDestAddr(ctx, &destAddr)
	require.NoError(t, err)
	// Check globalExitRoot
	// Get the claim data
	smtProof, smtRollupProof, globalExitRoot, err := opsman.GetClaimData(ctx, deposits[0].NetworkId, deposits[0].DepositCnt)
	require.NoError(t, err)
	// Claim funds in L1
	err = opsman.SendL1Claim(ctx, deposits[0], smtProof, smtRollupProof, globalExitRoot)
	require.NoError(t, err)
}

func TestEdgeCase(t *testing.T) {
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

	require.NoError(t, operations.StartBridge())
	opsman, err := operations.NewManager(ctx, opsCfg)
	require.NoError(t, err)
	const st time.Duration = 20 // wait until the syncing is finished
	time.Sleep(st * time.Second)

	t.Run("Test a case of restart with reorg.", func(t *testing.T) {
		depositFromL1(ctx, opsman, t)
		depositFromL2(ctx, opsman, t)
		// Stop the bridge service.
		require.NoError(t, operations.StopBridge())
		// Modify the L1 blocks for L1 reorg
		require.NoError(t, opsman.UpdateBlocksForTesting(ctx, 0, 1))
		// Restart the bridge service.
		require.NoError(t, operations.StartBridge())
		time.Sleep(st * time.Second)

		depositFromL2(ctx, opsman, t)
		depositFromL1(ctx, opsman, t)
	})
}
