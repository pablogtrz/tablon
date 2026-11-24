package api_test

import (
	"fmt"
	"net/http"
	"testing"
)

func seedBenchAPI(b *testing.B, issues int) (*testAPI, string, string) {
	b.Helper()
	a := newTestAPI(b, "")

	code, body := a.do(b, http.MethodPost, "/api/v1/projects", map[string]any{"name": "Bench"}, nil)
	if code != http.StatusCreated {
		b.Fatalf("create project = %d: %s", code, body)
	}
	project := decode[struct {
		ID         string `json:"id"`
		Identifier string `json:"identifier"`
	}](b, body)

	var firstIssue string
	for i := 0; i < issues; i++ {
		code, body := a.do(b, http.MethodPost, "/api/v1/projects/"+project.ID+"/issues", map[string]any{
			"title": fmt.Sprintf("Issue %d", i),
		}, nil)
		if code != http.StatusCreated {
			b.Fatalf("create issue %d = %d: %s", i, code, body)
		}
		issue := decode[struct {
			Identifier string `json:"identifier"`
		}](b, body)
		if firstIssue == "" {
			firstIssue = issue.Identifier
		}
	}
	return a, project.ID, firstIssue
}

func BenchmarkAPIListIssues(b *testing.B) {
	a, projectID, _ := seedBenchAPI(b, 100)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		code, body := a.do(b, http.MethodGet, "/api/v1/projects/"+projectID+"/issues?limit=50", nil, nil)
		if code != http.StatusOK {
			b.Fatalf("list issues = %d: %s", code, body)
		}
	}
}

func BenchmarkAPIGetIssue(b *testing.B) {
	a, _, issueRef := seedBenchAPI(b, 10)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		code, body := a.do(b, http.MethodGet, "/api/v1/issues/"+issueRef, nil, nil)
		if code != http.StatusOK {
			b.Fatalf("get issue = %d: %s", code, body)
		}
	}
}

func BenchmarkAPIGetProject(b *testing.B) {
	a, projectID, _ := seedBenchAPI(b, 1)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		code, body := a.do(b, http.MethodGet, "/api/v1/projects/"+projectID, nil, nil)
		if code != http.StatusOK {
			b.Fatalf("get project = %d: %s", code, body)
		}
	}
}

func BenchmarkAPICreateComment(b *testing.B) {
	a, _, issueRef := seedBenchAPI(b, 1)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		code, body := a.do(b, http.MethodPost, "/api/v1/issues/"+issueRef+"/comments", map[string]any{
			"content": "Benchmark comment",
		}, nil)
		if code != http.StatusCreated {
			b.Fatalf("create comment = %d: %s", code, body)
		}
	}
}

func BenchmarkAPIIdempotentComment(b *testing.B) {
	a, _, issueRef := seedBenchAPI(b, 1)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		code, body := a.do(b, http.MethodPost, "/api/v1/issues/"+issueRef+"/comments", map[string]any{
			"content": "Benchmark comment",
		}, map[string]string{"Idempotency-Key": fmt.Sprintf("bench-%d", i)})
		if code != http.StatusCreated {
			b.Fatalf("create comment = %d: %s", code, body)
		}
	}
}

func BenchmarkAPISystemStatus(b *testing.B) {
	a := newTestAPI(b, "")
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		code, body := a.do(b, http.MethodGet, "/api/v1/system/status", nil, nil)
		if code != http.StatusOK {
			b.Fatalf("status = %d: %s", code, body)
		}
	}
}
