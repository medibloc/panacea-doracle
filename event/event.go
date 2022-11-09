package event

import (
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Event interface {
	Prepare() error

	GetEventName() string

	GetEventQuery() string

	SetEnable(bool)

	Enabled() bool

	EventHandler(ctypes.ResultEvent) error
}
