package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pablogtrz/tablon/internal/version"
)

type fakeAPI struct {
	t         *testing.T
	lastQuery string
	lastBody  map[string]any
	lastKey   string
	lastAuth  string
}

func (f *fakeAPI) handler(w http.ResponseWriter, r *http.Request) {
	f.lastQuery = r.URL.RawQuery
	f.lastAuth = r.Header.Get("Authorization")
	body, _ := io.ReadAll(r.Body)
	if len(body) > 0 {
		_ = json.Unmarshal(body, &f.lastBody)
	}
	if key := r.Header.Get("Idempotency-Key"); key != "" {
		f.lastKey = key
	}

	write := func(status int, value any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if value != nil {
			_ = json.NewEncoder(w).Encode(value)
		}
	}

	issue := map[string]any{
		"id":          "issue-1",
		"identifier":  "PROJ-1",
		"project_id":  "project-1",
		"epic_id":     nil,
		"title":       "Implement auth",
		"description": "Details",
		"status":      "todo",
		"priority":    "high",
		"agent_id":    nil,
		"created_at":  "2026-11-24T10:00:00Z",
		"updated_at":  "2026-11-24T10:00:00Z",
	}
	comment := map[string]any{
		"id":         "comment-1",
		"issue_id":   "issue-1",
		"content":    "Started",
		"agent_id":   "bot",
		"created_at": "2026-11-24T10:00:00Z",
		"updated_at": "2026-11-24T10:00:00Z",
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
		write(200, map[string]any{"items": []any{map[string]any{
			"id": "project-1", "identifier": "PROJ", "name": "Project",
			"description": "", "status": "active", "logo_gradient": 0,
			"logo_emoji": "", "epic_count": 0, "issue_count": 1, "done_count": 0,
			"created_at": "2026-11-24T10:00:00Z", "updated_at": "2026-11-24T10:00:00Z",
		}}})
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/PROJ":
		write(200, map[string]any{
			"id": "project-1", "identifier": "PROJ", "name": "Project",
			"description": "A project", "status": "active", "logo_gradient": 0,
			"logo_emoji": "", "created_at": "2026-11-24T10:00:00Z",
			"updated_at": "2026-11-24T10:00:00Z",
		})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects":
		write(201, map[string]any{
			"id": "project-2", "identifier": "NEW", "name": f.lastBody["name"],
			"description": f.lastBody["description"], "status": "planning", "logo_gradient": 0,
			"logo_emoji": "", "created_at": "2026-11-24T10:00:00Z",
			"updated_at": "2026-11-24T10:00:00Z",
		})
	case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/projects/PROJ":
		write(200, map[string]any{
			"id": "project-1", "identifier": "PROJ", "name": f.lastBody["name"],
			"description": "A project", "status": f.lastBody["status"], "logo_gradient": 0,
			"logo_emoji": "", "created_at": "2026-11-24T10:00:00Z",
			"updated_at": "2026-11-24T11:00:00Z",
		})
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/PROJ/epics":
		write(200, map[string]any{"items": []any{map[string]any{
			"id": "epic-1", "project_id": "project-1", "title": "Epic one",
			"description": "", "status": "backlog",
			"created_at": "2026-11-24T10:00:00Z", "updated_at": "2026-11-24T10:00:00Z",
		}}})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/PROJ/epics":
		write(201, map[string]any{
			"id": "epic-2", "project_id": "project-1", "title": f.lastBody["title"],
			"description": f.lastBody["description"], "status": "backlog",
			"created_at": "2026-11-24T10:00:00Z", "updated_at": "2026-11-24T10:00:00Z",
		})
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/epics/epic-1":
		write(200, map[string]any{
			"id": "epic-1", "project_id": "project-1", "title": "Epic one",
			"description": "", "status": "backlog",
			"created_at": "2026-11-24T10:00:00Z", "updated_at": "2026-11-24T10:00:00Z",
		})
	case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/epics/epic-1":
		write(200, map[string]any{
			"id": "epic-1", "project_id": "project-1", "title": f.lastBody["title"],
			"description": "", "status": "in-progress",
			"created_at": "2026-11-24T10:00:00Z", "updated_at": "2026-11-24T11:00:00Z",
		})
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/PROJ/issues":
		write(200, map[string]any{"items": []any{issue}})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/PROJ/issues":
		write(201, issue)
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/issues/PROJ-1":
		write(200, issue)
	case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/issues/PROJ-1":
		for key, value := range f.lastBody {
			issue[key] = value
		}
		write(200, issue)
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/issues/PROJ-1/comments":
		write(200, map[string]any{"items": []any{comment}})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/issues/PROJ-1/comments":
		write(201, comment)
	case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/comments/comment-1":
		comment["content"] = "Edited"
		write(200, comment)
	case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/comments/comment-1":
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/issues/PROJ-1/claim":
		issue["agent_id"] = "bot"
		write(200, issue)
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/issues/PROJ-1/release":
		issue["agent_id"] = nil
		write(200, issue)
	case r.URL.Path == "/api/v1/projects/MISSING":
		write(404, map[string]any{"error": map[string]any{"code": "not_found", "message": "resource not found"}})
	default:
		write(404, map[string]any{"error": map[string]any{"code": "not_found", "message": "resource not found"}})
	}
}

