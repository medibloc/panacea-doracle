package config

import sdk "github.com/cosmos/cosmos-sdk/types"

type Config struct {
	BaseConfig `mapstructure:",squash"`

	Panacea PanaceaConfig `mapstructure:"panacea"`
}

type BaseConfig struct {
	LogLevel       string `mapstructure:"log-level"`
	OracleMnemonic string `mapstructure:"oracle-mnemonic"`
	ListenAddr     string `mapstructure:"listen_addr"`
}

type PanaceaConfig struct {
	ChainID          string `mapstructure:"chain-id"`
	GRPCAddr         string `mapstructure:"grpc-addr"`
	DefaultGasLimit  uint64 `mapstructure:"default-gas-limit"`
	DefaultFeeAmount string `mapstructure:"default-fee-amount"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			LogLevel:       "info",
			OracleMnemonic: "",
			ListenAddr:     "127.0.0.1:8080",
		},
		Panacea: PanaceaConfig{
			ChainID:          "",
			GRPCAddr:         "127.0.0.1:9090",
			DefaultGasLimit:  200000,
			DefaultFeeAmount: "1000000umed",
		},
	}
}

func (c *Config) validate() error {
	_, err := sdk.ParseCoinsNormalized(c.Panacea.DefaultFeeAmount)
	if err != nil {
		return err
	}

	return nil
}
