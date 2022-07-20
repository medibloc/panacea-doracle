package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/server/service"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
)

func Run(conf *config.Config) error {
	svc, err := service.New(conf)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	server := &http.Server{
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	httpServerErrCh := make(chan error, 1)
	go func() {
		log.Infof("Decentralized Oracle Started")
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				httpServerErrCh <- err
			} else {
				close(httpServerErrCh)
			}
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	select {
	case err := <-httpServerErrCh:
		if err != nil {
			log.Errorf("http server was closed with an error: %v", err)
		}
	case <-signalCh:
		log.Info("signal detected")
	}
	log.Info("starting the graceful shutdown")

	log.Info("terminating HTTP server")
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := server.Shutdown(ctxTimeout); err != nil {
		return fmt.Errorf("error occurs while server shutting down: %w", err)
	}

	return nil
}
