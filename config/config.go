package config

import (
	"fmt"
)

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
}

type EnclaveConfig struct {
	Enable                  bool   `mapstructure:"enable"`
	AttestationProviderAddr string `mapstructure:"attestation-provider-addr"`
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
		},
		Enclave: EnclaveConfig{
			Enable:                  true,
			AttestationProviderAddr: "127.0.0.1:9999",
		},
	}
}

func (c *Config) validate() error {
	if c.Enclave.Enable {
		if c.Enclave.AttestationProviderAddr == "" {
			return fmt.Errorf("attestation-provider-addr should be specified if enclave is enabled")
		}
	}

	//TODO: validate other configs

	return nil
}