func newFake(t *testing.T) (*fakeAPI, func(args []string, env map[string]string, stdin string) (int, string, string)) {
	t.Helper()
	f := &fakeAPI{t: t}
	srv := httptest.NewServer(http.HandlerFunc(f.handler))
	t.Cleanup(srv.Close)
	envFn := func(env map[string]string) func(string) string {
		return func(key string) string { return env[key] }
	}
	run := func(args []string, env map[string]string, stdin string) (int, string, string) {
		var out, errOut bytes.Buffer
		code := Run(args, envFn(env), strings.NewReader(stdin), &out, &errOut)
		return code, out.String(), errOut.String()
	}
	return f, func(args []string, env map[string]string, stdin string) (int, string, string) {
		return run(append([]string{"--url", srv.URL}, args...), env, stdin)
	}
}

func TestProjectListCompactAndJSON(t *testing.T) {
	_, run := newFake(t)

	code, out, errOut := run([]string{"project", "list", "--compact"}, nil, "")
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, errOut)
	}
	if out != "PROJ | active | Project\n" {
		t.Fatalf("out = %q", out)
	}

	code, out, errOut = run([]string{"project", "list", "--json"}, nil, "")
	if code != 0 {
		t.Fatalf("json code = %d, stderr = %s", code, errOut)
	}
	var page struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal([]byte(out), &page); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d", len(page.Items))
	}
}

func TestIssueListBuildsFilters(t *testing.T) {
	f, run := newFake(t)
	code, out, errOut := run([]string{
		"issue", "list", "--project", "PROJ", "--status", "todo",
		"--priority", "high", "--limit", "2", "--compact",
	}, nil, "")
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, errOut)
	}
	if out != "PROJ-1 | todo | high | Implement auth\n" {
		t.Fatalf("out = %q", out)
	}
	if !strings.Contains(f.lastQuery, "status=todo") ||
		!strings.Contains(f.lastQuery, "priority=high") ||
		!strings.Contains(f.lastQuery, "limit=2") {
		t.Fatalf("query = %q", f.lastQuery)
	}
}

func TestIssueCreateAndCommentIdempotency(t *testing.T) {
	f, run := newFake(t)

	code, out, errOut := run([]string{
		"issue", "create", "--project", "PROJ", "--title", "Implement auth",
		"--idempotency-key", "fixed-key", "--compact",
	}, nil, "")
	if code != 0 {
		t.Fatalf("create code = %d, stderr = %s", code, errOut)
	}
	if out != "PROJ-1 | todo | high | Implement auth\n" {
		t.Fatalf("out = %q", out)
	}
	if f.lastKey != "fixed-key" {
		t.Fatalf("idempotency key = %q", f.lastKey)
	}
	if f.lastBody["title"] != "Implement auth" {
		t.Fatalf("body = %+v", f.lastBody)
	}

	f.lastKey = ""
	code, out, errOut = run([]string{
		"issue", "comment", "PROJ-1", "--body", "-", "--agent", "bot", "--compact",
	}, nil, "Started\n")
	if code != 0 {
		t.Fatalf("comment code = %d, stderr = %s", code, errOut)
	}
	if !strings.Contains(out, "comment-1") || !strings.Contains(out, "Started") {
		t.Fatalf("out = %q", out)
	}
	if f.lastKey == "" {
		t.Fatal("generated idempotency key is empty")
	}
	if f.lastBody["content"] != "Started" || f.lastBody["agent_id"] != "bot" {
		t.Fatalf("body = %+v", f.lastBody)
	}
}

