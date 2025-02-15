package claimtxman

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/fiwallets/zkevm-bridge-service/claimtxman/types"
	ctmtypes "github.com/fiwallets/zkevm-bridge-service/claimtxman/types"
	"github.com/fiwallets/zkevm-bridge-service/etherman"
	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/fiwallets/zkevm-bridge-service/utils"
	"github.com/fiwallets/zkevm-bridge-service/utils/gerror"
	"github.com/0xPolygonHermez/zkevm-node/state/runtime"
	"github.com/fiwallets/go-ethereum"
	"github.com/fiwallets/go-ethereum/accounts/abi/bind"
	"github.com/fiwallets/go-ethereum/common"
	"github.com/jackc/pgx/v4"
)

const (
	maxHistorySize  = 10
	keyLen          = 32
	mtHeight        = 32
	LeafTypeMessage = uint8(1)
)

// ClaimTxManager is the claim transaction manager for L2.
type ClaimTxManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	// client is the ethereum client
	l2Node          *utils.Client
	l2NetworkID     uint32
	bridgeService   bridgeServiceInterface
	cfg             Config
	chExitRootEvent chan *etherman.GlobalExitRoot
	chSynced        chan uint32
	storage         StorageInterface
	auth            *bind.TransactOpts
	rollupID        uint32
	l2Synced        bool
	nonceCache      *NonceCache
	monitorTxs      types.TxMonitorer
}

// NewClaimTxManager creates a new claim transaction manager.
func NewClaimTxManager(ctx context.Context, cfg Config, chExitRootEvent chan *etherman.GlobalExitRoot,
	chSynced chan uint32,
	l2NodeURL string,
	l2NetworkID uint32,
	l2BridgeAddr common.Address,
	bridgeService bridgeServiceInterface,
	storage interface{},
	rollupID uint32,
	etherMan EthermanI,
	nonceCache *NonceCache,
	auth *bind.TransactOpts) (*ClaimTxManager, error) {
	client, err := utils.NewClient(ctx, l2NodeURL, l2BridgeAddr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)

	var monitorTx ctmtypes.TxMonitorer
	if cfg.GroupingClaims.Enabled {
		log.Info("ClaimTxManager working in compressor mode to group claim txs")
		monitorTx = NewMonitorCompressedTxs(ctx, storage.(StorageCompressedInterface), client, cfg, nonceCache, auth, etherMan, utils.NewTimeProviderSystemLocalTime(), cfg.GroupingClaims.GasOffset, rollupID)
	} else {
		log.Info("ClaimTxManager working in regular mode to send claim txs individually")
		monitorTx = NewMonitorTxs(ctx, storage.(StorageInterface), client, cfg, nonceCache, rollupID, auth)
	}
	return &ClaimTxManager{
		ctx:             ctx,
		cancel:          cancel,
		l2Node:          client,
		l2NetworkID:     l2NetworkID,
		bridgeService:   bridgeService,
		cfg:             cfg,
		chExitRootEvent: chExitRootEvent,
		chSynced:        chSynced,
		storage:         storage.(StorageInterface),
		auth:            auth,
		rollupID:        rollupID,
		nonceCache:      nonceCache,
		monitorTxs:      monitorTx,
	}, err
}

// Start will start the tx management, reading txs from storage,
// send then to the blockchain and keep monitoring them until they
// get mined
func (tm *ClaimTxManager) Start() {
	ticker := time.NewTicker(tm.cfg.FrequencyToMonitorTxs.Duration)
	compressorTicker := time.NewTicker(tm.cfg.GroupingClaims.FrequencyToProcessCompressedClaims.Duration)
	var ger = &etherman.GlobalExitRoot{}
	var latestProcessedGer common.Hash
	for {
		select {
		case <-tm.ctx.Done():
			ticker.Stop()
			return
		case netID := <-tm.chSynced:
			if netID == tm.l2NetworkID && !tm.l2Synced {
				log.Info("NetworkID synced: ", netID)
				tm.l2Synced = true
			}
		case ger = <-tm.chExitRootEvent:
			if tm.l2Synced {
				log.Debugf("RollupID: %d UpdateDepositsStatus for ger: %s", tm.rollupID, ger.GlobalExitRoot.String())
				if tm.cfg.GroupingClaims.Enabled {
					log.Debugf("rollupID: %d, Ger value updated and ready to be processed...", tm.rollupID)
					continue
				}
				go func() {
					err := tm.updateDepositsStatus(ger)
					if err != nil {
						log.Errorf("rollupID: %d, failed to update deposits status: %v", tm.rollupID, err)
					}
				}()
			} else {
				log.Infof("Waiting for networkID %d to be synced before processing deposits", tm.l2NetworkID)
			}
		case <-compressorTicker.C:
			if tm.l2Synced && tm.cfg.GroupingClaims.Enabled && ger.GlobalExitRoot != latestProcessedGer {
				log.Infof("RollupID: %d,Processing deposits for ger: %s", tm.rollupID, ger.GlobalExitRoot.String())
				go func() {
					err := tm.updateDepositsStatus(ger)
					if err != nil {
						log.Errorf("rollupID: %d, failed to update deposits status: %v", tm.rollupID, err)
					}
				}()
				latestProcessedGer = ger.GlobalExitRoot
			}
		case <-ticker.C:
			err := tm.monitorTxs.MonitorTxs(tm.ctx)
			if err != nil {
				log.Errorf("rollupID: %d, failed to monitor txs: %v", tm.rollupID, err)
			}
		}
	}
}

