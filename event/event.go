package event

import (
	"github.com/medibloc/panacea-doracle/service"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Event interface {
	GetEventType() string

	GetEventAttributeKey() string

	GetEventAttributeValue() string

	EventHandler(event ctypes.ResultEvent, svc *service.Service) error
}
