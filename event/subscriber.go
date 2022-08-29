package event

import (
	"context"
	"github.com/medibloc/panacea-doracle/service"
	log "github.com/sirupsen/logrus"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"time"
)

type PanaceaSubscriber struct {
	Service *service.Service
	Client  *rpchttp.HTTP
}

// NewSubscriber generates a rpc http client with websocket address.
func NewSubscriber(svc *service.Service) (*PanaceaSubscriber, error) {
	client, err := rpchttp.New(svc.Conf.Panacea.WSAddr, "/websocket")
	if err != nil {
		return nil, err
	}

	err = client.Start()
	if err != nil {
		return nil, err
	}

	return &PanaceaSubscriber{
		Service: svc,
		Client:  client,
	}, nil
}

func (s *PanaceaSubscriber) Run(events ...Event) error {
	log.Infof("start panacea event subscriber")

	for _, e := range events {
		query := e.GetEventType() + "." + e.GetEventAttributeKey() + "=" + e.GetEventAttributeValue()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		txs, err := s.Client.Subscribe(ctx, "", query)
		if err != nil {
			return err
		}

		go func(e Event) {
			for tx := range txs {
				if err := e.EventHandler(tx); err != nil {
					log.Errorf("failed to handle event '%s': %v", query, err)
				}
			}
		}(e)
	}

	return nil
}

func (s *PanaceaSubscriber) Close() error {
	log.Infof("closing Panacea event subscriber")
	return s.Client.Stop()
}
