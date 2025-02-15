package bridgectrl

import (
	"context"

	"github.com/fiwallets/zkevm-bridge-service/etherman"
	"github.com/fiwallets/go-ethereum/common"
	"github.com/jackc/pgx/v4"
)

// merkleTreeStore interface for the Merkle Tree
type merkleTreeStore interface {
	Get(ctx context.Context, key []byte, dbTx pgx.Tx) ([][]byte, error)
	BulkSet(ctx context.Context, rows [][]interface{}, dbTx pgx.Tx) error
	GetRoot(ctx context.Context, depositCount uint32, network uint32, dbTx pgx.Tx) ([]byte, error)
	SetRoot(ctx context.Context, root []byte, depositID uint64, network uint32, dbTx pgx.Tx) error
	GetLastDepositCount(ctx context.Context, networkID uint32, dbTx pgx.Tx) (uint32, error)
	AddRollupExitLeaves(ctx context.Context, rows [][]interface{}, dbTx pgx.Tx) error
	GetRollupExitLeavesByRoot(ctx context.Context, root common.Hash, dbTx pgx.Tx) ([]etherman.RollupExitLeaf, error)
	GetLatestRollupExitLeaves(ctx context.Context, dbTx pgx.Tx) ([]etherman.RollupExitLeaf, error)
	IsRollupExitRoot(ctx context.Context, root common.Hash, dbTx pgx.Tx) (bool, error)
}
