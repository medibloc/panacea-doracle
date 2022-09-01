package cmd

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/client"
	"github.com/medibloc/panacea-doracle/client/flags"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := client.GetContext(cmd)
			if err != nil {
				return err
			}

			conf, err := config.ReadConfigTOML(getConfigPath(ctx.HomeDir))
			if err != nil {
				return fmt.Errorf("failed to read config from file: %w", err)
			}

			if err := initLogger(conf); err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}

			oracleAccount, err := getOracleAccount(cmd, conf.OracleMnemonic)
			if err != nil {
				return fmt.Errorf("failed to get oracle account: %w", err)
			}

			svc, err := service.New(ctx, conf)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			defer svc.Close()

			svc.OracleAccount = oracleAccount

			subscriber, err := event.NewSubscriber(svc)
			if err != nil {
				return err
			}
			defer subscriber.Close()

			if err := subscriber.Run(event.RegisterOracleEvent{}); err != nil {
				return fmt.Errorf("failed to subscribe events: %w", err)
			}

			sigChan := make(chan os.Signal, 1)

			signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
			<-sigChan
			log.Infof("signal detected")

			return nil
		},
	}

	cmd.Flags().Uint32P(flags.FlagAccNum, "a", 0, "Account number of oracle")
	cmd.Flags().Uint32P(flags.FlagIndex, "i", 0, "Address index number for HD derivation of oracle")

	return cmd
}
