package config

import (
	"bytes"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// DefaultValues is the default configuration
const DefaultValues = `
[Log]
Level = "debug"
Outputs = ["stdout"]

[SyncDB]
Database = "postgres"
User = "test_user"
Password = "test_password"
Name = "test_db"
Host = "zkevm-bridge-db"
Port = "5432"
MaxConns = 20

[ClaimTxManager]
Enabled = false
FrequencyToMonitorTxs = "1s"
PrivateKey = {Path = "./test/test.keystore", Password = "testonly"}
RetryInterval = "1s"
RetryNumber = 10
AuthorizedClaimMessageAddresses = []
AreClaimsBetweenL2sEnabled = false
[ClaimTxManager.GroupingClaims]
    Enabled = false
    FrequencyToProcessCompressedClaims = "10m"
    TriggerNumberOfClaims = 10
    MaxNumberOfClaimsPerGroup = 10
    TriggerRetainedClaimPeriod = "30s"
    MaxRetries = 2
    RetryInterval = "10s"
    RetryTimeout = "30s"
    GasOffset = 0


[Etherman]
L1URL = "http://localhost:8545"
L2URLs = [""]

[Synchronizer]
SyncInterval = "2s"
SyncChunkSize = 100

[BridgeController]
Store = "postgres"
Height = 32

[BridgeServer]
GRPCPort = "9090"
HTTPPort = "8080"
DefaultPageLimit = 25
CacheSize = 100000
MaxPageLimit = 100
BridgeVersion = "v1"
    [BridgeServer.DB]
    Database = "postgres"
    User = "test_user"
    Password = "test_password"
    Name = "test_db"
    Host = "zkevm-bridge-db"
    Port = "5432"
    MaxConns = 20
`

// Default parses the default configuration values.
func Default() (*Config, error) {
	var cfg Config
	viper.SetConfigType("toml")

	err := viper.ReadConfig(bytes.NewBuffer([]byte(DefaultValues)))
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
