package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/medibloc/panacea-doracle/server"
	"github.com/medibloc/panacea-doracle/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	oracleevent "github.com/medibloc/panacea-doracle/event/oracle"
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

			svc, err := service.New(conf)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			defer svc.Close()

			err = svc.StartSubscriptions(
				oracleevent.NewRegisterOracleEvent(svc),
				oracleevent.NewUpgradeOracleEvent(svc),
			)
			if err != nil {
				return fmt.Errorf("failed to start event subscription: %w", err)
			}

			httpServerErrCh := make(chan error, 1)
			srv := server.New(svc)

			go func() {
				if err := srv.Run(); err != nil {
					if !errors.Is(err, http.ErrServerClosed) {
						httpServerErrCh <- err
					} else {
						close(httpServerErrCh)
					}
				}
			}()

			sigChan := make(chan os.Signal, 1)

			signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
			select {
			case err := <-httpServerErrCh:
				if err != nil {
					log.Errorf("http server was closed with an error: %v", err)
				}
			case <-sigChan:
				log.Info("signal detected")
			}

			log.Infof("starting graceful shutdown")

			ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			if err := srv.Shutdown(ctxTimeout); err != nil {
				return fmt.Errorf("error occurs while server shutting down: %w", err)
			}

			return nil
		},
	}

	return cmd
}
