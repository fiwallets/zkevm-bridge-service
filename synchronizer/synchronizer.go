package synchronizer

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/fiwallets/zkevm-bridge-service/etherman"
	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/fiwallets/zkevm-bridge-service/utils/gerror"
	"github.com/fiwallets/go-ethereum/common"
	"github.com/jackc/pgx/v4"
)

// Synchronizer connects L1 and L2
type Synchronizer interface {
	Sync() error
	Stop()
}

// ClientSynchronizer connects L1 and L2
type ClientSynchronizer struct {
	etherMan          ethermanInterface
	bridgeCtrl        bridgectrlInterface
	storage           storageInterface
	ctx               context.Context
	cancelCtx         context.CancelFunc
	genBlockNumber    uint64
	cfg               Config
	networkID         uint32
	chExitRootEventL2 chan *etherman.GlobalExitRoot
	chsExitRootEvent  []chan *etherman.GlobalExitRoot
	chSynced          chan uint32
	zkEVMClient       zkEVMClientInterface
	synced            bool
	l1RollupExitRoot  common.Hash
	allNetworkIDs     []uint32
	sovereignChain    bool
}

// NewSynchronizer creates and initializes an instance of Synchronizer
func NewSynchronizer(
	parentCtx context.Context,
	storage interface{},
	bridge bridgectrlInterface,
	ethMan ethermanInterface,
	zkEVMClient zkEVMClientInterface,
	genBlockNumber uint64,
	chExitRootEventL2 chan *etherman.GlobalExitRoot,
	chsExitRootEvent []chan *etherman.GlobalExitRoot,
	chSynced chan uint32,
	cfg Config,
	allNetworkIDs []uint32,
	sovereignChain bool) (Synchronizer, error) {
	ctx, cancel := context.WithCancel(parentCtx)
	networkID := ethMan.GetNetworkID()
	ger, err := storage.(storageInterface).GetLatestL1SyncedExitRoot(ctx, nil)
	if err != nil {
		if err == gerror.ErrStorageNotFound {
			ger.ExitRoots = []common.Hash{{}, {}}
		} else {
			log.Fatal("error getting last L1 synced exitroot. Error: ", err)
		}
	}

	if networkID == 0 {
		return &ClientSynchronizer{
			bridgeCtrl:       bridge,
			storage:          storage.(storageInterface),
			etherMan:         ethMan,
			ctx:              ctx,
			cancelCtx:        cancel,
			genBlockNumber:   genBlockNumber,
			cfg:              cfg,
			networkID:        networkID,
			chSynced:         chSynced,
			chsExitRootEvent: chsExitRootEvent,
			l1RollupExitRoot: ger.ExitRoots[1],
			allNetworkIDs:    allNetworkIDs,
		}, nil
	}
	return &ClientSynchronizer{
		bridgeCtrl:        bridge,
		storage:           storage.(storageInterface),
		etherMan:          ethMan,
		ctx:               ctx,
		cancelCtx:         cancel,
		genBlockNumber:    genBlockNumber,
		cfg:               cfg,
		chSynced:          chSynced,
		zkEVMClient:       zkEVMClient,
		chExitRootEventL2: chExitRootEventL2,
		networkID:         networkID,
		sovereignChain:    sovereignChain,
	}, nil
}

var waitDuration = time.Duration(0)

