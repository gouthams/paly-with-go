package events

// Interface for the event related helpers
// Interface is defined outside of the util helper to encourage structural typing
type EventInterface interface {
	PublishEvent(str string) error
	GetStats() string
}
