package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			// get oracle account from mnemonic.
			oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.AccNum, conf.AccIndex)
			if err != nil {
				return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
			}

			svc, err := service.New(conf, oracleAccount)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			defer svc.Close()

			err = svc.StartSubscriptions(
				event.NewRegisterOracleEvent(svc),
			)
			if err != nil {
				return fmt.Errorf("failed to start event subscription: %w", err)
			}

			sigChan := make(chan os.Signal, 1)

			signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
			<-sigChan
			log.Infof("signal detected")

			return nil
		},
	}

	return cmd
}
