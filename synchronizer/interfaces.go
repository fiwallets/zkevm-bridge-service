package synchronizer

import (
	"context"
	"math/big"

	"github.com/fiwallets/zkevm-bridge-service/etherman"
	rpcTypes "github.com/0xPolygonHermez/zkevm-node/jsonrpc/types"
	"github.com/fiwallets/go-ethereum/common"
	"github.com/fiwallets/go-ethereum/core/types"
	"github.com/jackc/pgx/v4"
)

// ethermanInterface contains the methods required to interact with ethereum.
type ethermanInterface interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	GetRollupInfoByBlockRange(ctx context.Context, fromBlock uint64, toBlock *uint64) ([]etherman.Block, map[common.Hash][]etherman.Order, error)
	GetNetworkID() uint32
}

type storageInterface interface {
	GetLastBlock(ctx context.Context, networkID uint32, dbTx pgx.Tx) (*etherman.Block, error)
	Rollback(ctx context.Context, dbTx pgx.Tx) error
	BeginDBTransaction(ctx context.Context) (pgx.Tx, error)
	Commit(ctx context.Context, dbTx pgx.Tx) error
	AddBlock(ctx context.Context, block *etherman.Block, dbTx pgx.Tx) (uint64, error)
	AddGlobalExitRoot(ctx context.Context, exitRoot *etherman.GlobalExitRoot, dbTx pgx.Tx) error
	AddDeposit(ctx context.Context, deposit *etherman.Deposit, dbTx pgx.Tx) (uint64, error)
	AddClaim(ctx context.Context, claim *etherman.Claim, dbTx pgx.Tx) error
	AddTokenWrapped(ctx context.Context, tokenWrapped *etherman.TokenWrapped, dbTx pgx.Tx) error
	Reset(ctx context.Context, blockNumber uint64, networkID uint32, dbTx pgx.Tx) error
	GetPreviousBlock(ctx context.Context, networkID uint32, offset uint64, dbTx pgx.Tx) (*etherman.Block, error)
	GetNumberDeposits(ctx context.Context, origNetworkID uint32, blockNumber uint64, dbTx pgx.Tx) (uint32, error)
	AddTrustedGlobalExitRoot(ctx context.Context, trustedExitRoot *etherman.GlobalExitRoot, dbTx pgx.Tx) (bool, error)
	GetLatestL1SyncedExitRoot(ctx context.Context, dbTx pgx.Tx) (*etherman.GlobalExitRoot, error)
	CheckIfRootExists(ctx context.Context, root []byte, network uint32, dbTx pgx.Tx) (bool, error)
	GetL1ExitRootByGER(ctx context.Context, ger common.Hash, dbTx pgx.Tx) (*etherman.GlobalExitRoot, error)
	GetL2ExitRootsByGER(ctx context.Context, ger common.Hash, dbTx pgx.Tx) ([]etherman.GlobalExitRoot, error)
	UpdateL2GER(ctx context.Context, ger etherman.GlobalExitRoot, dbTx pgx.Tx) error
	AddRemoveL2GER(ctx context.Context, globalExitRoot etherman.GlobalExitRoot, dbTx pgx.Tx) error
}

type bridgectrlInterface interface {
	AddDeposit(ctx context.Context, deposit *etherman.Deposit, depositID uint64, dbTx pgx.Tx) error
	ReorgMT(ctx context.Context, depositCount, networkID uint32, dbTx pgx.Tx) error
	AddRollupExitLeaf(ctx context.Context, rollupLeaf etherman.RollupExitLeaf, dbTx pgx.Tx) error
}

type zkEVMClientInterface interface {
	GetLatestGlobalExitRoot(ctx context.Context) (common.Hash, error)
	ExitRootsByGER(ctx context.Context, globalExitRoot common.Hash) (*rpcTypes.ExitRoots, error)
}
