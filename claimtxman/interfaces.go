package claimtxman

import (
	"context"

	"github.com/0xPolygonHermez/zkevm-bridge-service/bridgectrl"
	"github.com/0xPolygonHermez/zkevm-bridge-service/claimtxman/types"
	"github.com/0xPolygonHermez/zkevm-bridge-service/etherman"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
)

type StorageInterface interface {
	AddBlock(ctx context.Context, block *etherman.Block, dbTx pgx.Tx) (uint64, error)
	UpdateL1DepositsStatus(ctx context.Context, exitRoot []byte, destinationNetwork uint32, dbTx pgx.Tx) ([]*etherman.Deposit, error)
	UpdateL2DepositsStatus(ctx context.Context, exitRoot []byte, rollupID, networkID uint32, dbTx pgx.Tx) error
	GetDepositsFromOtherL2ToClaim(ctx context.Context, destinationNetwork uint32, dbTx pgx.Tx) ([]*etherman.Deposit, error)
	GetLatestTrustedGERByDeposit(ctx context.Context, depositCnt, networkID, destinationNetwork uint32, dbTx pgx.Tx) (common.Hash, error)
	AddClaimTx(ctx context.Context, mTx types.MonitoredTx, dbTx pgx.Tx) error
	UpdateClaimTx(ctx context.Context, mTx types.MonitoredTx, dbTx pgx.Tx) error
	GetClaimTxsByStatus(ctx context.Context, statuses []types.MonitoredTxStatus, rollupID uint32, dbTx pgx.Tx) ([]types.MonitoredTx, error)
	// atomic
	Rollback(ctx context.Context, dbTx pgx.Tx) error
	BeginDBTransaction(ctx context.Context) (pgx.Tx, error)
	Commit(ctx context.Context, dbTx pgx.Tx) error
}

type bridgeServiceInterface interface {
	GetClaimProofForCompressed(ger common.Hash, depositCnt, networkID uint32, dbTx pgx.Tx) (*etherman.GlobalExitRoot, [][bridgectrl.KeyLen]byte, [][bridgectrl.KeyLen]byte, error)
	GetDepositStatus(ctx context.Context, depositCount, networkID, destNetworkID uint32) (string, error)
}
