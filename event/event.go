package event

type Event interface {
	GetEventType() string

	GetEventAttributeKey() string

	GetEventAttributeValue() string

	GetEventHandler()
}
