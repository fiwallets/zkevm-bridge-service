package server

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/fiwallets/zkevm-bridge-service/bridgectrl"
	"github.com/fiwallets/zkevm-bridge-service/bridgectrl/pb"
	"github.com/fiwallets/zkevm-bridge-service/etherman"
	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/fiwallets/zkevm-bridge-service/utils/gerror"
	"github.com/fiwallets/go-ethereum/common"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/jackc/pgx/v4"
)

type bridgeService struct {
	storage          bridgeServiceStorage
	networkIDs       map[uint32]uint8
	height           uint8
	defaultPageLimit uint32
	maxPageLimit     uint32
	version          string
	cache            *lru.Cache[string, [][]byte]
	pb.UnimplementedBridgeServiceServer
}

// NewBridgeService creates new bridge service.
func NewBridgeService(cfg Config, height uint8, networks []uint32, storage interface{}) *bridgeService {
	var networkIDs = make(map[uint32]uint8)
	for i, network := range networks {
		networkIDs[network] = uint8(i)
	}
	cache, err := lru.New[string, [][]byte](cfg.CacheSize)
	if err != nil {
		panic(err)
	}
	return &bridgeService{
		storage:          storage.(bridgeServiceStorage),
		height:           height,
		networkIDs:       networkIDs,
		defaultPageLimit: cfg.DefaultPageLimit,
		maxPageLimit:     cfg.MaxPageLimit,
		version:          cfg.BridgeVersion,
		cache:            cache,
	}
}

// getNode returns the children hash pairs for a given parent hash.
func (s *bridgeService) getNode(ctx context.Context, parentHash [bridgectrl.KeyLen]byte, dbTx pgx.Tx) (left, right [bridgectrl.KeyLen]byte, err error) {
	value, ok := s.cache.Get(string(parentHash[:]))
	if !ok {
		var err error
		value, err = s.storage.Get(ctx, parentHash[:], dbTx)
		if err != nil {
			return left, right, fmt.Errorf("parentHash: %s, error: %v", common.BytesToHash(parentHash[:]).String(), err)
		}
		s.cache.Add(string(parentHash[:]), value)
	}
	copy(left[:], value[0])
	copy(right[:], value[1])
	return left, right, nil
}

// getProof returns the merkle proof for a given index and root.
func (s *bridgeService) getProof(index uint32, root [bridgectrl.KeyLen]byte, dbTx pgx.Tx) ([][bridgectrl.KeyLen]byte, error) {
	var siblings [][bridgectrl.KeyLen]byte

	cur := root
	ctx := context.Background()
	// It starts in height-1 because 0 is the level of the leafs
	for h := int(s.height - 1); h >= 0; h-- {
		left, right, err := s.getNode(ctx, cur, dbTx)
		if err != nil {
			return nil, fmt.Errorf("height: %d, cur: %s, error: %v", h, common.BytesToHash(cur[:]).String(), err)
		}
		/*
					*        Root                (level h=3 => height=4)
					*      /     \
					*	 O5       O6             (level h=2)
					*	/ \      / \
					*  O1  O2   O3  O4           (level h=1)
			        *  /\   /\   /\ /\
					* 0  1 2  3 4 5 6 7 Leafs    (level h=0)
					* Example 1:
					* Choose index = 3 => 011 binary
					* Assuming we are in level 1 => h=1; 1<<h = 010 binary
					* Now, let's do AND operation => 011&010=010 which is higher than 0 so we need the left sibling (O1)
					* Example 2:
					* Choose index = 4 => 100 binary
					* Assuming we are in level 1 => h=1; 1<<h = 010 binary
					* Now, let's do AND operation => 100&010=000 which is not higher than 0 so we need the right sibling (O4)
					* Example 3:
					* Choose index = 4 => 100 binary
					* Assuming we are in level 2 => h=2; 1<<h = 100 binary
					* Now, let's do AND operation => 100&100=100 which is higher than 0 so we need the left sibling (O5)
		*/

		if index&(1<<h) > 0 {
			siblings = append(siblings, left)
			cur = right
		} else {
			siblings = append(siblings, right)
			cur = left
		}
	}

	// We need to invert the siblings to go from leafs to the top
	for st, en := 0, len(siblings)-1; st < en; st, en = st+1, en-1 {
		siblings[st], siblings[en] = siblings[en], siblings[st]
	}

	return siblings, nil
}

