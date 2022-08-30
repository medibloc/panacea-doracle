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
	PrimaryAddr      string `mapstructure:"primary-addr"`
	WitnessesAddr    string `mapstructure:"witnesses-addr"`
	RpcAddr          string `mapstructure:"rpc-addr"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			LogLevel:       "info",
			OracleMnemonic: "",
			ListenAddr:     "127.0.0.1:8080",
		},
		Panacea: PanaceaConfig{
			ChainID:          "panacea-3",
			GRPCAddr:         "https://grpc.gopanacea.org:443",
			DefaultGasLimit:  200000,
			DefaultFeeAmount: "1000000umed",
			PrimaryAddr:      "https://rpc.gopanacea.org:443",
			WitnessesAddr:    "https://rpc.gopanacea.org:443",
			RpcAddr:          "https://rpc.gopanacea.org:443",
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
