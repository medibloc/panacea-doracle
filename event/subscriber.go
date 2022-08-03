package event

import (
	"context"
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	log "github.com/sirupsen/logrus"
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
		fmt.Println(query)
		txs, err := client.Subscribe(ctx, s.Subscriber, query)
		if err != nil {
			return err
		}
		go func() {
			select {
			case <-txs:
				fmt.Println(txs)
			}
		}()
	}

	return nil
}

func convertEventStatusToEvent(e PanaceaEventStatus) Event {
	switch e {
	case RegisterOracle:
		return RegisterOracleEvent{
			EventType:           "register",
			EventAttributeKey:   "oracle",
			EventAttributeValue: "RegisterOracleEvent",
		}
	default:
		return nil
	}
}
