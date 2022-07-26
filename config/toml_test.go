package config_test

import (
	"os"
	"testing"

	"github.com/medibloc/panacea-doracle/config"
	"github.com/stretchr/testify/require"
)

func TestWriteAndReadConfigTOML(t *testing.T) {
	path := "./config.toml"

	err := config.WriteConfigTOML(path, config.DefaultConfig())
	require.NoError(t, err)
	defer os.Remove(path)

	conf, err := config.ReadConfigTOML(path)
	require.NoError(t, err)
	require.EqualValues(t, config.DefaultConfig(), conf)
}
