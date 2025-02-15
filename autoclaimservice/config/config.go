package config

import (
	"path/filepath"
	"strings"

	"github.com/fiwallets/zkevm-bridge-service/autoclaimservice/autoclaim"
	"github.com/fiwallets/zkevm-bridge-service/autoclaimservice/blockchainmanager"
	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config struct
type Config struct {
	Log               log.Config
	AutoClaim         autoclaim.Config
	BlockchainManager blockchainmanager.Config
}

// Load loads the configuration
func Load(configFilePath string) (*Config, error) {
	cfg, err := Default()
	if err != nil {
		return nil, err
	}

	if configFilePath != "" {
		dirName, fileName := filepath.Split(configFilePath)

		fileExtension := strings.TrimPrefix(filepath.Ext(fileName), ".")
		fileNameWithoutExtension := strings.TrimSuffix(fileName, "."+fileExtension)

		viper.AddConfigPath(dirName)
		viper.SetConfigName(fileNameWithoutExtension)
		viper.SetConfigType(fileExtension)
	}

	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("ZKEVM_AUTOCLAIM")

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Infof("config file not found")
		} else {
			log.Infof("error reading config file: ", err)
			return nil, err
		}
	}

	decodeHooks := []viper.DecoderConfigOption{
		// this allows arrays to be decoded from env var separated by ",", example: MY_VAR="value1,value2,value3"
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(mapstructure.TextUnmarshallerHookFunc(), mapstructure.StringToSliceHookFunc(","))),
	}
	err = viper.Unmarshal(&cfg, decodeHooks...)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
