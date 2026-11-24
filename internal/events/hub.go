package events

import (
	"log/slog"
	"sync"
)

const subscriberBuffer = 16

type Event struct {
	Type       string `json:"type"`
	ProjectID  string `json:"project_id"`
	IssueID    string `json:"issue_id,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	CommentID  string `json:"comment_id,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan Event]struct{}
	logger      *slog.Logger
	closed      bool
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		subscribers: make(map[string]map[chan Event]struct{}),
		logger:      logger,
	}
}

func (h *Hub) Subscribe(projectID string) (<-chan Event, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Event, subscriberBuffer)
	if h.closed {
		close(ch)
		return ch, func() {}
	}
	if h.subscribers[projectID] == nil {
		h.subscribers[projectID] = make(map[chan Event]struct{})
	}
	h.subscribers[projectID][ch] = struct{}{}

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			subs := h.subscribers[projectID]
			if subs == nil {
				return
			}
			if _, ok := subs[ch]; !ok {
				return
			}
			delete(subs, ch)
			close(ch)
			if len(subs) == 0 {
				delete(h.subscribers, projectID)
			}
		})
	}
	return ch, cancel
}

func (h *Hub) Publish(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return
	}
	for ch := range h.subscribers[event.ProjectID] {
		select {
		case ch <- event:
		default:
			if h.logger != nil {
				h.logger.Debug("drop event for slow subscriber", "project_id", event.ProjectID, "type", event.Type)
			}
		}
	}
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for projectID, subs := range h.subscribers {
		for ch := range subs {
			close(ch)
		}
		delete(h.subscribers, projectID)
	}
}
