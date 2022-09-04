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
		err := s.subscribe(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PanaceaSubscriber) subscribe(event Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	query := event.GetEventType() + "." + event.GetEventAttributeKey() + "=" + event.GetEventAttributeValue()

	txs, err := s.Client.Subscribe(ctx, "", query)
	if err != nil {
		return err
	}

	go func(event Event) {
		for tx := range txs {
			if err := event.EventHandler(tx, s.Service); err != nil {
				log.Errorf("failed to handle event '%s': %v", query, err)
			}
		}
	}(event)

	return nil
}

func (s *PanaceaSubscriber) Close() error {
	log.Infof("closing Panacea event subscriber")
	return s.Client.Stop()
}