func TestTokenAndClaimCompact(t *testing.T) {
	f, run := newFake(t)
	code, out, errOut := run(
		[]string{"issue", "claim", "PROJ-1", "--agent", "bot", "--compact"},
		map[string]string{"TABLON_CLI_TOKEN": "secret"},
		"",
	)
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, errOut)
	}
	if out != "PROJ-1 | bot\n" {
		t.Fatalf("out = %q", out)
	}
	if f.lastAuth != "Bearer secret" {
		t.Fatalf("auth = %q", f.lastAuth)
	}
}

func TestCommentUpdateAndDelete(t *testing.T) {
	_, run := newFake(t)
	code, out, errOut := run(
		[]string{"comment", "update", "comment-1", "--body", "Edited", "--compact"},
		nil, "",
	)
	if code != 0 {
		t.Fatalf("update code = %d, stderr = %s", code, errOut)
	}
	if !strings.Contains(out, "Edited") {
		t.Fatalf("out = %q", out)
	}

	code, out, errOut = run([]string{"comment", "delete", "comment-1"}, nil, "")
	if code != 0 {
		t.Fatalf("delete code = %d, stderr = %s", code, errOut)
	}
	if out != "deleted\n" {
		t.Fatalf("out = %q", out)
	}
}

func TestExitCodesAndUsage(t *testing.T) {
	_, run := newFake(t)

	if code, _, _ := run([]string{"unknown"}, nil, ""); code != 2 {
		t.Fatalf("unknown code = %d, want 2", code)
	}
	if code, _, _ := run([]string{"project", "list", "--json", "--compact"}, nil, ""); code != 2 {
		t.Fatalf("conflicting output code = %d, want 2", code)
	}
	if code, _, _ := run([]string{"project", "get", "MISSING"}, nil, ""); code != 1 {
		t.Fatalf("api error code = %d, want 1", code)
	}
	if code, out, _ := run([]string{"--help"}, nil, ""); code != 0 || !strings.Contains(out, "tablon") {
		t.Fatalf("help code = %d out = %q", code, out)
	}
	if code, out, _ := run([]string{"--version"}, nil, ""); code != 0 || !strings.Contains(out, "tablon dev") {
		t.Fatalf("version code = %d out = %q", code, out)
	}
	code, out, _ := run([]string{"version", "--json"}, nil, "")
	if code != 0 {
		t.Fatalf("version json code = %d", code)
	}
	var info map[string]string
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("version json: %v (%q)", err, out)
	}
	if info["version"] == "" || info["commit"] == "" || info["date"] == "" {
		t.Fatalf("version fields = %#v", info)
	}
}

func TestVersionInjected(t *testing.T) {
	oldVersion, oldCommit, oldDate := version.Version, version.Commit, version.Date
	t.Cleanup(func() {
		version.Version, version.Commit, version.Date = oldVersion, oldCommit, oldDate
	})
	version.Version, version.Commit, version.Date = "1.2.3", "abc123", "2026-11-24T10:00:00Z"

	_, run := newFake(t)

	code, out, _ := run([]string{"--version"}, nil, "")
	if code != 0 || !strings.Contains(out, "tablon 1.2.3") || !strings.Contains(out, "abc123") {
		t.Fatalf("version code = %d out = %q", code, out)
	}

	code, out, _ = run([]string{"version", "--json"}, nil, "")
	if code != 0 {
		t.Fatalf("version json code = %d", code)
	}
	var info map[string]string
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("version json: %v (%q)", err, out)
	}
	if info["version"] != "1.2.3" || info["commit"] != "abc123" || info["date"] != "2026-11-24T10:00:00Z" {
		t.Fatalf("version fields = %#v", info)
	}
}

