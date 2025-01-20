package claimtxman_test

import (
	"bytes"
	"context"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0xPolygonHermez/zkevm-bridge-service/etherman/smartcontracts/claimcompressor"
	zkevmtypes "github.com/0xPolygonHermez/zkevm-node/config/types"
	"github.com/fiwallets/go-ethereum/accounts/abi"
	"github.com/fiwallets/go-ethereum/accounts/abi/bind"
	"github.com/fiwallets/go-ethereum/accounts/keystore"
	"github.com/fiwallets/go-ethereum/common"
	"github.com/fiwallets/go-ethereum/core/types"
	"github.com/fiwallets/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

const (
	ClaimCompressorAddress = "0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6"
	ChainID                = 1001
)

func getSignerFromKeystore(ks zkevmtypes.KeystoreFileConfig, chainID uint64) (*bind.TransactOpts, error) {
	keystoreEncrypted, err := os.ReadFile(filepath.Clean(ks.Path))
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(keystoreEncrypted, ks.Password)
	if err != nil {
		return nil, err
	}
	chainIDBigInt := big.NewInt(ChainID)
	if err != nil {
		return nil, err
	}
	return bind.NewKeyedTransactorWithChainID(key.PrivateKey, chainIDBigInt)
}

const (
	exampleMonitoredTxsData = "ccaa2d113483fbcb236c39345503fa907a171f69c06b47f48e3bc71a664301b0bc9cdf7bad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5b4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d3021ddb9a356815c3fac1026b6dec5df3124afbadb485c9ba5a3e3398a04b7ba85e58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a193440eb01ebfc9ed27500cd4dfc979272d1f0913cc9f66540d7e8005811109e1cf2d887c22bd8750d34016ac3c66b5ff102dacdd73f6b014e710b51e8022af9a1968ffd70157e48063fc33c97a050f7f640233bf646cc98d9524c6b92bcf3ab56f839867cc5f7f196b93bae1e27e6320742445d290f2263827498b54fec539f756afcefad4e508c098b9a7e1d8feb19955fb02ba9675585078710969d3440f5054e0f9dc3e7fe016e050eff260334f18a5d4fe391d82092319f5964f2e2eb7c1c3a5f8b13a49e282f609c317a833fb8d976d11517c571d1221a265d25af778ecf8923490c6ceeb450aecdc82e28293031d10c7d73bf85e57bf041a97360aa2c5d99cc1df82d9c4b87413eae2ef048f94b4d3554cea73d92b0f7af96e0271c691e2bb5c67add7c6caf302256adedf7ab114da0acfe870d449a3a489f781d659e8beccda7bce9f4e8618b6bd2f4132ce798cdc7a60e7e1460a7299e3c6342a579626d22733e50f526ec2fa19a22b31e8ed50f23cd1fdf94c9154ed3a7609a2f1ff981fe1d3b5c807b281e4683cc6d6315cf95b9ade8641defcb32372f1c126e398ef7a5a2dce0a8a7f68bb74560f8f71837c2c2ebbcbf7fffb42ae1896f13f7c7479a0b46a28b6f55540f89444f63de0378e3d121be09e06cc9ded1c20e65876d36aa0c65e9645644786b620e2dd2ad648ddfcbf4a7e5b1a3a4ecfe7f64667a3f0b7e2f4418588ed35a2458cffeb39b93d26f18d2ab13bdce6aee58e7b99359ec2dfd95a9c16dc00d6ef18b7933a6f8dc65ccb55667138776f7dea101070dc8796e3774df84f40ae0c8229d0d6069e5c8f39a7c299677a09d367fc7b05e3bc380ee652cdc72595f74c7b1043d0e1ffbab734648c838dfb0527d971b602bc216c9619ef0abf5ac974a1ed57f4050aa510dd9c74f508277b39d7973bb2dfccc5eeb0618db8cd74046ff337f0a7bf2c8e03e10f642c1886798d71806ab1e888d9e5ee87d0838c5655cb21c6cb83313b5a631175dff4963772cce9108188b34ac87c81c41e662ee4dd2dd7b2bc707961b1e646c4047669dcb6584f0d8d770daf5d7e7deb2e388ab20e2573d171a88108e79d820e98f26c0b84aa8b2f4aa4968dbb818ea32293237c50ba75ee485f4c22adf2f741400bdf8d6a9cc7df7ecae576221665d7358448818bb4ae4562849e949e17ac16e0be16688e156b5cf15e098c627c0056a9000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000fa8cb5cf94a1975aa92ca5e860141ccd8e30b556801539cd62e563bd0e29d24827ae5ba08d7291c96c8cbddcc148bf48a6d68c7974b94356f53754ef6171d757000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266000000000000000000000000000000000000000000000000013fbe85edc9000000000000000000000000000000000000000000000000000000000000000009200000000000000000000000000000000000000000000000000000000000000000"
)

func TestExploratory(t *testing.T) {
	t.Skip("this is a Exploratory test")
	//ctx := context.Background()
	client, err := ethclient.Dial("http://localhost:8123")
	require.NoError(t, err)
	addr := common.HexToAddress(ClaimCompressorAddress)
	claimCompressor, err := claimcompressor.NewClaimcompressor(addr, client)
	require.NoError(t, err)
	smcAbi, err := abi.JSON(strings.NewReader(claimcompressor.ClaimcompressorABI))
	method, ok := smcAbi.Methods["compressClaimCall"]
	require.True(t, ok)
	mainet := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	rollup := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	txData := common.Hex2Bytes(exampleMonitoredTxsData)
	//txCompressData := make([]byte, 4+32*2+len(txData)-4)
	var buf bytes.Buffer
	buf.Write(method.ID[0:4])
	buf.Write(mainet[:])
	buf.Write(rollup[:])
	buf.Write(txData[4:])

	//method.Inputs.Pack(mainet, rollup, []ClaimCompressor.ClaimCompressorCompressClaimCallData{})
	require.NoError(t, err)

	auth, err := getSignerFromKeystore(zkevmtypes.KeystoreFileConfig{
		Path:     "./test/test.keystore.claimtx",
		Password: "testonly",
	}, ChainID)
	require.NoError(t, err)
	toAddr := common.HexToAddress(ClaimCompressorAddress)
	fromAddr := auth.From
	Nonce := uint64(1)
	gas := uint64(1000000)
	gasPrice := big.NewInt(1000000000)
	tx := types.NewTx(&types.LegacyTx{
		To:       &toAddr,
		Nonce:    Nonce,
		Value:    nil,
		Data:     buf.Bytes(),
		Gas:      gas,
		GasPrice: gasPrice,
	})
	var signedTx *types.Transaction
	// sign tx
	signedTx, err = auth.Signer(fromAddr, tx)
	require.NoError(t, err)
	// send tx
	err = client.SendTransaction(context.Background(), signedTx)
	require.NoError(t, err)

	compressClaimCalldata := []claimcompressor.ClaimCompressorCompressClaimCallData{}
	bindcall := &bind.CallOpts{Pending: false}
	res, err := claimCompressor.CompressClaimCall(bindcall, mainet, rollup, compressClaimCalldata)
	require.NoError(t, err)
	t.Log(res)
}
