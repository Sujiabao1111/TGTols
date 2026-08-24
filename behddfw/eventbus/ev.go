package eventbus

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

// Handler is a type capable of handling events published through eventbus.
type Handler[T any] interface {
	OnEvent(event T)
}

type HandlerFunc[T any] func(event T)

func (f HandlerFunc[T]) OnEvent(event T) {
	f(event)
}

type handlerEntry struct {
	id      uint64
	handler interface{}
}

var (
	handlers            = make(map[reflect.Type][]handlerEntry)
	mu                  = sync.RWMutex{}
	subscriberId uint64 = 0
)

// Subscribe registers a handler for a given type. When this type is used with
// Publish or PublishAsync, the handler will be invoked. The return values is
// a subscription ID that can be used to unsubscribe the handler.
func Subscribe[T any](handler Handler[T]) uint64 {
	mu.Lock()
	defer mu.Unlock()

	id := generateHandlerId()
	eventType := reflect.TypeOf(*new(T))
	handlers[eventType] = append(handlers[eventType], handlerEntry{
		id:      id,
		handler: handler,
	})
	return id
}

// Unsubscribe removes a handler with the given subscription ID for the specified
// type. If the handler is not found, it returns false.
func Unsubscribe[T any](subscriptionID uint64) bool {
	mu.Lock()
	defer mu.Unlock()

	eventType := reflect.TypeOf(*new(T))
	handler, ok := handlers[eventType]
	if !ok {
		return false
	}

	for i, h := range handler {
		if h.id == subscriptionID {
			handlers[eventType] = append(handler[:i], handler[i+1:]...)
			return true
		}
	}
	return false
}

// Publish sends an event to all handlers registered for the event type. If no
// handlers are registered or the handler is not the correct type an error is
// returned. All handlers for the event type will be invoked in the order they
// were registered.
func Publish[T any](event T) error {
	mu.RLock()
	defer mu.RUnlock()

	eventType := reflect.TypeOf(event)
	handler, ok := handlers[eventType]
	if !ok {
		return fmt.Errorf("no handler for event %T", event)
	}

	for _, h := range handler {
		eventHandler, ok := h.handler.(Handler[T])
		if !ok {
			return fmt.Errorf("handler is not of type Handler[%T]", event)
		}
		eventHandler.OnEvent(event)
	}

	return nil
}

// MustPublish behaves like Publish sending an event to all handlers registered for
// the event type but panics on error.
func MustPublish[T any](event T) {
	if err := Publish(event); err != nil {
		panic(err)
	}
}

// PublishAsync sends an event to all handlers registered for the event type. If no
// handlers are registered or the handler is not the correct type an error is
// returned. All handlers for the event type will be invoked asynchronously in new
// goroutines.
func PublishAsync[T any](event T) error {
	mu.RLock()
	defer mu.RUnlock()

	eventType := reflect.TypeOf(event)
	handler, ok := handlers[eventType]
	if !ok {
		return fmt.Errorf("no handler for event %T", event)
	}

	for _, h := range handler {
		eventHandler, ok := h.handler.(Handler[T])
		if !ok {
			return fmt.Errorf("handler is not of type Handler[%T]", event)
		}
		go eventHandler.OnEvent(event)
	}

	return nil
}

// MustPublishAsync behaves like PublishAsync sending an event to all handlers
// registered for the event type asynchronously but panics on error.
func MustPublishAsync[T any](event T) {
	if err := PublishAsync(event); err != nil {
		panic(err)
	}
}

func generateHandlerId() uint64 {
	return atomic.AddUint64(&subscriberId, 1)
}
