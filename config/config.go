package config

type Config struct {
	BaseConfig `mapstructure:",squash"`

	Panacea PanaceaConfig `mapstructure:"panacea"`
	Enclave EnclaveConfig `mapstructure:"enclave"`
}

type BaseConfig struct {
	LogLevel       string `mapstructure:"log-level"`
	OracleMnemonic string `mapstructure:"doracle-mnemonic"`
}

type PanaceaConfig struct {
	GRPCAddr string `mapstructure:"grpc-addr"`
}

type EnclaveConfig struct {
	Enable bool `mapstructure:"enable"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			LogLevel:       "info",
			OracleMnemonic: "",
		},
		Panacea: PanaceaConfig{
			GRPCAddr: "127.0.0.1:9090",
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
