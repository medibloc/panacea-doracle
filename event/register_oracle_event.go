package event

import (
	"fmt"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct{}

func (e RegisterOracleEvent) GetEventType() string {
	return "message"
}

func (e RegisterOracleEvent) GetEventAttributeKey() string {
	return "action"
}

func (e RegisterOracleEvent) GetEventAttributeValue() string {
	return "'RegisterOracle'"
}

func (e RegisterOracleEvent) EventHandler(event ctypes.ResultEvent) error {
	// TODO: Verifying Remote Attestation and Executing Vote Txs
	s := event.Events[e.GetEventType()][0]
	fmt.Println(s)
	fmt.Println("RegisterOracle Event Handler")
	return nil
}
