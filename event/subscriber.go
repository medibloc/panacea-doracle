package event

import (
	"context"
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	log "github.com/sirupsen/logrus"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"time"
)

type PanaceaSubscriber struct {
	Client *rpchttp.HTTP
}

// NewSubscriber generates a rpc http client with websocket address.
func NewSubscriber(conf *config.Config) (*PanaceaSubscriber, error) {
	client, err := rpchttp.New(conf.Panacea.WSAddr, "/websocket")
	if err != nil {
		return nil, err
	}

	err = client.Start()
	if err != nil {
		return nil, err
	}

	return &PanaceaSubscriber{
		Client: client,
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
			for t := range txs {
				fmt.Println("got ", t.Events)
				_ = e.EventHandler(t)
			}
		}(e)
	}

	return nil
}

func (s *PanaceaSubscriber) Close() error {
	log.Infof("closing panacea event subscriber")
	return s.Client.Stop()
}
