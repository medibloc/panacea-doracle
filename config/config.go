package config

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
	GRPCAddr string `mapstructure:"grpc-addr"`
	WSAddr   string `mapstructure:"websocket-addr"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			LogLevel:       "info",
			OracleMnemonic: "",
			ListenAddr:     "127.0.0.1:8080",
			Subscriber:     "",
		},
		Panacea: PanaceaConfig{
			GRPCAddr: "127.0.0.1:9090",
			WSAddr:   "tcp://127.0.0.1:26657",
		},
	}
}

func (c *Config) validate() error {
	//TODO: validate other configs

	return nil
}
