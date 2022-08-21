package event

import (
	"context"
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	log "github.com/sirupsen/logrus"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/types"
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

type PanaceaEventStatus int32

const (
	RegisterOracle PanaceaEventStatus = 1
)

func (s PanaceaSubscriber) Run(event ...PanaceaEventStatus) error {
	log.Infof("Subscribe Panacea Event run")
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
		convertedEvent := convertEventStatusToEvent(e)
		query := convertedEvent.GetEventType() + "." + convertedEvent.GetEventAttributeKey() + "=" + convertedEvent.GetEventAttributeValue()
		txs, err := client.Subscribe(ctx, s.Subscriber, query)
		if err != nil {
			return err
		}

		go func() {
			for t := range txs {
				fmt.Println(t.Data.(types.EventDataTx))
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
		}
	default:
		return nil
	}
}
