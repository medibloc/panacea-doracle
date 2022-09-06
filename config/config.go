package config

import (
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Config struct {
	BaseConfig `mapstructure:",squash"`

	Panacea PanaceaConfig `mapstructure:"panacea"`
}

type BaseConfig struct {
	homeDir string // not read from toml file

	LogLevel       string `mapstructure:"log-level"`
	OracleMnemonic string `mapstructure:"oracle-mnemonic"`
	ListenAddr     string `mapstructure:"listen_addr"`
	Subscriber     string `mapstructure:"subscriber"`
	DataDir        string `mapstructure:"data_dir"`

	OraclePrivKeyFile string `mapstructure:"oracle_priv_key_file"`
	OraclePubKeyFile  string `mapstructure:"oracle_pub_key_file"`
	NodePrivKeyFile   string `mapstructure:"node_priv_key_file"`
}

type PanaceaConfig struct {
	GRPCAddr                string   `mapstructure:"grpc-addr"`
	RPCAddr                 string   `mapstructure:"rpc-addr"`
	ChainID                 string   `mapstructure:"chain-id"`
	DefaultGasLimit         uint64   `mapstructure:"default-gas-limit"`
	DefaultFeeAmount        string   `mapstructure:"default-fee-amount"`
	LightClientPrimaryAddr  string   `mapstructure:"light-client-primary-addr"`
	LightClientWitnessAddrs []string `mapstructure:"light-client-witness-addrs"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			homeDir: "",

			LogLevel:       "info",
			OracleMnemonic: "",
			ListenAddr:     "127.0.0.1:8080",
			DataDir:        "data",

			OraclePrivKeyFile: "oracle_priv_key.sealed",
			OraclePubKeyFile:  "oracle_pub_key.json",
			NodePrivKeyFile:   "node_priv_key.sealed",
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

func (c *Config) SetHomeDir(dir string) {
	c.homeDir = dir
}

func (c *Config) AbsDataDirPath() string {
	return rootify(c.DataDir, c.homeDir)
}

func (c *Config) AbsOraclePrivKeyPath() string {
	return rootify(c.OraclePrivKeyFile, c.homeDir)
}

func (c *Config) AbsOraclePubKeyPath() string {
	return rootify(c.OraclePubKeyFile, c.homeDir)
}

func (c *Config) AbsNodePrivKeyPath() string {
	return rootify(c.NodePrivKeyFile, c.homeDir)
}

func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
