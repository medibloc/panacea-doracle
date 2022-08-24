package event

var _ Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {
	EventType           string
	EventAttributeKey   string
	EventAttributeValue string
	EventHandler        func() error
}

func (e RegisterOracleEvent) GetEventType() string {
	return e.EventType
}

func (e RegisterOracleEvent) GetEventAttributeKey() string {
	return e.EventAttributeKey
}

func (e RegisterOracleEvent) GetEventAttributeValue() string {
	return e.EventAttributeValue
}

func (e RegisterOracleEvent) GetEventHandler() func() error {
	return e.EventHandler
}