// getRollupExitProof returns the merkle proof for the zkevm leaf.
func (s *bridgeService) getRollupExitProof(rollupIndex uint32, root common.Hash, dbTx pgx.Tx) ([][bridgectrl.KeyLen]byte, common.Hash, error) {
	ctx := context.Background()

	// Get leaves given the root
	leaves, err := s.storage.GetRollupExitLeavesByRoot(ctx, root, dbTx)
	if err != nil {
		err = fmt.Errorf("error getting leaves by ger: %s, error: %w", root.String(), err)
		return nil, common.Hash{}, err
	}
	// Compute Siblings
	var ls [][bridgectrl.KeyLen]byte
	for _, l := range leaves {
		var aux [bridgectrl.KeyLen]byte
		copy(aux[:], l.Leaf.Bytes())
		ls = append(ls, aux)
	}
	siblings, r, err := bridgectrl.ComputeSiblings(rollupIndex, ls, s.height)
	if err != nil {
		return nil, common.Hash{}, err
	} else if root != r {
		log.Warnf("error checking calculated root: required: %s, calculated:%s", root.String(), r.String())
		return nil, common.Hash{}, fmt.Errorf("error checking calculated root: required:%s, calculated: %s", root.String(), r.String())
	}
	if len(siblings) == 0 || len(ls) == 0 {
		return nil, common.Hash{}, fmt.Errorf("no siblings found for root: %s", root.String())
	}
	if len(ls) <= int(rollupIndex) {
		return siblings, common.Hash{}, fmt.Errorf("error getting rollupLeaf. Not synced yet")
	}
	return siblings, ls[rollupIndex], nil
}

// GetClaimProof returns the merkle proof to claim the given deposit.
func (s *bridgeService) GetClaimProof(depositCnt, networkID uint32, dbTx pgx.Tx) (*etherman.GlobalExitRoot, [][bridgectrl.KeyLen]byte, [][bridgectrl.KeyLen]byte, error) {
	ctx := context.Background()

	deposit, err := s.storage.GetDeposit(ctx, depositCnt, networkID, dbTx)
	if err != nil {
		return nil, nil, nil, err
	}

	if !deposit.ReadyForClaim {
		return nil, nil, nil, gerror.ErrDepositNotSynced
	}

	globalExitRoot, err := s.storage.GetLatestExitRoot(ctx, networkID, deposit.DestinationNetwork, dbTx)
	if err != nil {
		return nil, nil, nil, err
	}

	var (
		merkleProof       [][bridgectrl.KeyLen]byte
		rollupMerkleProof [][bridgectrl.KeyLen]byte
		rollupLeaf        common.Hash
	)
	if networkID == 0 { // Mainnet
		merkleProof, err = s.getProof(depositCnt, globalExitRoot.ExitRoots[0], dbTx)
		if err != nil {
			log.Error("error getting merkleProof. Error: ", err)
			return nil, nil, nil, fmt.Errorf("getting the proof failed, error: %v, network: %d", err, networkID)
		}
		rollupMerkleProof = emptyProof()
	} else { // Rollup
		rollupMerkleProof, rollupLeaf, err = s.getRollupExitProof(networkID-1, globalExitRoot.ExitRoots[1], dbTx)
		if err != nil {
			log.Error("error getting rollupProof. Error: ", err)
			return nil, nil, nil, fmt.Errorf("getting the rollup proof failed, error: %v, network: %d", err, networkID)
		}
		merkleProof, err = s.getProof(depositCnt, rollupLeaf, dbTx)
		if err != nil {
			log.Error("error getting merkleProof. Error: ", err)
			return nil, nil, nil, fmt.Errorf("getting the proof failed, error: %v, network: %d", err, networkID)
		}
	}

	return globalExitRoot, merkleProof, rollupMerkleProof, nil
}

