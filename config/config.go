package config

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
	GRPCAddr      string   `mapstructure:"grpc-addr"`
	PrimaryAddr   string   `mapstructure:"primary-addr"`
	WitnessesAddr []string `mapstructure:"witnesses-addr"`
	RpcAddr       string   `mapstructure:"rpc-addr"`
	ChainID       string   `mapstructure:"chain-id"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			LogLevel:       "info",
			OracleMnemonic: "",
			ListenAddr:     "127.0.0.1:8080",
		},
		Panacea: PanaceaConfig{
			GRPCAddr:      "127.0.0.1:9090",
			PrimaryAddr:   "https://rpc.gopanacea.org:443",
			WitnessesAddr: []string{"https://rpc.gopanacea.org:443"},
			RpcAddr:       "https://rpc.gopanacea.org:443",
			ChainID:       "panacea-3",
		},
	}
}

func (c *Config) validate() error {
	//TODO: validate other configs

	return nil
}