func (tm *ClaimTxManager) updateDepositsStatus(ger *etherman.GlobalExitRoot) error {
	dbTx, err := tm.storage.BeginDBTransaction(tm.ctx)
	if err != nil {
		return err
	}
	err = tm.processDepositStatus(ger, dbTx)
	if err != nil {
		log.Errorf("rollupID: %d, error processing ger. Error: %v", tm.rollupID, err)
		rollbackErr := tm.storage.Rollback(tm.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("rollupID: %d, claimtxman error rolling back state. RollbackErr: %v, err: %s", tm.rollupID, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}
	err = tm.storage.Commit(tm.ctx, dbTx)
	if err != nil {
		log.Errorf("rollupID: %d, AddClaimTx committing dbTx. Err: %v", tm.rollupID, err)
		rollbackErr := tm.storage.Rollback(tm.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("rollupID: %d, claimtxman error rolling back state. RollbackErr: %s, err: %s", tm.rollupID, rollbackErr.Error(), err.Error())
			return rollbackErr
		}
		return err
	}
	return nil
}

func (tm *ClaimTxManager) processDepositStatus(ger *etherman.GlobalExitRoot, dbTx pgx.Tx) error {
	var (
		deposits       []*etherman.Deposit
		globalExitRoot = ger.GlobalExitRoot
		err            error
	)
	if ger.BlockID != 0 && ger.NetworkID == 0 { // L2 exit root is updated
		log.Infof("RollupID: %d, Rollup exitroot %v is updated", tm.rollupID, ger.ExitRoots[1])
		err = tm.storage.UpdateL2DepositsStatus(tm.ctx, ger.ExitRoots[1][:], tm.rollupID, tm.l2NetworkID, dbTx)
		if err != nil {
			log.Errorf("rollupID: %d, error updating L2DepositsStatus. Error: %v", tm.rollupID, err)
			return err
		}
		// If L2 claims processor is enabled
		if tm.cfg.AreClaimsBetweenL2sEnabled {
			log.Debugf("rollupID: %d, getting L2 deposits to autoClaim", tm.rollupID)
			deposits, err = tm.storage.GetDepositsFromOtherL2ToClaim(tm.ctx, tm.l2NetworkID, dbTx)
			if err != nil {
				log.Errorf("rollupID: %d, error getting deposits from other L2 to claim. Error: %v", tm.rollupID, err)
				return err
			}
			if len(deposits) > 0 {
				globalExitRoot, err = tm.storage.GetLatestTrustedGERByDeposit(tm.ctx, deposits[0].DepositCount, deposits[0].NetworkID, deposits[0].DestinationNetwork, dbTx)
				if errors.Is(err, gerror.ErrStorageNotFound) {
					log.Infof("RollupID: %d, not fully synced yet. Retrying in 2s...")
					time.Sleep(tm.cfg.RetryInterval.Duration)
					globalExitRoot, err = tm.storage.GetLatestTrustedGERByDeposit(tm.ctx, deposits[0].DepositCount, deposits[0].NetworkID, deposits[0].DestinationNetwork, dbTx)
					if errors.Is(err, gerror.ErrStorageNotFound) {
						log.Infof("RollupID: %d, Still missing. Not fully synced yet. It will retry it later...")
					} else if err != nil {
						log.Errorf("rollupID: %d, error getting the latest trusted GER by deposit the second time. Error: %v", tm.rollupID, err)
						return err
					}
				} else if err != nil {
					log.Errorf("rollupID: %d, error getting the latest trusted GER by deposit. Error: %v", tm.rollupID, err)
					return err
				}
			}
		}
	} else { // L1 exit root is updated in the trusted state
		log.Infof("RollupID: %d, Mainnet exitroot %v is updated", tm.rollupID, ger.ExitRoots[0])
		deposits, err = tm.storage.UpdateL1DepositsStatus(tm.ctx, ger.ExitRoots[0][:], tm.l2NetworkID, dbTx)
		if err != nil {
			log.Errorf("rollupID: %d, error getting and updating L1DepositsStatus. Error: %v", tm.rollupID, err)
			return err
		}
	}
	for _, deposit := range deposits {
		if tm.l2NetworkID != deposit.DestinationNetwork {
			log.Infof("Ignoring deposit id: %d deposit count:%d dest_net: %d, we are:%d", deposit.Id, deposit.DepositCount, deposit.DestinationNetwork, tm.l2NetworkID)
			continue
		}

		claimHash, err := tm.bridgeService.GetDepositStatus(tm.ctx, deposit.DepositCount, deposit.NetworkID, deposit.DestinationNetwork)
		if err != nil {
			log.Errorf("rollupID: %d, error getting deposit status for deposit id %d. Error: %v", tm.rollupID, deposit.Id, err)
			return err
		}
		if len(claimHash) > 0 || deposit.LeafType == LeafTypeMessage && !tm.isDepositMessageAllowed(deposit) {
			log.Infof("RollupID: %d, Ignoring deposit Id: %d, leafType: %d, claimHash: %s, deposit.OriginalAddress: %s", tm.rollupID, deposit.Id, deposit.LeafType, claimHash, deposit.OriginalAddress.String())
			continue
		}

		log.Infof("RollupID: %d, create the claim tx for the deposit count %d. Deposit Id: %d", tm.rollupID, deposit.DepositCount, deposit.Id)
		ger, proof, rollupProof, err := tm.bridgeService.GetClaimProofForCompressed(globalExitRoot, deposit.DepositCount, deposit.NetworkID, dbTx)
		if err != nil {
			log.Errorf("rollupID: %d, error getting Claim Proof for deposit Id %d. Error: %v", tm.rollupID, deposit.Id, err)
			return err
		}
		var (
			mtProof       [mtHeight][keyLen]byte
			mtRollupProof [mtHeight][keyLen]byte
		)
		for i := 0; i < mtHeight; i++ {
			mtProof[i] = proof[i]
			mtRollupProof[i] = rollupProof[i]
		}
		tx, err := tm.l2Node.BuildSendClaim(tm.ctx, deposit, mtProof, mtRollupProof,
			&etherman.GlobalExitRoot{
				ExitRoots: []common.Hash{
					ger.ExitRoots[0],
					ger.ExitRoots[1],
				}}, 1, 1, 1,
			tm.auth)
		if err != nil {
			log.Errorf("rollupID: %d, error BuildSendClaim tx for deposit Id: %d. Error: %v", tm.rollupID, deposit.Id, err)
			return err
		}
		if err = tm.addClaimTx(deposit.Id, tm.auth.From, tx.To(), nil, tx.Data(), ger.GlobalExitRoot, dbTx); err != nil {
			log.Errorf("rollupID: %d, error adding claim tx for deposit Id: %d Error: %v", tm.rollupID, deposit.Id, err)
			return err
		}
	}
	return nil
}

func (tm *ClaimTxManager) isDepositMessageAllowed(deposit *etherman.Deposit) bool {
	for _, addr := range tm.cfg.AuthorizedClaimMessageAddresses {
		if deposit.OriginalAddress == addr {
			log.Infof("RollupID: %d, MessageBridge from authorized account detected: %+v, account: %s", tm.rollupID, deposit, addr.String())
			return true
		}
	}
	log.Infof("RollupID: %d, MessageBridge Not authorized. DepositCount: %d. DepositID: %d", tm.rollupID, deposit.DepositCount, deposit.Id)
	return false
}

func (tm *ClaimTxManager) addClaimTx(depositID uint64, from common.Address, to *common.Address, value *big.Int, data []byte, ger common.Hash, dbTx pgx.Tx) error {
	// get gas
	tx := ethereum.CallMsg{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	}
	gas, err := tm.l2Node.EstimateGas(tm.ctx, tx)
	for i := 1; err != nil && err.Error() != runtime.ErrExecutionReverted.Error() && i < tm.cfg.RetryNumber; i++ {
		log.Warnf("rollupID: %d, error while doing gas estimation. Retrying... Error: %v, Data: %s", tm.rollupID, err, common.Bytes2Hex(data))
		time.Sleep(tm.cfg.RetryInterval.Duration)
		gas, err = tm.l2Node.EstimateGas(tm.ctx, tx)
	}
	if err != nil {
		var b string
		block, err2 := tm.l2Node.Client.BlockByNumber(tm.ctx, nil)
		if err2 != nil {
			log.Error("error getting blockNumber. Error: ", err2)
			b = "latest"
		} else {
			b = fmt.Sprintf("%x", block.Number())
		}
		log.Warnf(`Use the next command to debug it manually.
		curl --location --request POST 'http://localhost:8545' \
		--header 'Content-Type: application/json' \
		--data-raw '{
			"jsonrpc": "2.0",
			"method": "eth_call",
			"params": [{"from": "%s","to":"%s","data":"0x%s"},"0x%s"],
			"id": 1
		}'`, from, to, common.Bytes2Hex(data), b)
		log.Errorf("rollupID: %d, failed to estimate gas. Ignoring tx... Error: %v, data: %s, GER: %s", tm.rollupID, err, common.Bytes2Hex(data), ger.String())
		return nil
	}
	// get next nonce
	nonce, err := tm.nonceCache.GetNextNonce(from)
	if err != nil {
		err := fmt.Errorf("rollupID: %d, failed to get current nonce: %v", tm.rollupID, err)
		log.Errorf("error getting next nonce. Error: %s", err.Error())
		return err
	}

	// create monitored tx
	mTx := ctmtypes.MonitoredTx{
		DepositID: depositID, From: from, To: to,
		Nonce: nonce, Value: value, Data: data,
		Gas: gas, Status: ctmtypes.MonitoredTxStatusCreated,
		GlobalExitRoot: ger,
	}

	// add to storage
	err = tm.storage.AddClaimTx(tm.ctx, mTx, dbTx)
	if err != nil {
		err := fmt.Errorf("rollupID: %d, failed to add tx to get monitored: %v", tm.rollupID, err)
		log.Errorf("error adding claim tx to db. Error: %s", err.Error())
		return err
	}

	return nil
}

// ReviewMonitoredTx checks if tx needs to be updated
// accordingly to the current information stored and the current
// state of the blockchain
func (tm *ClaimTxManager) ReviewMonitoredTx(ctx context.Context, mTx *ctmtypes.MonitoredTx, reviewNonce bool) error {
	mTxLog := log.WithFields("monitoredTx", mTx.DepositID, "rollupID", tm.rollupID)
	mTxLog.Debug("reviewing")
	// get gas
	tx := ethereum.CallMsg{
		From:  mTx.From,
		To:    mTx.To,
		Value: mTx.Value,
		Data:  mTx.Data,
	}
	gas, err := tm.l2Node.EstimateGas(ctx, tx)
	for i := 1; err != nil && err.Error() != runtime.ErrExecutionReverted.Error() && i < tm.cfg.RetryNumber; i++ {
		mTxLog.Warnf("error during gas estimation. Retrying... Error: %v, Data: %s", err, common.Bytes2Hex(tx.Data))
		time.Sleep(tm.cfg.RetryInterval.Duration)
		gas, err = tm.l2Node.EstimateGas(tm.ctx, tx)
	}
	if err != nil {
		err := fmt.Errorf("failed to estimate gas. Error: %v, Data: %s", err, common.Bytes2Hex(tx.Data))
		mTxLog.Errorf("error: %s", err.Error())
		return err
	}

	// check gas
	if gas > mTx.Gas {
		mTxLog.Infof("monitored tx gas updated from %v to %v", mTx.Gas, gas)
		mTx.Gas = gas
	}

	if reviewNonce {
		// check nonce
		nonce, err := tm.nonceCache.GetNextNonce(mTx.From)
		if err != nil {
			err := fmt.Errorf("failed to get nonce: %v", err)
			mTxLog.Errorf(err.Error())
			return err
		}
		if nonce > mTx.Nonce {
			mTxLog.Infof("monitored tx nonce updated from %v to %v", mTx.Nonce, nonce)
			mTx.Nonce = nonce
		}
	}

	return nil
}
