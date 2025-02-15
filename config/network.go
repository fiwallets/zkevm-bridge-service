package config

import (
	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/fiwallets/go-ethereum/common"
)

// NetworkConfig is the configuration struct for the different environments.
type NetworkConfig struct {
	GenBlockNumber                        uint64
	PolygonBridgeAddress                  common.Address
	PolygonZkEVMGlobalExitRootAddress     common.Address
	PolygonRollupManagerAddress           common.Address
	L2ClaimCompressorAddress              common.Address
	L2PolygonBridgeAddresses              []common.Address
	SovereignChains                       []bool
	L2PolygonZkEVMGlobalExitRootAddresses []common.Address
}

const (
	defaultNetwork = "mainnet"
)

//nolint:gomnd
var (
	networkConfigs = map[string]NetworkConfig{
		defaultNetwork: {
			GenBlockNumber:                        16896718,
			PolygonBridgeAddress:                  common.HexToAddress("0x2a3DD3EB832aF982ec71669E178424b10Dca2EDe"),
			PolygonZkEVMGlobalExitRootAddress:     common.HexToAddress("0x580bda1e7A0CFAe92Fa7F6c20A3794F169CE3CFb"),
			PolygonRollupManagerAddress:           common.HexToAddress("0x0000000000000000000000000000000000000000"),
			L2ClaimCompressorAddress:              common.HexToAddress("0x0000000000000000000000000000000000000000"),
			L2PolygonBridgeAddresses:              []common.Address{common.HexToAddress("0x2a3DD3EB832aF982ec71669E178424b10Dca2EDe")},
			SovereignChains:                       []bool{false},
			L2PolygonZkEVMGlobalExitRootAddresses: []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
		},

		"testnet": {
			GenBlockNumber:                        8572995,
			PolygonBridgeAddress:                  common.HexToAddress("0xF6BEEeBB578e214CA9E23B0e9683454Ff88Ed2A7"),
			PolygonZkEVMGlobalExitRootAddress:     common.HexToAddress("0x4d9427DCA0406358445bC0a8F88C26b704004f74"),
			PolygonRollupManagerAddress:           common.HexToAddress("0x0000000000000000000000000000000000000000"),
			L2ClaimCompressorAddress:              common.HexToAddress("0x0000000000000000000000000000000000000000"),
			L2PolygonBridgeAddresses:              []common.Address{common.HexToAddress("0xF6BEEeBB578e214CA9E23B0e9683454Ff88Ed2A7")},
			SovereignChains:                       []bool{false},
			L2PolygonZkEVMGlobalExitRootAddresses: []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
		},
		"internaltestnet": {
			GenBlockNumber:                        7674349,
			PolygonBridgeAddress:                  common.HexToAddress("0x47c1090bc966280000Fe4356a501f1D0887Ce840"),
			PolygonZkEVMGlobalExitRootAddress:     common.HexToAddress("0xA379Dd55Eb12e8FCdb467A814A15DE2b29677066"),
			PolygonRollupManagerAddress:           common.HexToAddress("0x0000000000000000000000000000000000000000"),
			L2ClaimCompressorAddress:              common.HexToAddress("0x0000000000000000000000000000000000000000"),
			L2PolygonBridgeAddresses:              []common.Address{common.HexToAddress("0xfC5b0c5F677a3f3E29DB2e98c9eD455c7ACfCf03")},
			SovereignChains:                       []bool{false},
			L2PolygonZkEVMGlobalExitRootAddresses: []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000000")},
		},
		"local": {
			GenBlockNumber:                        1,
			PolygonBridgeAddress:                  common.HexToAddress("0xFe12ABaa190Ef0c8638Ee0ba9F828BF41368Ca0E"),
			PolygonZkEVMGlobalExitRootAddress:     common.HexToAddress("0x8A791620dd6260079BF849Dc5567aDC3F2FdC318"),
			PolygonRollupManagerAddress:           common.HexToAddress("0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e"),
			L2ClaimCompressorAddress:              common.HexToAddress("0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6"),
			L2PolygonBridgeAddresses:              []common.Address{common.HexToAddress("0xFe12ABaa190Ef0c8638Ee0ba9F828BF41368Ca0E")},
			SovereignChains:                       []bool{false},
			L2PolygonZkEVMGlobalExitRootAddresses: []common.Address{common.HexToAddress("0xa40d5f56745a118d0906a34e69aec8c0db1cb8fa")},
		},
	}
)

func (cfg *Config) loadNetworkConfig(network string) {
	networkConfig, valid := networkConfigs[network]
	if valid {
		log.Debugf("Network '%v' selected", network)
		cfg.NetworkConfig = networkConfig
	} else {
		log.Debugf("Network '%v' is invalid. Selecting %v instead.", network, defaultNetwork)
		cfg.NetworkConfig = networkConfigs[defaultNetwork]
	}
}
