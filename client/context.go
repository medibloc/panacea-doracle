package client

import (
	"github.com/medibloc/panacea-doracle/types"
	"path/filepath"
)

type Context struct {
	HomeDir           string
	NodePrivKeyPath   string
	OraclePrivKeyPath string
	OraclePubKeyPath  string
}

func (ctx Context) WithHomeDir(homeDir string) Context {
	if homeDir != "" {
		ctx.HomeDir = homeDir
		ctx.NodePrivKeyPath = filepath.Join(homeDir, types.DefaultNodePrivKeyName)
		ctx.OraclePrivKeyPath = filepath.Join(homeDir, types.DefaultOraclePrivKeyName)
		ctx.OraclePubKeyPath = filepath.Join(homeDir, types.DefaultOraclePubKeyName)
	}
	return ctx
}
