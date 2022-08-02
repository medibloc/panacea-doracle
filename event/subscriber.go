package event

import (
	"context"
	"github.com/medibloc/panacea-doracle/config"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"time"
)

type PanaceaSubscriber struct {
	WSAddr     string
	Subscriber string
}

func NewSubscriber(conf *config.Config, subscriber string) (PanaceaSubscriber, error) {
	return PanaceaSubscriber{
		WSAddr:     conf.Panacea.WSAddr,
		Subscriber: subscriber,
	}, nil
}

func (s PanaceaSubscriber) Run(event ...Event) error {
	client, err := rpchttp.New(s.WSAddr, "/websocket")
	if err != nil {
		return err
	}

	err = client.Start()
	if err != nil {
		return err
	}
	defer client.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for _, e := range event {
		query := e.GetEventType() + "." + e.GetEventAttribute()
		txs, err := client.Subscribe(ctx, s.Subscriber, query)
		if err != nil {
			return err
		}
		go func() {
			select {
			case <-txs:
				e.GetEventHandler()
			}
		}()
	}

	return nil
}
