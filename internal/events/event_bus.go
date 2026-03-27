package events

import (
	"sync"
)

type EventType string

const (
	EventPersonCreated      EventType = "person.created"
	EventPersonUpdated      EventType = "person.updated"
	EventPersonConverted    EventType = "person.converted"
	EventPersonScoreChanged EventType = "person.score_changed"
	EventPersonAssigned     EventType = "person.assigned"

	EventDealCreated        EventType = "deal.created"
	EventDealUpdated        EventType = "deal.updated"
	EventDealStageChanged   EventType = "deal.stage_changed"
	EventDealWon            EventType = "deal.won"
	EventDealLost           EventType = "deal.lost"
	EventDealCloseDatePassed EventType = "deal.close_date_passed"

	EventActivityCreated    EventType = "activity.created"
	EventActivityCompleted  EventType = "activity.completed"

	EventCompanyCreated     EventType = "company.created"
	EventCompanyUpdated     EventType = "company.updated"
)

type Event struct {
	Type      EventType
	TenantID  string
	EntityID  string
	EntityType string
	Payload   interface{}
}

type Handler func(Event) error

type EventBus struct {
	mu      sync.RWMutex
	handlers map[EventType][]Handler
}

var bus *EventBus
var once sync.Once

func GetBus() *EventBus {
	once.Do(func() {
		bus = &EventBus{
			handlers: make(map[EventType][]Handler),
		}
	})
	return bus
}

func (b *EventBus) Subscribe(eventType EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *EventBus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers, ok := b.handlers[event.Type]
	if !ok {
		return
	}

	for _, handler := range handlers {
		go func(h Handler) {
			_ = h(event)
		}(handler)
	}
}

func (b *EventBus) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = make(map[EventType][]Handler)
}

func PublishPersonCreated(tenantID, personID string, payload interface{}) {
	GetBus().Publish(Event{
		Type:       EventPersonCreated,
		TenantID:   tenantID,
		EntityID:   personID,
		EntityType: "person",
		Payload:    payload,
	})
}

func PublishPersonUpdated(tenantID, personID string, payload interface{}) {
	GetBus().Publish(Event{
		Type:       EventPersonUpdated,
		TenantID:   tenantID,
		EntityID:   personID,
		EntityType: "person",
		Payload:    payload,
	})
}

func PublishDealCreated(tenantID, dealID string, payload interface{}) {
	GetBus().Publish(Event{
		Type:       EventDealCreated,
		TenantID:   tenantID,
		EntityID:   dealID,
		EntityType: "deal",
		Payload:    payload,
	})
}

func PublishDealStageChanged(tenantID, dealID string, payload interface{}) {
	GetBus().Publish(Event{
		Type:       EventDealStageChanged,
		TenantID:   tenantID,
		EntityID:   dealID,
		EntityType: "deal",
		Payload:    payload,
	})
}

func PublishDealWon(tenantID, dealID string, payload interface{}) {
	GetBus().Publish(Event{
		Type:       EventDealWon,
		TenantID:   tenantID,
		EntityID:   dealID,
		EntityType: "deal",
		Payload:    payload,
	})
}

func PublishDealLost(tenantID, dealID string, payload interface{}) {
	GetBus().Publish(Event{
		Type:       EventDealLost,
		TenantID:   tenantID,
		EntityID:   dealID,
		EntityType: "deal",
		Payload:    payload,
	})
}

func PublishActivityCreated(tenantID, activityID string, payload interface{}) {
	GetBus().Publish(Event{
		Type:       EventActivityCreated,
		TenantID:   tenantID,
		EntityID:   activityID,
		EntityType: "activity",
		Payload:    payload,
	})
}

func PublishActivityCompleted(tenantID, activityID string, payload interface{}) {
	GetBus().Publish(Event{
		Type:       EventActivityCompleted,
		TenantID:   tenantID,
		EntityID:   activityID,
		EntityType: "activity",
		Payload:    payload,
	})
}

func PublishCompanyCreated(tenantID, companyID string, payload interface{}) {
	GetBus().Publish(Event{
		Type:       EventCompanyCreated,
		TenantID:   tenantID,
		EntityID:   companyID,
		EntityType: "company",
		Payload:    payload,
	})
}