// Sync function will read the last state synced and will continue from that point.
// Sync() will read blockchain events to detect rollup updates
func (s *ClientSynchronizer) Sync() error {
	// If there is no lastEthereumBlock means that sync from the beginning is necessary. If not, it continues from the retrieved ethereum block
	// Get the latest synced block. If there is no block on db, use genesis block
	log.Infof("NetworkID: %d, Synchronization started", s.networkID)
	lastBlockSynced, err := s.storage.GetLastBlock(s.ctx, s.networkID, nil)
	if err != nil {
		if err == gerror.ErrStorageNotFound {
			log.Warnf("networkID: %d, error getting the latest ethereum block. No data stored. Setting genesis block. Error: %v", s.networkID, err)
			lastBlockSynced = &etherman.Block{
				BlockNumber: s.genBlockNumber,
				NetworkID:   s.networkID,
			}
			log.Warnf("networkID: %d, error getting the latest block. No data stored. Using initial block: %+v. Error: %s",
				s.networkID, lastBlockSynced, err.Error())
		} else {
			log.Fatalf("networkID: %d, unexpected error getting the latest block. Error: %s", s.networkID, err.Error())
		}
	}
	log.Debugf("NetworkID: %d, initial lastBlockSynced: %+v", s.networkID, lastBlockSynced)
	for {
		select {
		case <-s.ctx.Done():
			log.Debugf("NetworkID: %d, synchronizer ctx done", s.networkID)
			return nil
		case <-time.After(waitDuration):
			log.Debugf("NetworkID: %d, syncing...", s.networkID)
			//Sync L1Blocks
			if lastBlockSynced, err = s.syncBlocks(lastBlockSynced); err != nil {
				log.Warnf("networkID: %d, error syncing blocks: %v", s.networkID, err)
				lastBlockSynced, err = s.storage.GetLastBlock(s.ctx, s.networkID, nil)
				if errors.Is(err, gerror.ErrStorageNotFound) {
					lastBlockSynced = &etherman.Block{
						BlockNumber: s.genBlockNumber,
						NetworkID:   s.networkID,
					}
					log.Warnf("networkID: %d, error getting the latest block. No data stored. Using genesis as initial block: %+v. Error: %s",
						s.networkID, lastBlockSynced, err.Error())
				} else if err != nil {
					log.Fatalf("networkID: %d, error getting lastBlockSynced to resume the synchronization... Error: ", s.networkID, err)
				}
				if s.ctx.Err() != nil {
					continue
				}
			}
			if !s.synced {
				// Check latest Block
				header, err := s.etherMan.HeaderByNumber(s.ctx, nil)
				if err != nil {
					log.Warnf("networkID: %d, error getting latest block from. Error: %s", s.networkID, err.Error())
					continue
				}
				lastKnownBlock := header.Number.Uint64()
				if lastBlockSynced.BlockNumber == lastKnownBlock && !s.synced {
					log.Infof("NetworkID %d Synced!", s.networkID)
					waitDuration = s.cfg.SyncInterval.Duration
					s.synced = true
					s.chSynced <- s.networkID
				}
				if lastBlockSynced.BlockNumber > lastKnownBlock {
					if s.networkID == 0 {
						log.Fatalf("networkID: %d, error: latest Synced BlockNumber (%d) is higher than the latest Proposed block (%d) in the network", s.networkID, lastBlockSynced.BlockNumber, lastKnownBlock)
					} else {
						log.Errorf("networkID: %d, error: latest Synced BlockNumber (%d) is higher than the latest Proposed block (%d) in the network", s.networkID, lastBlockSynced.BlockNumber, lastKnownBlock)
						err = s.resetState(lastKnownBlock)
						if err != nil {
							log.Errorf("networkID: %d, error resetting the state to a previous block. Error: %v", s.networkID, err)
							continue
						}
					}
				}
			} else { // Sync Trusted GlobalExitRoots if L2 network is synced
				if s.networkID == 0 || s.sovereignChain { // if it is L1 or sovereignChain, trustedsync must be disabled
					continue
				}
				log.Infof("networkID: %d, Virtual state is synced, getting trusted state", s.networkID)
				err = s.syncTrustedState()
				if err != nil {
					log.Errorf("networkID: %d, error getting current trusted state", s.networkID)
				}
			}
		}
	}
}

// Stop function stops the synchronizer
func (s *ClientSynchronizer) Stop() {
	log.Infof("NetworkID: %d, Stopping synchronizer and cancelling context", s.networkID)
	s.cancelCtx()
}

func (s *ClientSynchronizer) syncTrustedState() error {
	lastGER, err := s.zkEVMClient.GetLatestGlobalExitRoot(s.ctx)
	if err != nil {
		log.Warnf("networkID: %d, failed to get latest ger from trusted state. Error: %v", s.networkID, err)
		return err
	}
	if lastGER == (common.Hash{}) {
		log.Debugf("networkID: %d, syncTrustedState: skipping GlobalExitRoot because there is no result", s.networkID)
		return nil
	}
	exitRoots, err := s.zkEVMClient.ExitRootsByGER(s.ctx, lastGER)
	if err != nil {
		log.Warnf("networkID: %d, failed to get exitRoots from trusted state. Error: %v", s.networkID, err)
		return err
	}
	if exitRoots == nil {
		log.Debugf("networkID: %d, syncTrustedState: skipping exitRoots because there is no result", s.networkID)
		return nil
	}
	ger := &etherman.GlobalExitRoot{
		NetworkID:      s.networkID,
		GlobalExitRoot: lastGER,
		ExitRoots: []common.Hash{
			exitRoots.MainnetExitRoot,
			exitRoots.RollupExitRoot,
		},
	}
	isUpdated, err := s.storage.AddTrustedGlobalExitRoot(s.ctx, ger, nil)
	if err != nil {
		log.Errorf("networkID: %d, error storing latest trusted globalExitRoot. Error: %v", s.networkID, err)
		return err
	}
	if isUpdated {
		log.Debug("adding trusted ger to the channels. GER: ", lastGER)
		s.chExitRootEventL2 <- ger
	}
	return nil
}