func TestProjectAndEpicCommands(t *testing.T) {
	f, run := newFake(t)

	code, out, errOut := run([]string{"project", "get", "PROJ"}, nil, "")
	if code != 0 {
		t.Fatalf("project get code = %d, stderr = %s", code, errOut)
	}
	if !strings.Contains(out, "ID: PROJ") || !strings.Contains(out, "A project") {
		t.Fatalf("project get out = %q", out)
	}

	code, out, errOut = run([]string{
		"project", "create", "--name", "New project", "--identifier", "NEW", "--compact",
	}, nil, "")
	if code != 0 {
		t.Fatalf("project create code = %d, stderr = %s", code, errOut)
	}
	if out != "NEW | planning | New project\n" {
		t.Fatalf("project create out = %q", out)
	}

	code, out, errOut = run([]string{
		"project", "update", "PROJ", "--name", "Renamed", "--status", "active", "--compact",
	}, nil, "")
	if code != 0 {
		t.Fatalf("project update code = %d, stderr = %s", code, errOut)
	}
	if !strings.Contains(out, "Renamed") || f.lastBody["status"] != "active" {
		t.Fatalf("project update out = %q body = %+v", out, f.lastBody)
	}

	code, out, errOut = run([]string{"epic", "list", "--project", "PROJ", "--compact"}, nil, "")
	if code != 0 {
		t.Fatalf("epic list code = %d, stderr = %s", code, errOut)
	}
	if out != "epic-1 | backlog | Epic one\n" {
		t.Fatalf("epic list out = %q", out)
	}

	code, out, errOut = run([]string{"epic", "get", "epic-1", "--compact"}, nil, "")
	if code != 0 || !strings.Contains(out, "Epic one") {
		t.Fatalf("epic get code = %d out = %q stderr = %s", code, out, errOut)
	}

	code, out, errOut = run([]string{
		"epic", "create", "--project", "PROJ", "--title", "Epic two", "--compact",
	}, nil, "")
	if code != 0 || !strings.Contains(out, "Epic two") {
		t.Fatalf("epic create code = %d out = %q stderr = %s", code, out, errOut)
	}

	code, out, errOut = run([]string{
		"epic", "update", "epic-1", "--title", "Renamed epic", "--status", "in-progress", "--compact",
	}, nil, "")
	if code != 0 || !strings.Contains(out, "Renamed epic") {
		t.Fatalf("epic update code = %d out = %q stderr = %s", code, out, errOut)
	}
}

func TestIssueGetUpdateCommentsAndRelease(t *testing.T) {
	f, run := newFake(t)

	code, out, errOut := run([]string{"issue", "get", "PROJ-1", "--compact"}, nil, "")
	if code != 0 {
		t.Fatalf("issue get code = %d, stderr = %s", code, errOut)
	}
	if out != "PROJ-1 | todo | high | Implement auth\n" {
		t.Fatalf("issue get out = %q", out)
	}

	code, out, errOut = run([]string{
		"issue", "update", "PROJ-1", "--status", "in-progress", "--priority", "urgent", "--compact",
	}, nil, "")
	if code != 0 {
		t.Fatalf("issue update code = %d, stderr = %s", code, errOut)
	}
	if f.lastBody["status"] != "in-progress" || f.lastBody["priority"] != "urgent" {
		t.Fatalf("issue update body = %+v", f.lastBody)
	}
	if !strings.Contains(out, "in-progress") {
		t.Fatalf("issue update out = %q", out)
	}

	code, out, errOut = run([]string{"issue", "comments", "PROJ-1", "--compact"}, nil, "")
	if code != 0 || !strings.Contains(out, "Started") {
		t.Fatalf("issue comments code = %d out = %q stderr = %s", code, out, errOut)
	}

	code, out, errOut = run(
		[]string{"issue", "release", "PROJ-1", "--agent", "bot", "--compact"},
		nil,
		"",
	)
	if code != 0 {
		t.Fatalf("issue release code = %d, stderr = %s", code, errOut)
	}
	if out != "PROJ-1 | bot\n" {
		t.Fatalf("issue release out = %q", out)
	}
}
