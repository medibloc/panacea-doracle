package event

import ctypes "github.com/tendermint/tendermint/rpc/core/types"

type Event interface {
	GetEventType() string

	GetEventAttributeKey() string

	GetEventAttributeValue() string

	EventHandler(event ctypes.ResultEvent) error
}
