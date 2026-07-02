package tui

import (
	"sync"

	"github.com/lobaton-dev/chigui-db/internal/models"
)

type EventType int

const (
	EventConnChanged EventType = iota
	EventQueryExecuted
	EventSchemaChanged
	EventDataChanged
)

type Event struct {
	Type EventType
	Data any
}

type EventBus struct {
	mu        sync.RWMutex
	listeners map[EventType][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[EventType][]chan Event),
	}
}

func (b *EventBus) Subscribe(et EventType, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.listeners[et] = append(b.listeners[et], ch)
}

func (b *EventBus) Unsubscribe(et EventType, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	listeners := b.listeners[et]
	for i, l := range listeners {
		if l == ch {
			b.listeners[et] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

func (b *EventBus) Emit(e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, ch := range b.listeners[e.Type] {
		select {
		case ch <- e:
		default:
		}
	}
}

// Screen messages

type NavigateToMsg struct {
	Screen Screen
	Data   any
}

type ConnSelectedMsg struct {
	Conn *models.ConnectionInfo
	Pool any // *conn.Pool — defined in Fase 2
}

type ErrorMsg struct {
	Err error
}

type QueryResultMsg struct {
	Result any
	Error  error
}

type SchemaLoadedMsg struct {
	Database string
	Tables   []any
}

type DataPageMsg struct {
	Columns []models.Column
	Rows    [][]any
	Offset  int
	Total   int
}
