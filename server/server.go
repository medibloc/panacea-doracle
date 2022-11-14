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
	log.Infof("server started on %s", srv.Addr)
	return srv.ListenAndServe()
}
