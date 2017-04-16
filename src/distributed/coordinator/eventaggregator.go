package coordinator

import (
	"time"
)

type EventRaiser interface {
	AddListener(eventName string, f func(interface{}))
	GetSensors() <-chan string 
}

type EventAggregator struct {
	listeners map[string][]func(interface{})
	sensors chan string
}

func NewEventAggregator() *EventAggregator {
	ea := EventAggregator{
		listeners: make(map[string][]func(interface{})),
	}
	return &ea
}

func (ea *EventAggregator) AddListener(name string, f func(interface{})) {
	ea.listeners[name] = append(ea.listeners[name], f)
}

// trigger event and call every event handler func
func (ea *EventAggregator) PublishEvent(name string, eventData interface{}) {
	if ea.listeners[name] != nil {
		for _, r := range ea.listeners[name] {
			r(eventData)
		}
	}
}

func (ea *EventAggregator) GetSensors()  <-chan string {
	return ea.sensors
}

type EventData struct {
	Name      string
	Value     float64
	Timestamp time.Time
}
