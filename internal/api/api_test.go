package api_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pablogtrz/tablon/internal/api"
	"github.com/pablogtrz/tablon/internal/config"
	"github.com/pablogtrz/tablon/internal/events"
	"github.com/pablogtrz/tablon/internal/service"
	"github.com/pablogtrz/tablon/internal/store"
)

type testAPI struct {
	srv *httptest.Server
}

func newTestAPI(t *testing.T, token string) *testAPI {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	cfg := config.Config{
		Host:     "127.0.0.1",
		Port:     0,
		DataDir:  t.TempDir(),
		APIToken: token,
		LogLevel: slog.LevelError,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := events.NewHub(logger)
	t.Cleanup(hub.Close)
	srv := httptest.NewServer(api.New(cfg, service.New(st, hub), hub, logger))
	t.Cleanup(srv.Close)
	return &testAPI{srv: srv}
}

func (a *testAPI) do(t *testing.T, method, path string, body any, headers map[string]string) (int, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, a.srv.URL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	res, err := a.srv.Client().Do(req)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return res.StatusCode, raw
}

func decode[T any](t *testing.T, body []byte) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatalf("decode response %q: %v", string(body), err)
	}
	return value
}

func TestProjectEpicIssueFlow(t *testing.T) {
	a := newTestAPI(t, "")

	status, body := a.do(t, http.MethodPost, "/api/v1/projects", map[string]any{
		"name":       "API Project",
		"identifier": "API",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("create project status = %d, body = %s", status, body)
	}
	project := decode[map[string]any](t, body)
	projectID := project["id"].(string)

	status, body = a.do(t, http.MethodGet, "/api/v1/projects/API", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("get project status = %d, body = %s", status, body)
	}

	status, body = a.do(t, http.MethodPost, "/api/v1/projects/"+projectID+"/epics", map[string]any{
		"title": "Epic One",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("create epic status = %d, body = %s", status, body)
	}
	epic := decode[map[string]any](t, body)
	epicID := epic["id"].(string)

	status, body = a.do(t, http.MethodPost, "/api/v1/projects/"+projectID+"/issues", map[string]any{
		"title":    "Issue One",
		"epic_id":  epicID,
		"priority": "high",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("create issue status = %d, body = %s", status, body)
	}
	issue := decode[map[string]any](t, body)
	if issue["identifier"] != "API-1" {
		t.Fatalf("issue identifier = %v, want API-1", issue["identifier"])
	}
	issueID := issue["id"].(string)

	status, body = a.do(t, http.MethodGet, "/api/v1/issues/API-1", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("get issue status = %d, body = %s", status, body)
	}

	status, body = a.do(t, http.MethodPatch, "/api/v1/issues/"+issueID, map[string]any{"epic_id": nil}, nil)
	if status != http.StatusOK {
		t.Fatalf("clear epic status = %d, body = %s", status, body)
	}
	cleared := decode[map[string]any](t, body)
	if cleared["epic_id"] != nil {
		t.Fatalf("epic_id = %v, want null", cleared["epic_id"])
	}

	status, body = a.do(t, http.MethodPatch, "/api/v1/issues/"+issueID+"/status", map[string]any{
		"status": "in-progress",
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", status, body)
	}
	updated := decode[map[string]any](t, body)
	if updated["status"] != "in-progress" {
		t.Fatalf("status = %v, want in-progress", updated["status"])
	}

	status, body = a.do(t, http.MethodPost, "/api/v1/issues/"+issueID+"/claim", map[string]any{
		"agent_id": "agent-a",
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("claim status = %d, body = %s", status, body)
	}
	status, body = a.do(t, http.MethodPost, "/api/v1/issues/"+issueID+"/claim", map[string]any{
		"agent_id": "agent-b",
	}, nil)
	if status != http.StatusConflict {
		t.Fatalf("second claim status = %d, body = %s", status, body)
	}
	status, _ = a.do(t, http.MethodPost, "/api/v1/issues/"+issueID+"/release", map[string]any{
		"agent_id": "agent-b",
	}, nil)
	if status != http.StatusConflict {
		t.Fatalf("wrong release status = %d", status)
	}
	status, _ = a.do(t, http.MethodPost, "/api/v1/issues/"+issueID+"/release", map[string]any{
		"agent_id": "agent-a",
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("release status = %d", status)
	}

	status, _ = a.do(t, http.MethodDelete, "/api/v1/issues/"+issueID, nil, nil)
	if status != http.StatusNoContent {
		t.Fatalf("delete issue status = %d", status)
	}
	status, _ = a.do(t, http.MethodGet, "/api/v1/issues/"+issueID, nil, nil)
	if status != http.StatusNotFound {
		t.Fatalf("get deleted issue status = %d", status)
	}

	status, _ = a.do(t, http.MethodDelete, "/api/v1/projects/"+projectID, nil, nil)
	if status != http.StatusNoContent {
		t.Fatalf("delete project status = %d", status)
	}
}

func TestAuthIsOptionalAndEnforcedWhenConfigured(t *testing.T) {
	a := newTestAPI(t, "secret")

	status, _ := a.do(t, http.MethodGet, "/api/v1/projects", nil, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("unauth status = %d, want 401", status)
	}
	status, _ = a.do(t, http.MethodGet, "/api/v1/projects", nil, map[string]string{
		"Authorization": "Bearer secret",
	})
	if status != http.StatusOK {
		t.Fatalf("auth status = %d, want 200", status)
	}
	status, _ = a.do(t, http.MethodGet, "/health", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("health status = %d, want 200", status)
	}
	status, _ = a.do(t, http.MethodGet, "/api/v1/issues/PROJ-1/comments", nil, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("comments unauth status = %d, want 401", status)
	}
	status, _ = a.do(t, http.MethodGet, "/api/v1/projects/PROJ/events", nil, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("events unauth status = %d, want 401", status)
	}
}

func TestValidationAndDecodeErrors(t *testing.T) {
	a := newTestAPI(t, "")

	status, body := a.do(t, http.MethodPost, "/api/v1/projects", map[string]any{"name": ""}, nil)
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("empty name status = %d, body = %s", status, body)
	}
	status, body = a.do(t, http.MethodPost, "/api/v1/projects", map[string]any{
		"name": "X", "status": "unknown",
	}, nil)
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("bad status = %d, body = %s", status, body)
	}
	status, _ = a.do(t, http.MethodPost, "/api/v1/projects", map[string]any{"unknown": true}, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("unknown field status = %d, want 400", status)
	}
}

func TestProjectCountsArchiveAndIssuePagination(t *testing.T) {
	a := newTestAPI(t, "")

	status, body := a.do(t, http.MethodPost, "/api/v1/projects", map[string]any{
		"name": "Board", "identifier": "BRD",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("create project status = %d, body = %s", status, body)
	}
	project := decode[map[string]any](t, body)
	projectID := project["id"].(string)

	_, body = a.do(t, http.MethodPost, "/api/v1/projects/"+projectID+"/epics", map[string]any{"title": "Epic"}, nil)
	epic := decode[map[string]any](t, body)
	for _, input := range []map[string]any{
		{"title": "Low", "priority": "low"},
		{"title": "Urgent", "priority": "urgent"},
		{"title": "High", "priority": "high"},
	} {
		input["epic_id"] = epic["id"]
		status, body = a.do(t, http.MethodPost, "/api/v1/projects/"+projectID+"/issues", input, nil)
		if status != http.StatusCreated {
			t.Fatalf("create issue status = %d, body = %s", status, body)
		}
	}

	status, body = a.do(t, http.MethodGet, "/api/v1/projects?include_archived=true", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("list projects status = %d", status)
	}
	var projects struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(body, &projects); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	if len(projects.Items) != 1 {
		t.Fatalf("projects = %d, want 1", len(projects.Items))
	}
	if projects.Items[0]["epic_count"].(float64) != 1 || projects.Items[0]["issue_count"].(float64) != 3 {
		t.Fatalf("unexpected counts: %+v", projects.Items[0])
	}

	status, body = a.do(t, http.MethodGet, "/api/v1/projects/"+projectID+"/issues?limit=2&sort=priority&order=desc", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("list issues status = %d, body = %s", status, body)
	}
	var page struct {
		Items      []map[string]any `json:"items"`
		NextCursor string           `json:"next_cursor"`
	}
	if err := json.Unmarshal(body, &page); err != nil {
		t.Fatalf("decode issues: %v", err)
	}
	if len(page.Items) != 2 || page.NextCursor == "" {
		t.Fatalf("unexpected page: %+v", page)
	}
	if page.Items[0]["title"] != "Urgent" || page.Items[1]["title"] != "High" {
		t.Fatalf("unexpected order: %v, %v", page.Items[0]["title"], page.Items[1]["title"])
	}

	status, body = a.do(t, http.MethodGet, "/api/v1/projects/"+projectID+"/issues?limit=2&sort=priority&order=desc&cursor="+page.NextCursor, nil, nil)
	if status != http.StatusOK {
		t.Fatalf("next page status = %d, body = %s", status, body)
	}
	if err := json.Unmarshal(body, &page); err != nil {
		t.Fatalf("decode next page: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0]["title"] != "Low" {
		t.Fatalf("unexpected next page: %+v", page.Items)
	}

	status, _ = a.do(t, http.MethodPatch, "/api/v1/projects/"+projectID, map[string]any{"status": "archived"}, nil)
	if status != http.StatusOK {
		t.Fatalf("archive status = %d", status)
	}
	status, body = a.do(t, http.MethodGet, "/api/v1/projects", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("list archived projects status = %d", status)
	}
	if err := json.Unmarshal(body, &projects); err != nil {
		t.Fatalf("decode archived projects: %v", err)
	}
	if len(projects.Items) != 0 {
		t.Fatalf("archived project leaked into default list")
	}
}

func TestCommentsCRUDAndIdempotency(t *testing.T) {
	a := newTestAPI(t, "")

	status, body := a.do(t, http.MethodPost, "/api/v1/projects", map[string]any{
		"name": "Comments", "identifier": "CMS",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("create project status = %d, body = %s", status, body)
	}
	project := decode[map[string]any](t, body)
	projectID := project["id"].(string)

	status, body = a.do(t, http.MethodPost, "/api/v1/projects/"+projectID+"/issues", map[string]any{
		"title": "Commented issue",
	}, map[string]string{"Idempotency-Key": "issue-key"})
	if status != http.StatusCreated {
		t.Fatalf("create issue status = %d, body = %s", status, body)
	}
	issue := decode[map[string]any](t, body)
	issueID := issue["id"].(string)
	identifier := issue["identifier"].(string)

	status, body = a.do(t, http.MethodPost, "/api/v1/projects/"+projectID+"/issues", map[string]any{
		"title": "Commented issue",
	}, map[string]string{"Idempotency-Key": "issue-key"})
	if status != http.StatusCreated {
		t.Fatalf("replayed issue status = %d, body = %s", status, body)
	}
	replayed := decode[map[string]any](t, body)
	if replayed["id"] != issueID {
		t.Fatalf("replayed issue id = %v, want %v", replayed["id"], issueID)
	}

	status, body = a.do(t, http.MethodGet, "/api/v1/projects/"+projectID+"/issues", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("list issues status = %d", status)
	}
	var page struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(body, &page); err != nil {
		t.Fatalf("decode issues: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("issues = %d, want 1 after idempotent replay", len(page.Items))
	}

	status, body = a.do(t, http.MethodPost, "/api/v1/issues/"+identifier+"/comments", map[string]any{
		"content":  "first **comment**",
		"agent_id": "agent-a",
	}, map[string]string{"Idempotency-Key": "comment-key"})
	if status != http.StatusCreated {
		t.Fatalf("create comment status = %d, body = %s", status, body)
	}
	comment := decode[map[string]any](t, body)
	commentID := comment["id"].(string)
	if comment["agent_id"] != "agent-a" {
		t.Fatalf("agent_id = %v", comment["agent_id"])
	}

	status, body = a.do(t, http.MethodPost, "/api/v1/issues/"+identifier+"/comments", map[string]any{
		"content":  "first **comment**",
		"agent_id": "agent-a",
	}, map[string]string{"Idempotency-Key": "comment-key"})
	if status != http.StatusCreated {
		t.Fatalf("replayed comment status = %d, body = %s", status, body)
	}
	dupe := decode[map[string]any](t, body)
	if dupe["id"] != commentID {
		t.Fatalf("replayed comment id = %v, want %v", dupe["id"], commentID)
	}

	status, body = a.do(t, http.MethodGet, "/api/v1/issues/"+identifier+"/comments", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("list comments status = %d, body = %s", status, body)
	}
	var comments struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(body, &comments); err != nil {
		t.Fatalf("decode comments: %v", err)
	}
	if len(comments.Items) != 1 {
		t.Fatalf("comments = %d, want 1", len(comments.Items))
	}

	status, body = a.do(t, http.MethodPatch, "/api/v1/comments/"+commentID, map[string]any{
		"content": "edited",
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("update comment status = %d, body = %s", status, body)
	}
	edited := decode[map[string]any](t, body)
	if edited["content"] != "edited" {
		t.Fatalf("content = %v", edited["content"])
	}

	status, body = a.do(t, http.MethodDelete, "/api/v1/comments/"+commentID, nil, nil)
	if status != http.StatusNoContent {
		t.Fatalf("delete comment status = %d, body = %s", status, body)
	}
	status, _ = a.do(t, http.MethodPatch, "/api/v1/comments/"+commentID, map[string]any{"content": "x"}, nil)
	if status != http.StatusNotFound {
		t.Fatalf("update deleted comment status = %d", status)
	}

	status, body = a.do(t, http.MethodPost, "/api/v1/issues/"+identifier+"/comments", map[string]any{
		"content": "different",
	}, map[string]string{"Idempotency-Key": "comment-key"})
	if status != http.StatusConflict {
		t.Fatalf("different payload status = %d, body = %s", status, body)
	}
}

func TestProjectEventsStreamsIssueCreated(t *testing.T) {
	a := newTestAPI(t, "")

	status, body := a.do(t, http.MethodPost, "/api/v1/projects", map[string]any{
		"name": "Events", "identifier": "EVT",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("create project status = %d, body = %s", status, body)
	}
	project := decode[map[string]any](t, body)
	projectID := project["id"].(string)

	req, err := http.NewRequest(http.MethodGet, a.srv.URL+"/api/v1/projects/"+projectID+"/events", nil)
	if err != nil {
		t.Fatalf("new stream request: %v", err)
	}
	resp, err := a.srv.Client().Do(req)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stream status = %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("content type = %q", ct)
	}
	if resp.Header.Get("X-Accel-Buffering") != "no" {
		t.Fatalf("x-accel-buffering = %q", resp.Header.Get("X-Accel-Buffering"))
	}

	lines := make(chan string)
	go func() {
		defer close(lines)
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			lines <- line
		}
	}()
	readLine := func() string {
		t.Helper()
		select {
		case line, ok := <-lines:
			if !ok {
				t.Fatal("stream closed")
			}
			return line
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for stream line")
			return ""
		}
	}

	if line := readLine(); !strings.HasPrefix(line, "retry:") {
		t.Fatalf("first line = %q", line)
	}
	if line := readLine(); strings.TrimSpace(line) != "" {
		t.Fatalf("expected blank after retry, got %q", line)
	}

	status, body = a.do(t, http.MethodPost, "/api/v1/projects/"+projectID+"/issues", map[string]any{
		"title": "Streamed issue",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("create issue status = %d, body = %s", status, body)
	}

	var eventName, data string
	for {
		line := readLine()
		if strings.TrimSpace(line) == "" {
			break
		}
		if strings.HasPrefix(line, "event: ") {
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event: "))
		}
		if strings.HasPrefix(line, "data: ") {
			data = strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		}
	}
	if eventName != "issue.created" {
		t.Fatalf("event = %q", eventName)
	}
	var event map[string]any
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		t.Fatalf("decode event: %v (%s)", err, data)
	}
	if event["type"] != "issue.created" || event["project_id"] != projectID {
		t.Fatalf("event payload = %+v", event)
	}
}
