package cmd

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/client/flags"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/server/regoracle"
	"github.com/medibloc/panacea-doracle/server/service"
	"github.com/spf13/cobra"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

func RegisterOracleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-oracle",
		Short: "Register an oracle",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get config
			conf, err := config.ReadConfigTOML(getConfigPath())
			if err != nil {
				log.Errorf("failed to read config.toml: %v", err)
				return err
			}

			if err := initLogger(conf); err != nil {
				log.Errorf("failed to init logger: %v", err)
				return fmt.Errorf("failed to init logger: %w", err)
			}

			svc, err := service.New(cmd, conf)
			if err != nil {
				log.Errorf("failed to create a new service: %v", err)
				return err
			}
			defer svc.Close()

			errChan := make(chan error, 1)

			srv := regoracle.NewServer(svc)
			go func() {
				errChan <- srv.Run()
			}()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)

			select {
			case err := <-errChan:
				if err != nil {
					log.Errorf("error occured while registering oracle: %v", err)
				} else {
					log.Info("Oracle registration success")
				}
			case <-sigChan:
				log.Info("signal detected")
			}

			log.Info("starting the graceful shutdown")

			defer func() {
				log.Info("terminating oracle register process")
				if err := srv.Close(); err != nil {
					log.Infof("error occured while closing the oracle register process: %v", err)
				}
			}()

			return nil
		},
	}

	cmd.Flags().Uint32P(flags.FlagAccNum, "a", 0, "Account number of oracle")
	cmd.Flags().Uint32P(flags.FlagIndex, "i", 0, "Address index number for HD derivation of oracle")
	cmd.Flags().Int64(flags.FlagTrustedBlockHeight, 0, "Trusted block height")
	cmd.Flags().String(flags.FlagTrustedBlockHash, "", "Trusted block hash")
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHeight)
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHash)

	return cmd
}