// GetClaimProofbyGER returns the merkle proof to claim the given deposit.
func (s *bridgeService) GetClaimProofbyGER(depositCnt, networkID uint32, GER common.Hash, dbTx pgx.Tx) (*etherman.GlobalExitRoot, [][bridgectrl.KeyLen]byte, [][bridgectrl.KeyLen]byte, error) {
	ctx := context.Background()

	if dbTx == nil { // if the call comes from the rest API
		deposit, err := s.storage.GetDeposit(ctx, depositCnt, networkID, nil)
		if err != nil {
			err = fmt.Errorf("error getting deposit %d for network: %d. Err: %w", depositCnt, networkID, err)
			return nil, nil, nil, err
		}

		if !deposit.ReadyForClaim {
			log.Warnf("Deposit not ready for claim. Deposit: %d, Network: %d", depositCnt, networkID)
			//return nil, nil, nil, gerror.ErrDepositNotSynced
		}
	}

	globalExitRoot, err := s.storage.GetL1ExitRootByGER(ctx, GER, dbTx)
	if err != nil {
		err = fmt.Errorf("error getting GlobalExitRoot data for GER: %s. Err: %w", GER.String(), err)
		return nil, nil, nil, err
	}

	var (
		merkleProof       [][bridgectrl.KeyLen]byte
		rollupMerkleProof [][bridgectrl.KeyLen]byte
		rollupLeaf        common.Hash
	)
	if networkID == 0 { // Mainnet
		merkleProof, err = s.getProof(depositCnt, globalExitRoot.ExitRoots[0], dbTx)
		if err != nil {
			log.Errorf("error getting merkleProof. Error: %w", err)
			return nil, nil, nil, fmt.Errorf("getting the proof failed (MAINNET), error: %v, network: %d", err, networkID)
		}
		rollupMerkleProof = emptyProof()
	} else { // Rollup
		rollupMerkleProof, rollupLeaf, err = s.getRollupExitProof(networkID-1, globalExitRoot.ExitRoots[1], dbTx)
		if err != nil {
			log.Errorf("error getting rollupProof. Error: %w", err)
			return nil, nil, nil, fmt.Errorf("getting the rollupexit proof failed, error: %v, network: %d", err, networkID)
		}
		merkleProof, err = s.getProof(depositCnt, rollupLeaf, dbTx)
		if err != nil {
			log.Errorf("error getting merkleProof. Error: %w", err)
			return nil, nil, nil, fmt.Errorf("getting the proof failed (ROLLUP), error: %v, network: %d", err, networkID)
		}
	}

	return globalExitRoot, merkleProof, rollupMerkleProof, nil
}

