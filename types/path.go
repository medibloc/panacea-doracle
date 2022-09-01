package types

import "path/filepath"

func GetNodePrivKeyPath(homeDir string) string {
	return filepath.Join(homeDir, DefaultNodePrivKeyName)
}

func GetOraclePrivKeyPath(homeDir string) string {
	return filepath.Join(homeDir, DefaultOraclePrivKeyName)
}

func GetOraclePubKeyPath(homeDir string) string {
	return filepath.Join(homeDir, DefaultOraclePubKeyName)
}
