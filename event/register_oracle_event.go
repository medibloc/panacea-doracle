package event

var _ Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {
	EventType      string
	EventAttribute string
}

func (e RegisterOracleEvent) GetEventType() string {
	return e.EventType
}

func (e RegisterOracleEvent) GetEventAttribute() string {
	return e.EventAttribute
}

func (e RegisterOracleEvent) GetEventHandler() {
	panic("implement me")
}
