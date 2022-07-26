package types

import "path/filepath"

var (
	DefaultDataDir    = "data"
	DefaultDoracleDir = ".doracle"
	DefaultConfigDir  = filepath.Join(DefaultDataDir, DefaultDoracleDir)

	DefaultOraclePrivKeyName = "oracle_priv_key.sealed"
	DefaultNodePrivKeyName   = "node_priv_key.sealed"

	DefaultOraclePubKeyName = "oracle_pub_key.json"

	DefaultNodePrivKeyFilePath   = filepath.Join(DefaultConfigDir, DefaultNodePrivKeyName)
	DefaultOraclePrivKeyFilePath = filepath.Join(DefaultConfigDir, DefaultOraclePrivKeyName)
	DefaultOraclePubKeyFilePath  = filepath.Join(DefaultConfigDir, DefaultOraclePubKeyName)
)
