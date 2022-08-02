package config

type Config struct {
	BaseConfig `mapstructure:",squash"`

	Panacea PanaceaConfig `mapstructure:"panacea"`
	Enclave EnclaveConfig `mapstructure:"enclave"`
}

type BaseConfig struct {
	LogLevel       string `mapstructure:"log-level"`
	OracleMnemonic string `mapstructure:"oracle-mnemonic"`
	ListenAddr     string `mapstructure:"listen_addr"`
}

type PanaceaConfig struct {
	GRPCAddr string `mapstructure:"grpc-addr"`
	WSAddr   string `mapstructure:"websocket-addr"`
}

type EnclaveConfig struct {
	Enable bool `mapstructure:"enable"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			LogLevel:       "info",
			OracleMnemonic: "",
			ListenAddr:     "127.0.0.1:8080",
		},
		Panacea: PanaceaConfig{
			GRPCAddr: "127.0.0.1:9090",
			WSAddr:   "tcp://127.0.0.1:26657",
		},
		Enclave: EnclaveConfig{
			Enable: true,
		},
	}
}

func (c *Config) validate() error {
	//TODO: validate other configs

	return nil
}