// GetClaimProofForCompressed returns the merkle proof to claim the given deposit.
func (s *bridgeService) GetClaimProofForCompressed(ger common.Hash, depositCnt, networkID uint32, dbTx pgx.Tx) (*etherman.GlobalExitRoot, [][bridgectrl.KeyLen]byte, [][bridgectrl.KeyLen]byte, error) {
	ctx := context.Background()

	if dbTx == nil { // if the call comes from the rest API
		deposit, err := s.storage.GetDeposit(ctx, depositCnt, networkID, nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if !deposit.ReadyForClaim {
			return nil, nil, nil, gerror.ErrDepositNotSynced
		}
	}

	globalExitRoot, err := s.storage.GetL1ExitRootByGER(ctx, ger, dbTx)
	if err != nil {
		return nil, nil, nil, err
	}

	var (
		merkleProof       [][bridgectrl.KeyLen]byte
		rollupMerkleProof [][bridgectrl.KeyLen]byte
		rollupLeaf        common.Hash
	)
	if networkID == 0 { // Mainnet
		merkleProof, err = s.getProof(depositCnt, globalExitRoot.ExitRoots[0], dbTx)
		if err != nil {
			log.Error("error getting merkleProof. Error: ", err)
			return nil, nil, nil, fmt.Errorf("getting the proof failed, error: %v, network: %d", err, networkID)
		}
		rollupMerkleProof = emptyProof()
	} else { // Rollup
		rollupMerkleProof, rollupLeaf, err = s.getRollupExitProof(networkID-1, globalExitRoot.ExitRoots[1], dbTx)
		if err != nil {
			log.Error("error getting rollupProof. Error: ", err)
			return nil, nil, nil, fmt.Errorf("getting the rollup proof failed, error: %v, network: %d", err, networkID)
		}
		merkleProof, err = s.getProof(depositCnt, rollupLeaf, dbTx)
		if err != nil {
			log.Error("error getting merkleProof. Error: ", err)
			return nil, nil, nil, fmt.Errorf("getting the proof failed, error: %v, network: %d", err, networkID)
		}
	}

	return globalExitRoot, merkleProof, rollupMerkleProof, nil
}

func emptyProof() [][bridgectrl.KeyLen]byte {
	var proof [][bridgectrl.KeyLen]byte
	for i := 0; i < 32; i++ {
		proof = append(proof, common.Hash{})
	}
	return proof
}

// GetDepositStatus returns deposit with ready_for_claim status.
func (s *bridgeService) GetDepositStatus(ctx context.Context, depositCount, originNetworkID, destNetworkID uint32) (string, error) {
	var (
		claimTxHash string
	)
	// Get the claim tx hash
	claim, err := s.storage.GetClaim(ctx, depositCount, originNetworkID, destNetworkID, nil)
	if err != nil {
		if err != gerror.ErrStorageNotFound {
			return "", err
		}
	} else {
		claimTxHash = claim.TxHash.String()
	}
	return claimTxHash, nil
}

// CheckAPI returns api version.
// Bridge rest API endpoint
func (s *bridgeService) CheckAPI(ctx context.Context, req *pb.CheckAPIRequest) (*pb.CheckAPIResponse, error) {
	return &pb.CheckAPIResponse{
		Api: s.version,
	}, nil
}

// GetBridges returns bridges for the destination address both in L1 and L2.
// Bridge rest API endpoint
func (s *bridgeService) GetBridges(ctx context.Context, req *pb.GetBridgesRequest) (*pb.GetBridgesResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = s.defaultPageLimit
	}
	if limit > s.maxPageLimit {
		limit = s.maxPageLimit
	}
	totalCount, err := s.storage.GetDepositCount(ctx, req.DestAddr, nil)
	if err != nil {
		return nil, err
	}
	deposits, err := s.storage.GetDeposits(ctx, req.DestAddr, limit, req.Offset, nil)
	if err != nil {
		return nil, err
	}

	var pbDeposits []*pb.Deposit
	for _, deposit := range deposits {
		claimTxHash, err := s.GetDepositStatus(ctx, deposit.DepositCount, deposit.NetworkID, deposit.DestinationNetwork)
		if err != nil {
			return nil, err
		}
		mainnetFlag := deposit.NetworkID == 0
		var rollupIndex uint32
		if !mainnetFlag {
			rollupIndex = deposit.NetworkID - 1
		}
		localExitRootIndex := deposit.DepositCount
		pbDeposits = append(
			pbDeposits, &pb.Deposit{
				LeafType:      uint32(deposit.LeafType),
				OrigNet:       deposit.OriginalNetwork,
				OrigAddr:      deposit.OriginalAddress.Hex(),
				Amount:        deposit.Amount.String(),
				DestNet:       deposit.DestinationNetwork,
				DestAddr:      deposit.DestinationAddress.Hex(),
				BlockNum:      deposit.BlockNumber,
				DepositCnt:    deposit.DepositCount,
				NetworkId:     deposit.NetworkID,
				TxHash:        deposit.TxHash.String(),
				ClaimTxHash:   claimTxHash,
				Metadata:      "0x" + hex.EncodeToString(deposit.Metadata),
				ReadyForClaim: deposit.ReadyForClaim,
				GlobalIndex:   etherman.GenerateGlobalIndex(mainnetFlag, rollupIndex, localExitRootIndex).String(),
			},
		)
	}

	return &pb.GetBridgesResponse{
		Deposits: pbDeposits,
		TotalCnt: totalCount,
	}, nil
}