// This function syncs the node from a specific block to the latest
func (s *ClientSynchronizer) syncBlocks(lastBlockSynced *etherman.Block) (*etherman.Block, error) {
	// Call the blockchain to retrieve data
	header, err := s.etherMan.HeaderByNumber(s.ctx, nil)
	if err != nil {
		return lastBlockSynced, err
	}
	lastKnownBlock := header.Number
	// This function will read events fromBlockNum to latestEthBlock. Check reorg to be sure that everything is ok.
	block, err := s.checkReorg(lastBlockSynced, nil)
	if err != nil {
		log.Errorf("networkID: %d, error checking reorgs. Retrying... Err: %s", s.networkID, err.Error())
		return lastBlockSynced, fmt.Errorf("networkID: %d, error checking reorgs", s.networkID)
	}
	if block != nil {
		err = s.resetState(block.BlockNumber)
		if err != nil {
			log.Errorf("networkID: %d, error resetting the state to a previous block. Retrying... Error: %s", s.networkID, err.Error())
			return lastBlockSynced, fmt.Errorf("networkID: %d, error resetting the state to a previous block", s.networkID)
		}
		return block, nil
	}
	log.Debugf("NetworkID: %d, after checkReorg: no reorg detected", s.networkID)

	fromBlock := lastBlockSynced.BlockNumber + 1
	if s.synced {
		fromBlock = lastBlockSynced.BlockNumber
	}
	toBlock := fromBlock + s.cfg.SyncChunkSize

	for {
		if toBlock > lastKnownBlock.Uint64() {
			log.Debugf("NetworkID: %d, Setting toBlock to the lastKnownBlock: %s", s.networkID, lastKnownBlock.String())
			toBlock = lastKnownBlock.Uint64()
			if !s.synced {
				fromBlock = lastBlockSynced.BlockNumber
				log.Infof("NetworkID %d Synced!", s.networkID)
				waitDuration = s.cfg.SyncInterval.Duration
				s.synced = true
				s.chSynced <- s.networkID
			}
		}
		if fromBlock > toBlock {
			log.Debugf("NetworkID: %d, FromBlock is higher than toBlock. Skipping...", s.networkID)
			return lastBlockSynced, nil
		}

		log.Debugf("NetworkID: %d, Getting bridge info from block %d to block %d", s.networkID, fromBlock, toBlock)
		// This function returns the rollup information contained in the ethereum blocks and an extra param called order.
		// Order param is a map that contains the event order to allow the synchronizer store the info in the same order that is read.
		// Name can be different in the order struct. This name is an identifier to check if the next info that must be stored in the db.
		// The value pos (position) tells what is the array index where this value is.
		blocks, order, err := s.etherMan.GetRollupInfoByBlockRange(s.ctx, fromBlock, &toBlock)
		if err != nil {
			return lastBlockSynced, err
		}

		if fromBlock == s.genBlockNumber {
			if len(blocks) == 0 || (len(blocks) != 0 && blocks[0].BlockNumber != s.genBlockNumber) {
				log.Debugf("NetworkID: %d. adding genesis empty block", s.networkID)
				blocks = append([]etherman.Block{{}}, blocks...)
			}
		} else if fromBlock < s.genBlockNumber {
			err := fmt.Errorf("networkID: %d. fromBlock %d is lower than the genesisBlockNumber %d", s.networkID, fromBlock, s.genBlockNumber)
			log.Warn(err)
			return lastBlockSynced, err
		}
		if s.synced {
			var initBlockReceived *etherman.Block
			if len(blocks) != 0 {
				initBlockReceived = &blocks[0]
				// First position of the array must be deleted
				blocks = removeBlockElement(blocks, 0)
			} else {
				// Reorg detected
				log.Infof("NetworkID: %d, reorg detected in block %d while querying GetRollupInfoByBlockRange. Rolling back to at least the previous block", s.networkID, fromBlock)
				prevBlock, err := s.storage.GetPreviousBlock(s.ctx, s.networkID, 1, nil)
				if errors.Is(err, gerror.ErrStorageNotFound) {
					log.Warnf("networkID: %d, error checking reorg: previous block not found in db: %v", s.networkID, err)
					prevBlock = &etherman.Block{}
				} else if err != nil {
					log.Errorf("networkID: %d, error getting previousBlock from db. Error: %v", s.networkID, err)
					return lastBlockSynced, err
				}
				blockReorged, err := s.checkReorg(prevBlock, nil)
				if err != nil {
					log.Errorf("networkID: %d, error checking reorgs in previous blocks. Error: %v", s.networkID, err)
					return lastBlockSynced, err
				}
				if blockReorged == nil {
					blockReorged = prevBlock
				}
				err = s.resetState(blockReorged.BlockNumber)
				if err != nil {
					log.Errorf("networkID: %d, error resetting the state to a previous block. Retrying... Err: %v", s.networkID, err)
					return lastBlockSynced, fmt.Errorf("error resetting the state to a previous block")
				}
				return blockReorged, nil
			}
			// Check reorg again to be sure that the chain has not changed between the previous checkReorg and the call GetRollupInfoByBlockRange
			block, err := s.checkReorg(lastBlockSynced, initBlockReceived)
			if err != nil {
				log.Errorf("networkID: %d, error checking reorgs. Retrying... Err: %v", s.networkID, err)
				return lastBlockSynced, fmt.Errorf("networkID: %d, error checking reorgs", s.networkID)
			}
			if block != nil {
				err = s.resetState(block.BlockNumber)
				if err != nil {
					log.Errorf("networkID: %d, error resetting the state to a previous block. Retrying... Err: %v", s.networkID, err)
					return lastBlockSynced, fmt.Errorf("networkID: %d, error resetting the state to a previous block", s.networkID)
				}
				return block, nil
			}
		}

		err = s.processBlockRange(blocks, order)
		if err != nil {
			return lastBlockSynced, err
		}
		if len(blocks) > 0 {
			lastBlockSynced = &blocks[len(blocks)-1]
			for i := range blocks {
				log.Debug("NetworkID: ", s.networkID, ", Position: ", i, ". BlockNumber: ", blocks[i].BlockNumber, ". BlockHash: ", blocks[i].BlockHash)
			}
		}

		if lastKnownBlock.Cmp(new(big.Int).SetUint64(toBlock)) < 1 { // lastKnownBlock <= toBlock
			if !s.synced {
				log.Infof("NetworkID %d Synced!", s.networkID)
				waitDuration = s.cfg.SyncInterval.Duration
				s.synced = true
				s.chSynced <- s.networkID
			}
			break
		} else if !s.synced {
			fromBlock = toBlock + 1
			toBlock = fromBlock + s.cfg.SyncChunkSize
			log.Debugf("NetworkID: %d, not synced yet. Avoid check the same interval. New interval: from block %d, to block %d", s.networkID, fromBlock, toBlock)
		} else {
			fromBlock = lastBlockSynced.BlockNumber
			toBlock = toBlock + s.cfg.SyncChunkSize
			log.Debugf("NetworkID: %d, synced!. New interval: from block %d, to block %d", s.networkID, fromBlock, toBlock)
		}
	}

	return lastBlockSynced, nil
}

