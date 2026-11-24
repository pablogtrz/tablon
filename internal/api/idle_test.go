package api_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/pablogtrz/tablon/internal/idle"
)

type statusResponse struct {
	Database string `json:"database"`
	Idle     struct {
		Enabled        bool       `json:"enabled"`
		State          string     `json:"state"`
		TimeoutSeconds int        `json:"timeout_seconds"`
		LastActivity   time.Time  `json:"last_activity"`
		IdleSince      *time.Time `json:"idle_since"`
	} `json:"idle"`
}

func idleTestController(t *testing.T, timeout time.Duration) *idle.Controller {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctl := idle.New(true, timeout, logger, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go ctl.Run(ctx)
	return ctl
}

func getStatus(t *testing.T, a *testAPI, headers map[string]string) statusResponse {
	t.Helper()
	code, body := a.do(t, http.MethodGet, "/api/v1/system/status", nil, headers)
	if code != http.StatusOK {
		t.Fatalf("status code = %d: %s", code, body)
	}
	return decode[statusResponse](t, body)
}

func waitForIdleState(t *testing.T, a *testAPI, headers map[string]string, want string) statusResponse {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		status := getStatus(t, a, headers)
		if status.Idle.State == want {
			return status
		}
		if time.Now().After(deadline) {
			t.Fatalf("idle state = %q, want %q", status.Idle.State, want)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestIdleStatusAndReactivation(t *testing.T) {
	ctl := idleTestController(t, 50*time.Millisecond)
	a := newTestAPIWithIdle(t, "", nil, ctl)

	status := waitForIdleState(t, a, nil, "idle")
	if !status.Idle.Enabled {
		t.Fatal("idle should be enabled")
	}
	if status.Database != "idle" {
		t.Fatalf("database = %q, want idle", status.Database)
	}
	if status.Idle.IdleSince == nil {
		t.Fatal("idle_since is nil")
	}

	code, body := a.do(t, http.MethodGet, "/api/v1/projects", nil, nil)
	if code != http.StatusOK {
		t.Fatalf("projects status = %d: %s", code, body)
	}

	status = getStatus(t, a, nil)
	if status.Idle.State != "active" {
		t.Fatalf("state after request = %q", status.Idle.State)
	}
	if status.Database != "ok" {
		t.Fatalf("database after request = %q", status.Database)
	}
	if status.Idle.IdleSince != nil {
		t.Fatalf("idle_since after request = %v", status.Idle.IdleSince)
	}
}

func TestIdleIgnoresHealthAndStatus(t *testing.T) {
	ctl := idleTestController(t, 50*time.Millisecond)
	a := newTestAPIWithIdle(t, "", nil, ctl)

	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
		if code, body := a.do(t, http.MethodGet, "/health", nil, nil); code != http.StatusOK {
			t.Fatalf("health = %d: %s", code, body)
		}
		if code, body := a.do(t, http.MethodGet, "/ready", nil, nil); code != http.StatusOK {
			t.Fatalf("ready = %d: %s", code, body)
		}
		time.Sleep(10 * time.Millisecond)
	}

	status := waitForIdleState(t, a, nil, "idle")
	if status.Database != "idle" {
		t.Fatalf("database = %q, want idle", status.Database)
	}
}

func TestUnauthorizedRequestDoesNotReactivate(t *testing.T) {
	ctl := idleTestController(t, 50*time.Millisecond)
	a := newTestAPIWithIdle(t, "secret", nil, ctl)
	auth := map[string]string{"Authorization": "Bearer secret"}

	waitForIdleState(t, a, auth, "idle")

	if code, _ := a.do(t, http.MethodGet, "/api/v1/projects", nil, nil); code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", code)
	}
	time.Sleep(150 * time.Millisecond)
	if status := getStatus(t, a, auth); status.Idle.State != "idle" {
		t.Fatalf("unauthorized request woke the server: %q", status.Idle.State)
	}
}

func TestOpenSSEKeepsActive(t *testing.T) {
	ctl := idleTestController(t, 50*time.Millisecond)
	a := newTestAPIWithIdle(t, "", nil, ctl)

	code, body := a.do(t, http.MethodPost, "/api/v1/projects", map[string]any{"name": "SSE idle"}, nil)
	if code != http.StatusCreated {
		t.Fatalf("create project = %d: %s", code, body)
	}
	project := decode[struct {
		Identifier string `json:"identifier"`
	}](t, body)

	req, err := http.NewRequest(http.MethodGet, a.srv.URL+"/api/v1/projects/"+project.Identifier+"/events", nil)
	if err != nil {
		t.Fatalf("new SSE request: %v", err)
	}
	res, err := a.srv.Client().Do(req)
	if err != nil {
		t.Fatalf("open SSE: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		t.Fatalf("SSE status = %d", res.StatusCode)
	}

	time.Sleep(150 * time.Millisecond)
	if status := getStatus(t, a, nil); status.Idle.State != "active" {
		t.Fatalf("SSE did not keep server active: %q", status.Idle.State)
	}

	res.Body.Close()
	waitForIdleState(t, a, nil, "idle")
}

func TestStatusJSONShape(t *testing.T) {
	a := newTestAPI(t, "")
	code, body := a.do(t, http.MethodGet, "/api/v1/system/status", nil, nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var idleRaw map[string]json.RawMessage
	if err := json.Unmarshal(raw["idle"], &idleRaw); err != nil {
		t.Fatalf("decode idle: %v", err)
	}
	for _, key := range []string{"enabled", "state", "timeout_seconds", "last_activity", "idle_since"} {
		if _, ok := idleRaw[key]; !ok {
			t.Fatalf("missing idle.%s", key)
		}
	}
}