// GetClaims returns claims for the specific smart contract address both in L1 and L2.
// Bridge rest API endpoint
func (s *bridgeService) GetClaims(ctx context.Context, req *pb.GetClaimsRequest) (*pb.GetClaimsResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = s.defaultPageLimit
	}
	if limit > s.maxPageLimit {
		limit = s.maxPageLimit
	}
	totalCount, err := s.storage.GetClaimCount(ctx, req.DestAddr, nil)
	if err != nil {
		return nil, err
	}
	claims, err := s.storage.GetClaims(ctx, req.DestAddr, limit, req.Offset, nil) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	var pbClaims []*pb.Claim
	for _, claim := range claims {
		pbClaims = append(pbClaims, &pb.Claim{
			Index:       claim.Index,
			OrigNet:     claim.OriginalNetwork,
			OrigAddr:    claim.OriginalAddress.Hex(),
			Amount:      claim.Amount.String(),
			NetworkId:   claim.NetworkID,
			DestAddr:    claim.DestinationAddress.Hex(),
			BlockNum:    claim.BlockNumber,
			TxHash:      claim.TxHash.String(),
			RollupIndex: claim.RollupIndex,
			MainnetFlag: claim.MainnetFlag,
		})
	}

	return &pb.GetClaimsResponse{
		Claims:   pbClaims,
		TotalCnt: totalCount,
	}, nil
}

// GetProof returns the merkle proof for the given deposit.
// Bridge rest API endpoint
func (s *bridgeService) GetProof(ctx context.Context, req *pb.GetProofRequest) (*pb.GetProofResponse, error) {
	globalExitRoot, merkleProof, rollupMerkleProof, err := s.GetClaimProof(req.DepositCnt, req.NetId, nil)
	if err != nil {
		return nil, err
	}
	var (
		proof       []string
		rollupProof []string
	)
	if len(proof) != len(rollupProof) {
		return nil, fmt.Errorf("proofs have different lengths. MerkleProof: %d. RollupMerkleProof: %d", len(merkleProof), len(rollupMerkleProof))
	}
	for i := 0; i < len(merkleProof); i++ {
		proof = append(proof, "0x"+hex.EncodeToString(merkleProof[i][:]))
		rollupProof = append(rollupProof, "0x"+hex.EncodeToString(rollupMerkleProof[i][:]))
	}

	return &pb.GetProofResponse{
		Proof: &pb.Proof{
			RollupMerkleProof: rollupProof,
			MerkleProof:       proof,
			MainExitRoot:      globalExitRoot.ExitRoots[0].Hex(),
			RollupExitRoot:    globalExitRoot.ExitRoots[1].Hex(),
		},
	}, nil
}

// GetBridge returns the bridge  with status whether it is able to send a claim transaction or not.
// Bridge rest API endpoint
func (s *bridgeService) GetBridge(ctx context.Context, req *pb.GetBridgeRequest) (*pb.GetBridgeResponse, error) {
	deposit, err := s.storage.GetDeposit(ctx, req.DepositCnt, req.NetId, nil)
	if err != nil {
		return nil, err
	}

	claimTxHash, err := s.GetDepositStatus(ctx, req.DepositCnt, deposit.NetworkID, deposit.DestinationNetwork)
	if err != nil {
		return nil, err
	}
	mainnetFlag := deposit.NetworkID == 0
	var rollupIndex uint32
	if !mainnetFlag {
		rollupIndex = deposit.NetworkID - 1
	}
	localExitRootIndex := deposit.DepositCount

	return &pb.GetBridgeResponse{
		Deposit: &pb.Deposit{
			LeafType:      uint32(deposit.LeafType),
			OrigNet:       deposit.OriginalNetwork,
			OrigAddr:      deposit.OriginalAddress.Hex(),
			Amount:        deposit.Amount.String(),
			DestNet:       deposit.DestinationNetwork,
			DestAddr:      deposit.DestinationAddress.Hex(),
			BlockNum:      deposit.BlockNumber,
			DepositCnt:    deposit.DepositCount,
			NetworkId:     deposit.NetworkID,
			TxHash:        deposit.TxHash.String(),
			ClaimTxHash:   claimTxHash,
			Metadata:      "0x" + hex.EncodeToString(deposit.Metadata),
			ReadyForClaim: deposit.ReadyForClaim,
			GlobalIndex:   etherman.GenerateGlobalIndex(mainnetFlag, rollupIndex, localExitRootIndex).String(),
		},
	}, nil
}