func removeBlockElement(slice []etherman.Block, s int) []etherman.Block {
	ret := make([]etherman.Block, 0)
	ret = append(ret, slice[:s]...)
	return append(ret, slice[s+1:]...)
}

func (s *ClientSynchronizer) processBlockRange(blocks []etherman.Block, order map[common.Hash][]etherman.Order) error {
	// New info has to be included into the db using the state
	var isNewGer bool
	for i := range blocks {
		// Begin db transaction
		dbTx, err := s.storage.BeginDBTransaction(s.ctx)
		if err != nil {
			log.Errorf("networkID: %d, error creating db transaction to store block. BlockNumber: %d. Error: %v",
				s.networkID, blocks[i].BlockNumber, err)
			return err
		}
		// Add block information
		blocks[i].NetworkID = s.networkID
		log.Infof("NetworkID: %d. Syncing block: %d", s.networkID, blocks[i].BlockNumber)
		blockID, err := s.storage.AddBlock(s.ctx, &blocks[i], dbTx)
		if err != nil {
			log.Errorf("networkID: %d, error storing block. BlockNumber: %d, error: %v", s.networkID, blocks[i].BlockNumber, err)
			rollbackErr := s.storage.Rollback(s.ctx, dbTx)
			if rollbackErr != nil {
				log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %d, rollbackErr: %v, err: %s",
					s.networkID, blocks[i].BlockNumber, rollbackErr, err.Error())
				return rollbackErr
			}
			return err
		}
		for _, element := range order[blocks[i].BlockHash] {
			switch element.Name {
			case etherman.GlobalExitRootsOrder:
				isNewGer = true
				err = s.processGlobalExitRoot(blocks[i].GlobalExitRoots[element.Pos], blockID, dbTx)
				if err != nil {
					return err
				}
			case etherman.RemoveL2GEROrder:
				err = s.processRemoveL2GlobalExitRoot(blocks[i].RemoveL2GER[element.Pos], blockID, dbTx)
				if err != nil {
					return err
				}
			case etherman.DepositsOrder:
				err = s.processDeposit(blocks[i].Deposits[element.Pos], blockID, dbTx)
				if err != nil {
					return err
				}
			case etherman.ClaimsOrder:
				err = s.processClaim(blocks[i].Claims[element.Pos], blockID, dbTx)
				if err != nil {
					return err
				}
			case etherman.TokensOrder:
				err = s.processTokenWrapped(blocks[i].Tokens[element.Pos], blockID, dbTx)
				if err != nil {
					return err
				}
			case etherman.VerifyBatchOrder:
				err = s.processVerifyBatch(blocks[i].VerifiedBatches[element.Pos], blockID, dbTx)
				if err != nil {
					return err
				}
			}
		}
		err = s.storage.Commit(s.ctx, dbTx)
		if err != nil {
			log.Errorf("networkID: %d, error committing state to store block. BlockNumber: %d, err: %v",
				s.networkID, blocks[i].BlockNumber, err)
			rollbackErr := s.storage.Rollback(s.ctx, dbTx)
			if rollbackErr != nil {
				log.Errorf("networkID: %d, error rolling back state. BlockNumber: %d, rollbackErr: %v, err: %s",
					s.networkID, blocks[i].BlockNumber, rollbackErr, err.Error())
				return rollbackErr
			}
			return err
		}
	}
	if isNewGer {
		// Send latest GER stored to claimTxManager
		ger, err := s.storage.GetLatestL1SyncedExitRoot(s.ctx, nil)
		if err != nil {
			log.Errorf("networkID: %d, error getting latest GER stored on database. Error: %v", s.networkID, err)
			return err
		}
		if s.l1RollupExitRoot != ger.ExitRoots[1] {
			log.Debugf("Updating ger: %+v", ger)
			s.l1RollupExitRoot = ger.ExitRoots[1]
			for _, ch := range s.chsExitRootEvent {
				ch <- ger
			}
		}
	}
	return nil
}

