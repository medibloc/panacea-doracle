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
	Subscriber     string `mapstructure:"subscriber"`
}

type PanaceaConfig struct {
	GRPCAddr                string `mapstructure:"grpc-addr"`
	RPCAddr                 string `mapstructure:"rpc-addr"`
	ChainID                 string `mapstructure:"chain-id"`
	DefaultGasLimit         uint64 `mapstructure:"default-gas-limit"`
	DefaultFeeAmount        string `mapstructure:"default-fee-amount"`
	LightClientPrimaryAddr  string `mapstructure:"light-client-primary-addr"`
	LightClientWitnessAddrs []string `mapstructure:"light-client-witness-addrs"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			LogLevel:       "info",
			OracleMnemonic: "",
			ListenAddr:     "127.0.0.1:8080",
		},
		Panacea: PanaceaConfig{
			GRPCAddr:                "127.0.0.1:9090",
			RPCAddr:                 "tcp://127.0.0.1:26657",
			ChainID:                 "panacea-3",
			DefaultGasLimit:         300000,
			DefaultFeeAmount:        "1500000umed",
			LightClientPrimaryAddr:  "tcp://127.0.0.1:26657",
			LightClientWitnessAddrs: []string{"tcp://127.0.0.1:26657"},
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
