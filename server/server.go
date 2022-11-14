package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-doracle/service"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	*http.Server
}

func New(svc *service.Service) *Server {
	router := mux.NewRouter()
	router.HandleFunc("/datadeal/request_verification", svc.ValidateData)

	return &Server{
		&http.Server{
			Handler:      router,
			Addr:         svc.Config().ListenAddr,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}
}

func (srv *Server) Run() error {
	log.Info("server started!!!!")
	return srv.ListenAndServe()
	//httpServerErrCh := make(chan error, 1)
	//go func() {
	//	log.Infof("Decentralized Oracle Started")
	//	if err := srv.ListenAndServe(); err != nil {
	//		if !errors.Is(err, http.ErrServerClosed) {
	//			httpServerErrCh <- err
	//		} else {
	//			close(httpServerErrCh)
	//		}
	//	}
	//}()
	//
	//signalCh := make(chan os.Signal, 1)
	//signal.Notify(signalCh, os.Interrupt)
	//select {
	//case err := <-httpServerErrCh:
	//	if err != nil {
	//		log.Errorf("http server was closed with an error: %v", err)
	//	}
	//case <-signalCh:
	//	log.Info("signal detected")
	//}
	//
	//log.Info("starting the graceful shutdown")
	//
	//log.Info("terminating HTTP server")
	//ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	//defer cancel()
	//
	//if err := srv.Shutdown(ctxTimeout); err != nil {
	//	return fmt.Errorf("error occurs while server shutting down: %w", err)
	//}
	//
	//return nil
}
