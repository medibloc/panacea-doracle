package client

import (
	"github.com/medibloc/panacea-doracle/client/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const ContextKey = "client.context"

func GetContext(cmd *cobra.Command) (Context, error) {
	ctx := GetContextFromCmd(cmd)
	return ReadPersistentCommandFlags(ctx, cmd.Flags())
}

func GetContextFromCmd(cmd *cobra.Command) Context {
	if v := cmd.Context().Value(ContextKey); v != nil {
		ctxPtr := v.(*Context)
		return *ctxPtr
	}
	return Context{}
}

func ReadPersistentCommandFlags(ctx Context, flagSet *pflag.FlagSet) (Context, error) {
	if ctx.HomeDir == "" || flagSet.Changed(flags.FlagHome) {
		homeDir, _ := flagSet.GetString(flags.FlagHome)
		ctx = ctx.WithHomeDir(homeDir)
	}

	return ctx, nil
}
