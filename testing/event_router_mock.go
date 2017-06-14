package testing

import (
	"sync"

	"github.com/cloudfoundry/sonde-go/events"
)

type MockEventRouter struct {
	lock   sync.Mutex
	events []*events.Envelope
}

func NewMockEventRouter() *MockEventRouter {
	return &MockEventRouter{}
}

func (router *MockEventRouter) RouteEvent(msg *events.Envelope) {
	router.lock.Lock()
	router.events = append(router.events, msg)
	router.lock.Unlock()
}

func (router *MockEventRouter) Events() []*events.Envelope {
	var events []*events.Envelope

	router.lock.Lock()
	events = append(events, router.events...)
	router.lock.Unlock()

	return events
}