// GetTokenWrapped returns the token wrapped created for a specific network
// Bridge rest API endpoint
func (s *bridgeService) GetTokenWrapped(ctx context.Context, req *pb.GetTokenWrappedRequest) (*pb.GetTokenWrappedResponse, error) {
	tokenWrapped, err := s.storage.GetTokenWrapped(ctx, req.OrigNet, common.HexToAddress(req.OrigTokenAddr), nil)
	if err != nil {
		return nil, err
	}
	return &pb.GetTokenWrappedResponse{
		Tokenwrapped: &pb.TokenWrapped{
			OrigNet:           uint32(tokenWrapped.OriginalNetwork),
			OriginalTokenAddr: tokenWrapped.OriginalTokenAddress.Hex(),
			WrappedTokenAddr:  tokenWrapped.WrappedTokenAddress.Hex(),
			NetworkId:         uint32(tokenWrapped.NetworkID),
			Name:              tokenWrapped.Name,
			Symbol:            tokenWrapped.Symbol,
			Decimals:          uint32(tokenWrapped.Decimals),
		},
	}, nil
}

func (s *bridgeService) GetProofByGER(ctx context.Context, req *pb.GetProofByGERRequest) (*pb.GetProofResponse, error) {
	ger := common.HexToHash(req.Ger)
	globalExitRoot, merkleProof, rollupMerkleProof, err := s.GetClaimProofbyGER(req.DepositCnt, req.NetId, ger, nil)
	if err != nil {
		return nil, err
	}
	var (
		proof       []string
		rollupProof []string
	)
	if len(proof) != len(rollupProof) {
		return nil, fmt.Errorf("proofs have different lengths. MerkleProof: %d. RollupMerkleProof: %d", len(merkleProof), len(rollupMerkleProof))
	}
	for i := 0; i < len(merkleProof); i++ {
		proof = append(proof, "0x"+hex.EncodeToString(merkleProof[i][:]))
		rollupProof = append(rollupProof, "0x"+hex.EncodeToString(rollupMerkleProof[i][:]))
	}

	return &pb.GetProofResponse{
		Proof: &pb.Proof{
			RollupMerkleProof: rollupProof,
			MerkleProof:       proof,
			MainExitRoot:      globalExitRoot.ExitRoots[0].Hex(),
			RollupExitRoot:    globalExitRoot.ExitRoots[1].Hex(),
		},
	}, nil
}

// GetPendingBridgesToClaim returns the pending bridges to claim by destination address, destination network and leaf type in L1 and L2's.
// Bridge rest API endpoint
func (s *bridgeService) GetPendingBridgesToClaim(ctx context.Context, req *pb.GetPendingBridgesRequest) (*pb.GetBridgesResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = s.defaultPageLimit
	}
	if limit > s.maxPageLimit {
		limit = s.maxPageLimit
	}
	destAddr := common.HexToAddress(req.DestAddr)
	deposits, totalDeposits, err := s.storage.GetPendingDepositsToClaim(ctx, destAddr, req.DestNet, req.LeafType, limit, req.Offset, nil)
	if err != nil {
		return nil, err
	}

	var pbDeposits []*pb.Deposit
	for _, deposit := range deposits {
		mainnetFlag := deposit.NetworkID == 0
		var rollupIndex uint32
		if !mainnetFlag {
			rollupIndex = deposit.NetworkID - 1
		}
		localExitRootIndex := deposit.DepositCount
		pbDeposits = append(
			pbDeposits, &pb.Deposit{
				LeafType:      uint32(deposit.LeafType),
				OrigNet:       deposit.OriginalNetwork,
				OrigAddr:      deposit.OriginalAddress.Hex(),
				Amount:        deposit.Amount.String(),
				DestNet:       deposit.DestinationNetwork,
				DestAddr:      deposit.DestinationAddress.Hex(),
				BlockNum:      deposit.BlockNumber,
				DepositCnt:    deposit.DepositCount,
				NetworkId:     deposit.NetworkID,
				TxHash:        deposit.TxHash.String(),
				ClaimTxHash:   "",
				Metadata:      "0x" + hex.EncodeToString(deposit.Metadata),
				ReadyForClaim: deposit.ReadyForClaim,
				GlobalIndex:   etherman.GenerateGlobalIndex(mainnetFlag, rollupIndex, localExitRootIndex).String(),
			},
		)
	}

	return &pb.GetBridgesResponse{
		Deposits: pbDeposits,
		TotalCnt: totalDeposits,
	}, nil
}