// This function allows reset the state until an specific ethereum block
func (s *ClientSynchronizer) resetState(blockNumber uint64) error {
	log.Infof("NetworkID: %d. Reverting synchronization to block: %d", s.networkID, blockNumber)
	dbTx, err := s.storage.BeginDBTransaction(s.ctx)
	if err != nil {
		log.Errorf("networkID: %d, Error starting a db transaction to reset the state. Error: %v", s.networkID, err)
		return err
	}
	err = s.storage.Reset(s.ctx, blockNumber, s.networkID, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, error resetting the state. Error: %v", s.networkID, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %d, rollbackErr: %v, error : %s",
				s.networkID, blockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}
	depositCnt, err := s.storage.GetNumberDeposits(s.ctx, s.networkID, blockNumber, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, error getting GetNumberDeposits. Error: %v", s.networkID, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %d, rollbackErr: %v, error : %s",
				s.networkID, blockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}

	err = s.bridgeCtrl.ReorgMT(s.ctx, depositCnt, s.networkID, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, error resetting ReorgMT the state. Error: %v", s.networkID, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %d, rollbackErr: %v, error : %s",
				s.networkID, blockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}
	err = s.storage.Commit(s.ctx, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, error committing the resetted state. Error: %v", s.networkID, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %d, rollbackErr: %v, error : %s",
				s.networkID, blockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}

	return nil
}

