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
	WSAddr     string
	Subscriber string
	Client     *rpchttp.HTTP
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
		WSAddr:     conf.Panacea.WSAddr,
		Subscriber: conf.BaseConfig.Subscriber,
		Client:     client,
	}, nil
}

type PanaceaEventStatus int32

const (
	RegisterOracle PanaceaEventStatus = 1
)

func (s *PanaceaSubscriber) Run(event ...PanaceaEventStatus) error {
	log.Infof("start panacea event subscriber")
	client, err := rpchttp.New(s.WSAddr, "/websocket")
	if err != nil {
		return err
	}

	err = client.Start()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for _, e := range event {
		convertedEvent := convertEventStatusToEvent(e)
		query := convertedEvent.GetEventType() + "." + convertedEvent.GetEventAttributeKey() + "=" + convertedEvent.GetEventAttributeValue()
		txs, err := client.Subscribe(ctx, s.Subscriber, query)
		if err != nil {
			return err
		}

		go func() {
			for _ = range txs {
				convertedEvent.GetEventHandler()
			}
		}()
	}

	return nil
}

func convertEventStatusToEvent(e PanaceaEventStatus) Event {
	switch e {
	case RegisterOracle:
		return RegisterOracleEvent{
			EventType:           "message",
			EventAttributeKey:   "action",
			EventAttributeValue: "'RegisterOracle'",
			EventHandler: func() error {
				// TODO: Executing Voting Tx
				fmt.Println("RegisterOracle Event Handler")
				return nil
			},
		}
	default:
		return nil
	}
}

func (s *PanaceaSubscriber) Close() error {
	log.Infof("closing panacea event subscriber")
	if err := s.Client.Stop(); err != nil {
		return err
	}
	return nil
}
