package events

import (
	"testing"
	"time"
)

func TestHubProjectIsolationAndUnsubscribe(t *testing.T) {
	h := NewHub(nil)
	defer h.Close()

	a, cancelA := h.Subscribe("a")
	defer cancelA()
	b, cancelB := h.Subscribe("b")
	defer cancelB()

	h.Publish(Event{Type: "issue.updated", ProjectID: "a", Identifier: "A-1"})

	select {
	case event := <-a:
		if event.Type != "issue.updated" || event.Identifier != "A-1" {
			t.Fatalf("event = %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("project a did not receive event")
	}

	select {
	case event := <-b:
		t.Fatalf("project b received %+v", event)
	case <-time.After(20 * time.Millisecond):
	}

	cancelA()
	select {
	case _, ok := <-a:
		if ok {
			t.Fatal("channel still open after unsubscribe")
		}
	case <-time.After(time.Second):
		t.Fatal("channel was not closed")
	}
}

func TestHubDropsForSlowSubscriber(t *testing.T) {
	h := NewHub(nil)
	defer h.Close()
	ch, cancel := h.Subscribe("a")
	defer cancel()

	for i := 0; i < subscriberBuffer+5; i++ {
		h.Publish(Event{Type: "issue.updated", ProjectID: "a"})
	}

	count := 0
	for {
		select {
		case <-ch:
			count++
			if count == subscriberBuffer {
				if len(ch) != 0 {
					t.Fatalf("buffer length = %d, want 0", len(ch))
				}
				return
			}
		case <-time.After(time.Second):
			t.Fatalf("received %d events, want %d", count, subscriberBuffer)
		}
	}
}

func TestHubCloseClosesSubscribers(t *testing.T) {
	h := NewHub(nil)
	ch, _ := h.Subscribe("a")
	h.Close()
	if _, ok := <-ch; ok {
		t.Fatal("channel still open after Close")
	}
	h.Publish(Event{Type: "issue.updated", ProjectID: "a"})
}
