package event

import (
	"context"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
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

func (s *PanaceaSubscriber) Run(eventType ...string) error {
	log.Infof("start panacea event subscriber")

	for _, e := range eventType {
		convertedEvent := convertEventStatusToEvent(e)
		query := convertedEvent.GetEventType() + "." + convertedEvent.GetEventAttributeKey() + "=" + convertedEvent.GetEventAttributeValue()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		txs, err := s.Client.Subscribe(ctx, "", query)
		if err != nil {
			return err
		}

		go func() {
			for tx := range txs {
				if err := convertedEvent.EventHandler(tx); err != nil {
					log.Errorf("failed to handle event '%s': %v", query, err)
				}
			}
		}()
	}

	return nil
}

func convertEventStatusToEvent(eventType string) Event {
	switch eventType {
	case types.EventTypeRegistrationVote:
		return RegisterOracleEvent{}
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
