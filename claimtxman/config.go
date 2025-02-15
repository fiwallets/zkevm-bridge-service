package claimtxman

import (
	"github.com/0xPolygonHermez/zkevm-node/config/types"
	"github.com/fiwallets/go-ethereum/common"
)

// Config is configuration for L2 claim transaction manager
type Config struct {
	//Enabled whether to enable this module
	Enabled bool `mapstructure:"Enabled"`
	// FrequencyToMonitorTxs frequency of the resending failed txs
	FrequencyToMonitorTxs types.Duration `mapstructure:"FrequencyToMonitorTxs"`
	// PrivateKey defines the key store file that is going
	// to be read in order to provide the private key to sign the claim txs
	PrivateKey types.KeystoreFileConfig `mapstructure:"PrivateKey"`
	// RetryInterval is time between each retry
	RetryInterval types.Duration `mapstructure:"RetryInterval"`
	// RetryNumber is the number of retries before giving up
	RetryNumber int `mapstructure:"RetryNumber"`
	// AuthorizedClaimMessageAddresses are the allowed address to bridge message with autoClaim
	AuthorizedClaimMessageAddresses []common.Address `mapstructure:"AuthorizedClaimMessageAddresses"`
	// Enables the ability to Claim bridges between L2s automatically
	AreClaimsBetweenL2sEnabled bool `mapstructure:"AreClaimsBetweenL2sEnabled"`

	// GroupingClaims is the configuration for grouping claims
	GroupingClaims ConfigGroupingClaims `mapstructure:"GroupingClaims"`
}

type ConfigGroupingClaims struct {
	//Enabled whether to enable this module
	Enabled bool `mapstructure:"Enabled"`
	//FrequencyToProcessCompressedClaims wait time to process compressed claims
	FrequencyToProcessCompressedClaims types.Duration `mapstructure:"FrequencyToProcessCompressedClaims"`
	// TriggerNumberOfClaims is the number of claims to trigger sending the grouped claim tx
	TriggerNumberOfClaims int `mapstructure:"TriggerNumberOfClaims"`
	// MaxNumberOfClaimsPerGroup is the maximum number of claims per group
	MaxNumberOfClaimsPerGroup int `mapstructure:"MaxNumberOfClaimsPerGroup"`
	// TriggerRetainedClaimPeriod is maximum time that a claim can be retainer before creating a group
	TriggerRetainedClaimPeriod types.Duration `mapstructure:"TriggerRetainedClaimPeriod"`
	// MaxRetries is the maximum number of retries to send a compressed claim tx
	MaxRetries int `mapstructure:"MaxRetries"`
	// RetryInterval is time between each retry
	RetryInterval types.Duration `mapstructure:"RetryInterval"`
	// RetryTimeout is the maximum time to wait for a claim tx to be mined
	RetryTimeout types.Duration `mapstructure:"RetryTimeout"`
	// GasOffset is the offset for the gas estimation
	GasOffset uint64 `mapstructure:"GasOffset"`
}
