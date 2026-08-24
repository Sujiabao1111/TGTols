package eventbus

import (
	"fmt"
	"testing"
)

type Status string

const (
	StatusUp       Status = "UP"
	StatusDegraded Status = "DEGRADED"
	StatusDown     Status = "DOWN"
)

type ApplicationHealthChangedEvent struct {
	NewStatus string
}

func TestPublish2(t *testing.T) {
	// how to
	handlerFn := HandlerFunc[ApplicationHealthChangedEvent](func(event ApplicationHealthChangedEvent) {
		fmt.Println("Application status changed to", event.NewStatus)
		// todo: do something useful with the event
	})

	// Register the handler for the event type ApplicationHealthChangedEvent
	Subscribe[ApplicationHealthChangedEvent](handlerFn)

	// Publish some events
	MustPublish(ApplicationHealthChangedEvent{NewStatus: string(StatusDegraded)})
	MustPublish(ApplicationHealthChangedEvent{NewStatus: string(StatusUp)})
}