/*
This function will check if there is a reorg.
As input param needs the last ethereum block synced. Retrieve the block info from the blockchain
to compare it with the stored info. If hash and hash parent matches, then no reorg is detected and return a nil.
If hash or hash parent don't match, reorg detected and the function will return the block until the sync process
must be reverted. Then, check the previous ethereum block synced, get block info from the blockchain and check
hash and has parent. This operation has to be done until a match is found.
*/
func (s *ClientSynchronizer) checkReorg(latestStoredBlock, syncedBlock *etherman.Block) (*etherman.Block, error) {
	// This function only needs to worry about reorgs if some of the reorganized blocks contained rollup info.
	latestStoredEthBlock := *latestStoredBlock
	reorgedBlock := *latestStoredBlock
	var depth uint64
	block := syncedBlock
	for {
		if block == nil {
			log.Infof("NetworkID: %d, [checkReorg function] Checking Block %d in L1", s.networkID, reorgedBlock.BlockNumber)
			b, err := s.etherMan.HeaderByNumber(s.ctx, new(big.Int).SetUint64(reorgedBlock.BlockNumber))
			if err != nil {
				log.Errorf("networkID: %d, error getting latest block synced from blockchain. Block: %d, error: %v", s.networkID, reorgedBlock.BlockNumber, err)
				return nil, err
			}
			block = &etherman.Block{
				BlockNumber: b.Number.Uint64(),
				BlockHash:   b.Hash(),
				ParentHash:  b.ParentHash,
			}
			if block.BlockNumber != reorgedBlock.BlockNumber {
				err := fmt.Errorf("networkID: %d, wrong ethereum block retrieved from blockchain. Block numbers don't match. BlockNumber stored: %d. BlockNumber retrieved: %d",
					s.networkID, reorgedBlock.BlockNumber, block.BlockNumber)
				log.Error("error: ", err)
				return nil, err
			}
		}

		// Compare hashes
		if (block.BlockHash != reorgedBlock.BlockHash || block.ParentHash != reorgedBlock.ParentHash) && reorgedBlock.BlockNumber > s.genBlockNumber {
			log.Info("NetworkID: ", s.networkID, ", [checkReorg function] => reorgedBlockNumber: ", reorgedBlock.BlockNumber)
			log.Info("NetworkID: ", s.networkID, ", [checkReorg function] => reorgedBlockHash: ", reorgedBlock.BlockHash)
			log.Info("NetworkID: ", s.networkID, ", [checkReorg function] => reorgedBlockHashParent: ", reorgedBlock.ParentHash)
			log.Info("NetworkID: ", s.networkID, ", [checkReorg function] => BlockNumber: ", reorgedBlock.BlockNumber, block.BlockNumber)
			log.Info("NetworkID: ", s.networkID, ", [checkReorg function] => BlockHash: ", block.BlockHash)
			log.Info("NetworkID: ", s.networkID, ", [checkReorg function] => BlockHashParent: ", block.ParentHash)
			depth++
			log.Info("NetworkID: ", s.networkID, ", REORG: Looking for the latest correct ethereum block. Depth: ", depth)
			// Reorg detected. Getting previous block
			lb, err := s.storage.GetPreviousBlock(s.ctx, s.networkID, depth, nil)
			if errors.Is(err, gerror.ErrStorageNotFound) {
				log.Warnf("networkID: %d, error checking reorg: previous block not found in db: %v", s.networkID, err)
				reorgedBlock = etherman.Block{
					BlockNumber: s.genBlockNumber,
				}
				return &reorgedBlock, nil
			} else if err != nil {
				log.Errorf("networkID: %d, error getting previousBlock from db. Error: %v", s.networkID, err)
				return nil, err
			}
			reorgedBlock = *lb
		} else {
			log.Debugf("networkID: %d, checkReorg: Block %d hashOk %t parentHashOk %t", s.networkID, reorgedBlock.BlockNumber, block.BlockHash == reorgedBlock.BlockHash, block.ParentHash == reorgedBlock.ParentHash)
			break
		}
		// This forces to get the block from L1 in the next iteration of the loop
		block = nil
	}
	if latestStoredEthBlock.BlockHash != reorgedBlock.BlockHash {
		latestStoredBlock = &reorgedBlock
		log.Info("NetworkID: ", s.networkID, ", reorg detected in block: ", latestStoredEthBlock.BlockNumber, " last block OK: ", latestStoredBlock.BlockNumber)
		return latestStoredBlock, nil
	}
	log.Debugf("NetworkID: %d, no reorg detected in block: %d. BlockHash: %s", s.networkID, latestStoredEthBlock.BlockNumber, latestStoredEthBlock.BlockHash.String())
	return nil, nil
}

func (s *ClientSynchronizer) processVerifyBatch(verifyBatch etherman.VerifiedBatch, blockID uint64, dbTx pgx.Tx) error {
	if verifyBatch.LocalExitRoot == (common.Hash{}) {
		log.Debugf("networkID: %d, skipping empty local exit root in verifyBatch event. VerifyBatch: %+v", s.networkID, verifyBatch)
		return nil
	}
	var isRollupSyncing bool
	for _, n := range s.allNetworkIDs {
		if verifyBatch.RollupID == n {
			isRollupSyncing = true
		}
	}
	if isRollupSyncing {
		// Just check that the calculated RollupExitRoot is fine
		ok, err := s.storage.CheckIfRootExists(s.ctx, verifyBatch.LocalExitRoot.Bytes(), verifyBatch.RollupID, dbTx)
		if err != nil {
			log.Errorf("networkID: %d, error Checking if root exists. Error: %v", s.networkID, err)
			rollbackErr := s.storage.Rollback(s.ctx, dbTx)
			if rollbackErr != nil {
				log.Errorf("networkID: %d, error rolling back state. BlockNumber: %d, rollbackErr: %v, error : %s",
					s.networkID, verifyBatch.BlockNumber, rollbackErr, err.Error())
				return rollbackErr
			}
			return err
		}
		if !ok {
			log.Errorf("networkID: %d, Root: %s doesn't exist!", s.networkID, verifyBatch.LocalExitRoot.String())
			rollbackErr := s.storage.Rollback(s.ctx, dbTx)
			if rollbackErr != nil {
				log.Errorf("networkID: %d, error rolling back state. BlockNumber: %d, rollbackErr: %v",
					s.networkID, verifyBatch.BlockNumber, rollbackErr)
				return rollbackErr
			}
			return fmt.Errorf("networkID: %d, Root: %s doesn't exist!", s.networkID, verifyBatch.LocalExitRoot.String())
		}
	}
	rollupLeaf := etherman.RollupExitLeaf{
		BlockID:  blockID,
		Leaf:     verifyBatch.LocalExitRoot,
		RollupId: verifyBatch.RollupID,
	}
	// Update rollupExitRoot
	err := s.bridgeCtrl.AddRollupExitLeaf(s.ctx, rollupLeaf, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, error adding rollup exit leaf. Error: %v", s.networkID, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state. BlockNumber: %d, rollbackErr: %v, error : %s",
				s.networkID, verifyBatch.BlockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}
	return nil
}

