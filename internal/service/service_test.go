package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pablogtrz/tablon/internal/domain"
	"github.com/pablogtrz/tablon/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	svc := New(st)
	var tick int
	svc.now = func() time.Time {
		tick++
		return time.Date(2026, 1, 1, 0, 0, 0, tick, time.UTC)
	}
	return svc
}

func mustProject(t *testing.T, svc *Service, name, identifier string) domain.Project {
	t.Helper()
	p, err := svc.CreateProject(context.Background(), CreateProjectInput{
		Name:       name,
		Identifier: identifier,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return p
}

func TestCreateProjectDefaultsAndIdentifierCollision(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	p, err := svc.CreateProject(ctx, CreateProjectInput{Name: "Test Project"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if p.Identifier != "TESTPROJ" {
		t.Fatalf("identifier = %q, want TESTPROJ", p.Identifier)
	}
	if p.Status != domain.ProjectPlanning {
		t.Fatalf("status = %q, want planning", p.Status)
	}
	if p.LogoGradient < 0 || p.LogoGradient > 7 {
		t.Fatalf("logo gradient = %d, want 0..7", p.LogoGradient)
	}

	p2, err := svc.CreateProject(ctx, CreateProjectInput{Name: "Test Project"})
	if err != nil {
		t.Fatalf("create second project: %v", err)
	}
	if p2.Identifier == p.Identifier {
		t.Fatalf("second identifier was not made unique")
	}
}

func TestExplicitIdentifierConflict(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	mustProject(t, svc, "One", "TAB")
	_, err := svc.CreateProject(ctx, CreateProjectInput{Name: "Two", Identifier: "TAB"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error = %v, want conflict", err)
	}
}

func TestProjectIdentifierIsImmutable(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	p := mustProject(t, svc, "One", "TAB")

	next := "OTHER"
	_, err := svc.UpdateProject(ctx, p.ID, UpdateProjectInput{Identifier: &next})
	if _, ok := domain.AsValidation(err); !ok {
		t.Fatalf("error = %v, want validation error", err)
	}
}

func TestIssueNumberingSurvivesDelete(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	p := mustProject(t, svc, "Numbers", "NUM")

	first, err := svc.CreateIssue(ctx, p.ID, CreateIssueInput{Title: "First"})
	if err != nil {
		t.Fatalf("create first issue: %v", err)
	}
	if first.Identifier != "NUM-1" {
		t.Fatalf("first identifier = %q, want NUM-1", first.Identifier)
	}
	if err := svc.DeleteIssue(ctx, first.ID); err != nil {
		t.Fatalf("delete first issue: %v", err)
	}
	second, err := svc.CreateIssue(ctx, p.ID, CreateIssueInput{Title: "Second"})
	if err != nil {
		t.Fatalf("create second issue: %v", err)
	}
	if second.Identifier != "NUM-2" {
		t.Fatalf("second identifier = %q, want NUM-2", second.Identifier)
	}
}

func TestEpicMustBelongToProject(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	a := mustProject(t, svc, "A", "AA")
	b := mustProject(t, svc, "B", "BB")

	epic, err := svc.CreateEpic(ctx, a.ID, CreateEpicInput{Title: "Epic"})
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}
	_, err = svc.CreateIssue(ctx, b.ID, CreateIssueInput{Title: "Issue", EpicID: &epic.ID})
	if _, ok := domain.AsValidation(err); !ok {
		t.Fatalf("error = %v, want validation error", err)
	}
}

func TestClaimAndRelease(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	p := mustProject(t, svc, "Claims", "CLM")
	issue, err := svc.CreateIssue(ctx, p.ID, CreateIssueInput{Title: "Claim me"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	claimed, err := svc.ClaimIssue(ctx, issue.ID, "agent-a")
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if claimed.AgentID == nil || *claimed.AgentID != "agent-a" {
		t.Fatalf("agent = %v, want agent-a", claimed.AgentID)
	}

	if _, err := svc.ClaimIssue(ctx, issue.ID, "agent-b"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("second claim error = %v, want conflict", err)
	}
	if _, err := svc.ReleaseIssue(ctx, issue.ID, "agent-b"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("wrong release error = %v, want conflict", err)
	}
	released, err := svc.ReleaseIssue(ctx, issue.ID, "agent-a")
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	if released.AgentID != nil {
		t.Fatalf("agent = %v, want nil", released.AgentID)
	}
}

func TestProjectDeleteCascadesButKeepsIdentifier(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	p := mustProject(t, svc, "Cascade", "CAS")
	epic, err := svc.CreateEpic(ctx, p.ID, CreateEpicInput{Title: "Epic"})
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}
	issue, err := svc.CreateIssue(ctx, p.ID, CreateIssueInput{Title: "Issue", EpicID: &epic.ID})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	if err := svc.DeleteProject(ctx, p.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if _, err := svc.GetIssue(ctx, issue.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("issue error = %v, want not found", err)
	}
	if _, err := svc.GetEpic(ctx, epic.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("epic error = %v, want not found", err)
	}
	if _, err := svc.CreateProject(ctx, CreateProjectInput{Name: "Again", Identifier: "CAS"}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("reused identifier error = %v, want conflict", err)
	}
}

func TestIssueListFiltersSortsAndPaginates(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	p := mustProject(t, svc, "List", "LST")

	for i, title := range []string{"Low", "Urgent", "High"} {
		priority := []string{"low", "urgent", "high"}[i]
		if _, err := svc.CreateIssue(ctx, p.ID, CreateIssueInput{Title: title, Priority: priority}); err != nil {
			t.Fatalf("create issue: %v", err)
		}
	}

	page, err := svc.ListIssues(ctx, p.ID, domain.IssueQuery{Sort: "priority", Order: "desc", Limit: 2})
	if err != nil {
		t.Fatalf("list issues: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(page.Items))
	}
	if page.Items[0].Title != "Urgent" || page.Items[1].Title != "High" {
		t.Fatalf("unexpected order: %q, %q", page.Items[0].Title, page.Items[1].Title)
	}
	if page.NextCursor == "" {
		t.Fatal("next cursor is empty")
	}

	next, err := svc.ListIssues(ctx, p.ID, domain.IssueQuery{Sort: "priority", Order: "desc", Limit: 2, Cursor: page.NextCursor})
	if err != nil {
		t.Fatalf("list next page: %v", err)
	}
	if len(next.Items) != 1 || next.Items[0].Title != "Low" {
		t.Fatalf("unexpected next page: %+v", next.Items)
	}
}
