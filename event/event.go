package event

type Event interface {
	GetEventType() string

	GetEventAttribute() string

	GetEventHandler()
}