func (s *ClientSynchronizer) processGlobalExitRoot(globalExitRoot etherman.GlobalExitRoot, blockID uint64, dbTx pgx.Tx) error {
	// Store GlobalExitRoot
	globalExitRoot.BlockID = blockID
	globalExitRoot.NetworkID = s.networkID
	if len(globalExitRoot.ExitRoots) == 2 { //nolint:gomnd
		log.Debugf("networkID: %d, Storing L1 Ger: ", s.networkID, globalExitRoot.GlobalExitRoot)
		// Check if there is some globalExitRoot in L2. If so, it must be incompleted. It must be updated.
		// A race condition between dbTxs (L1 dbTx and L2 dbTxs) is very unlikely because L1 sync takes usually takes more time than L2 sync.
		gers, err := s.storage.GetL2ExitRootsByGER(s.ctx, globalExitRoot.GlobalExitRoot, nil)
		if err != nil && !errors.Is(err, gerror.ErrStorageNotFound) {
			log.Errorf("networkID: %d, error storing the GlobalExitRoot in processGlobalExitRoot. BlockNumber: %d. Error: %v", s.networkID, globalExitRoot.BlockNumber, err)
			rollbackErr := s.storage.Rollback(s.ctx, dbTx)
			if rollbackErr != nil {
				log.Errorf("networkID: %d, error rolling back state. BlockNumber: %d, rollbackErr: %v, error : %s",
					s.networkID, globalExitRoot.BlockNumber, rollbackErr, err.Error())
				return rollbackErr
			}
			return err
		}
		for _, ger := range gers {
			ger.ExitRoots = globalExitRoot.ExitRoots
			err = s.storage.UpdateL2GER(s.ctx, ger, dbTx)
			if err != nil {
				log.Errorf("networkID: %d, error storing the GlobalExitRoot updated in processGlobalExitRoot. BlockNumber: %d. Error: %v", s.networkID, globalExitRoot.BlockNumber, err)
				rollbackErr := s.storage.Rollback(s.ctx, dbTx)
				if rollbackErr != nil {
					log.Errorf("networkID: %d, error rolling back state. BlockNumber: %d, rollbackErr: %v, error : %s",
						s.networkID, globalExitRoot.BlockNumber, rollbackErr, err.Error())
					return rollbackErr
				}
				return err
			}
		}
		err = s.storage.AddGlobalExitRoot(s.ctx, &globalExitRoot, dbTx)
		if err != nil {
			log.Errorf("networkID: %d, error storing the GlobalExitRoot in processGlobalExitRoot. BlockNumber: %d. Error: %v", s.networkID, globalExitRoot.BlockNumber, err)
			rollbackErr := s.storage.Rollback(s.ctx, dbTx)
			if rollbackErr != nil {
				log.Errorf("networkID: %d, error rolling back state. BlockNumber: %d, rollbackErr: %v, error : %s",
					s.networkID, globalExitRoot.BlockNumber, rollbackErr, err.Error())
				return rollbackErr
			}
			return err
		}
	} else if len(globalExitRoot.ExitRoots) == 0 {
		log.Debugf("networkID: %d, Storing L2 Ger: %s", s.networkID, globalExitRoot.GlobalExitRoot)
		// First read the mainnetExitRoot and rollupsExitRoot to store all the information in the db.
		ger, err := s.storage.GetL1ExitRootByGER(s.ctx, globalExitRoot.GlobalExitRoot, nil)
		if errors.Is(err, gerror.ErrStorageNotFound) {
			log.Warnf("networkID: %d, L1Ger entry not found in the database. GER: %s", s.networkID, globalExitRoot.GlobalExitRoot.String())
		} else if err != nil {
			log.Errorf("networkID: %d, error getting the GlobalExitRoot in processGlobalExitRoot. BlockNumber: %d. Error: %v", s.networkID, globalExitRoot.BlockNumber, err)
			rollbackErr := s.storage.Rollback(s.ctx, dbTx)
			if rollbackErr != nil {
				log.Errorf("networkID: %d, error rolling back state. BlockNumber: %d, rollbackErr: %v, error : %s",
					s.networkID, globalExitRoot.BlockNumber, rollbackErr, err.Error())
				return rollbackErr
			}
			return err
		} else {
			globalExitRoot.ExitRoots = ger.ExitRoots
		}
		// Store the GlobalExitRoot
		err = s.storage.AddGlobalExitRoot(s.ctx, &globalExitRoot, dbTx)
		if err != nil {
			log.Errorf("networkID: %d, error storing the GlobalExitRoot in processGlobalExitRoot. BlockNumber: %d. Error: %v", s.networkID, globalExitRoot.BlockNumber, err)
			rollbackErr := s.storage.Rollback(s.ctx, dbTx)
			if rollbackErr != nil {
				log.Errorf("networkID: %d, error rolling back state. BlockNumber: %d, rollbackErr: %v, error : %s",
					s.networkID, globalExitRoot.BlockNumber, rollbackErr, err.Error())
				return rollbackErr
			}
			return err
		}
		log.Infof("networkID: %d, adding L2 ger to the channel. GER: %s", s.networkID, globalExitRoot.GlobalExitRoot.String())
		s.chExitRootEventL2 <- &globalExitRoot
	} else {
		return fmt.Errorf("networkID: %d, error exitRoots have a wrong length. Length: %d", s.networkID, len(globalExitRoot.ExitRoots))
	}
	return nil
}

