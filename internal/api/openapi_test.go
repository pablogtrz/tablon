package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pablogtrz/tablon/internal/config"
	"github.com/pablogtrz/tablon/internal/domain"
	"github.com/pablogtrz/tablon/internal/events"
	"github.com/pablogtrz/tablon/internal/service"
	"github.com/pablogtrz/tablon/internal/store"
)

func loadSpec(t *testing.T) (*openapi3.T, []byte) {
	t.Helper()
	path := filepath.Join("..", "..", "docs", "api", "openapi.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	spec, err := openapi3.NewLoader().LoadFromData(raw)
	if err != nil {
		t.Fatalf("load spec: %v", err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		t.Fatalf("validate spec: %v", err)
	}
	return spec, raw
}

func asJSON(t *testing.T, value any) any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	return out
}

func validateComponent(t *testing.T, spec *openapi3.T, name string, value any) {
	t.Helper()
	ref := spec.Components.Schemas[name]
	if ref == nil || ref.Value == nil {
		t.Fatalf("missing schema %s", name)
	}
	if err := ref.Value.VisitJSON(asJSON(t, value)); err != nil {
		t.Fatalf("schema %s: %v", name, err)
	}
}

func TestOpenAPIContract(t *testing.T) {
	spec, raw := loadSpec(t)

	var s Server
	routes := s.apiRoutes()
	routeSet := make(map[string]bool, len(routes))
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		routeSet[key] = true
		item := spec.Paths.Find(route.Pattern)
		if item == nil {
			t.Errorf("route %s is missing from the spec", key)
			continue
		}
		if item.GetOperation(route.Method) == nil {
			t.Errorf("route %s is missing from the spec", key)
		}
	}

	for path, item := range spec.Paths.Map() {
		if path == "/openapi.yaml" {
			continue
		}
		for _, method := range []string{
			http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete,
			http.MethodPut, http.MethodHead,
		} {
			op := item.GetOperation(method)
			if op == nil {
				continue
			}
			if !routeSet[method+" "+path] {
				t.Errorf("spec operation %s %s has no registered route", method, path)
			}
			if op.OperationID == "" {
				t.Errorf("%s %s has no operationId", method, path)
			}
			if op.Summary == "" {
				t.Errorf("%s %s has no summary", method, path)
			}
			if len(op.Tags) == 0 {
				t.Errorf("%s %s has no tags", method, path)
			}
			if op.Responses == nil || op.Responses.Len() == 0 {
				t.Errorf("%s %s has no responses", method, path)
			}
			hasSuccess := false
			for code := range op.Responses.Map() {
				if strings.HasPrefix(code, "2") {
					hasSuccess = true
					break
				}
			}
			if !hasSuccess {
				t.Errorf("%s %s has no success response", method, path)
			}
		}
	}

	now := time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)
	project := domain.Project{
		ID:           "01a20a51-31ec-73eb-b322-d0f6f3640534",
		Identifier:   "TABLON",
		Name:         "Tablon",
		Description:  "Lightweight issue tracking.",
		Status:       domain.ProjectActive,
		LogoGradient: 5,
		LogoEmoji:    "🧭",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	issue := domain.Issue{
		ID:          "01a20a51-320b-7a6b-8197-e47fca889efb",
		Identifier:  "TABLON-25",
		ProjectID:   project.ID,
		Title:       "Idle mode",
		Description: "Activity gate.",
		Status:      domain.IssueBacklog,
		Priority:    domain.PriorityHigh,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	comment := domain.Comment{
		ID:        "01a20a51-330b-7a6b-8197-e47fca889efb",
		IssueID:   issue.ID,
		Content:   "Ready.",
		CreatedAt: now,
		UpdatedAt: now,
	}

	validateComponent(t, spec, "Health", map[string]any{"status": "ok"})
	validateComponent(t, spec, "Ready", map[string]any{"status": "ready"})
	validateComponent(t, spec, "Project", project)
	validateComponent(t, spec, "ProjectList", map[string]any{
		"items": []domain.ProjectListItem{{
			Project:    project,
			EpicCount:  1,
			IssueCount: 2,
			DoneCount:  1,
		}},
	})
	validateComponent(t, spec, "EpicList", map[string]any{
		"items": []domain.Epic{{
			ID:          "01a20a51-31ef-7655-add4-c1f034c13bab",
			ProjectID:   project.ID,
			Title:       "M2",
			Description: "Collaboration.",
			Status:      domain.EpicDone,
			CreatedAt:   now,
			UpdatedAt:   now,
		}},
	})
	validateComponent(t, spec, "Issue", issue)
	validateComponent(t, spec, "IssuePage", domain.Page[domain.Issue]{
		Items:      []domain.Issue{issue},
		NextCursor: "cursor",
	})
	validateComponent(t, spec, "CommentPage", domain.Page[domain.Comment]{
		Items: []domain.Comment{comment},
	})
	validateComponent(t, spec, "SystemStatus", map[string]any{
		"version":      "0.1.0",
		"auth_enabled": false,
		"database":     "ok",
		"idle": map[string]any{
			"enabled":         true,
			"state":           "active",
			"timeout_seconds": 300,
			"last_activity":   now.Format(time.RFC3339Nano),
			"idle_since":      nil,
		},
	})
	validateComponent(t, spec, "Error", map[string]any{
		"error": map[string]any{"code": "not_found", "message": "resource not found"},
	})

	// Serve the spec and validate a real 401 error envelope against it.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	hub := events.NewHub(logger)
	t.Cleanup(hub.Close)
	cfg := config.Config{Host: "127.0.0.1", DataDir: t.TempDir(), APIToken: "secret", LogLevel: slog.LevelError}
	srv := httptest.NewServer(New(cfg, service.New(st, hub), hub, nil, logger, nil, raw))
	t.Cleanup(srv.Close)

	res, err := srv.Client().Get(srv.URL + "/openapi.yaml")
	if err != nil {
		t.Fatalf("get spec: %v", err)
	}
	served, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("spec status = %d", res.StatusCode)
	}
	if string(served) != string(raw) {
		t.Fatal("served spec differs from docs/api/openapi.yaml")
	}
	if got := res.Header.Get("Content-Type"); got != "application/yaml; charset=utf-8" {
		t.Fatalf("spec content type = %q", got)
	}

	res, err = srv.Client().Get(srv.URL + "/api/v1/projects")
	if err != nil {
		t.Fatalf("get unauthorized: %v", err)
	}
	errorBody, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", res.StatusCode)
	}
	var errorValue any
	if err := json.Unmarshal(errorBody, &errorValue); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if err := spec.Components.Schemas["Error"].Value.VisitJSON(errorValue); err != nil {
		t.Fatalf("401 body does not match Error schema: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/projects", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("new validation request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "application/json")
	res, err = srv.Client().Do(req)
	if err != nil {
		t.Fatalf("validation request: %v", err)
	}
	validationBody, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("validation status = %d: %s", res.StatusCode, validationBody)
	}
	var validationValue any
	if err := json.Unmarshal(validationBody, &validationValue); err != nil {
		t.Fatalf("decode validation error: %v", err)
	}
	if err := spec.Components.Schemas["Error"].Value.VisitJSON(validationValue); err != nil {
		t.Fatalf("422 body does not match Error schema: %v", err)
	}
}
