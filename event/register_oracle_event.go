package event

import "fmt"

var _ Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {}

func (e RegisterOracleEvent) GetEventType() string {
	return "message"
}

func (e RegisterOracleEvent) GetEventAttributeKey() string {
	return "action"
}

func (e RegisterOracleEvent) GetEventAttributeValue() string {
	return "'RegisterOracle'"
}

func (e RegisterOracleEvent) EventHandler() error {
	// TODO: Executing Voting Tx
	fmt.Println("RegisterOracle Event Handler")
	return nil
}