func (s *ClientSynchronizer) processDeposit(deposit etherman.Deposit, blockID uint64, dbTx pgx.Tx) error {
	deposit.BlockID = blockID
	deposit.NetworkID = s.networkID
	depositID, err := s.storage.AddDeposit(s.ctx, &deposit, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, failed to store new deposit locally, BlockNumber: %d, Deposit: %+v err: %v", s.networkID, deposit.BlockNumber, deposit, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %v, rollbackErr: %v, err: %s",
				s.networkID, deposit.BlockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}

	err = s.bridgeCtrl.AddDeposit(s.ctx, &deposit, depositID, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, failed to store new deposit in the bridge tree, BlockNumber: %d, Deposit: %+v err: %v", s.networkID, deposit.BlockNumber, deposit, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %v, rollbackErr: %v, err: %s",
				s.networkID, deposit.BlockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}
	return nil
}

func (s *ClientSynchronizer) processClaim(claim etherman.Claim, blockID uint64, dbTx pgx.Tx) error {
	claim.BlockID = blockID
	claim.NetworkID = s.networkID
	err := s.storage.AddClaim(s.ctx, &claim, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, error storing new Claim in Block:  %d, Claim: %+v, err: %v", s.networkID, claim.BlockNumber, claim, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %d, rollbackErr: %v, err: %s",
				s.networkID, claim.BlockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}
	return nil
}

func (s *ClientSynchronizer) processTokenWrapped(tokenWrapped etherman.TokenWrapped, blockID uint64, dbTx pgx.Tx) error {
	tokenWrapped.BlockID = blockID
	tokenWrapped.NetworkID = s.networkID
	err := s.storage.AddTokenWrapped(s.ctx, &tokenWrapped, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, error storing new L1 TokenWrapped in Block:  %d, TokenWrapped: %+v, err: %v", s.networkID, tokenWrapped.BlockNumber, tokenWrapped, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %d, rollbackErr: %v, err: %s",
				s.networkID, tokenWrapped.BlockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}
	return nil
}

func (s *ClientSynchronizer) processRemoveL2GlobalExitRoot(ger etherman.GlobalExitRoot, blockID uint64, dbTx pgx.Tx) error {
	ger.BlockID = blockID
	ger.NetworkID = s.networkID
	err := s.storage.AddRemoveL2GER(s.ctx, ger, dbTx)
	if err != nil {
		log.Errorf("networkID: %d, error storing removeL2Ger in Block:  %d, GER: %+v, err: %v", s.networkID, ger.BlockNumber, ger, err)
		rollbackErr := s.storage.Rollback(s.ctx, dbTx)
		if rollbackErr != nil {
			log.Errorf("networkID: %d, error rolling back state to store block. BlockNumber: %d, rollbackErr: %v, err: %s",
				s.networkID, ger.BlockNumber, rollbackErr, err.Error())
			return rollbackErr
		}
		return err
	}
	return nil
}
