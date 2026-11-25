package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pablogtrz/tablon/internal/domain"
	"github.com/pablogtrz/tablon/internal/events"
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

func TestProjectCustomGradientRoundTrip(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	angle := 45

	p, err := svc.CreateProject(ctx, CreateProjectInput{
		Name:              "Custom gradient",
		LogoGradientFrom:  "#FFA443",
		LogoGradientTo:    "#8B5CF6",
		LogoGradientAngle: &angle,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if p.LogoGradientFrom != "#ffa443" || p.LogoGradientTo != "#8b5cf6" || p.LogoGradientAngle != 45 {
		t.Fatalf("custom gradient = %#v", p)
	}

	got, err := svc.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if got.LogoGradientFrom != "#ffa443" || got.LogoGradientTo != "#8b5cf6" || got.LogoGradientAngle != 45 {
		t.Fatalf("stored custom gradient = %#v", got)
	}

	list, err := svc.ListProjects(ctx, true)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	found := false
	for _, item := range list {
		if item.ID != p.ID {
			continue
		}
		found = true
		if item.LogoGradientFrom != "#ffa443" || item.LogoGradientTo != "#8b5cf6" || item.LogoGradientAngle != 45 {
			t.Fatalf("listed custom gradient = %#v", item)
		}
	}
	if !found {
		t.Fatal("project missing from list")
	}

	empty := ""
	cleared, err := svc.UpdateProject(ctx, p.ID, UpdateProjectInput{
		LogoGradientFrom: &empty,
		LogoGradientTo:   &empty,
	})
	if err != nil {
		t.Fatalf("clear custom gradient: %v", err)
	}
	if cleared.LogoGradientFrom != "" || cleared.LogoGradientTo != "" || cleared.LogoGradientAngle != 135 {
		t.Fatalf("custom gradient was not cleared: %#v", cleared)
	}
	got, err = svc.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("re-read cleared project: %v", err)
	}
	if got.LogoGradientFrom != "" || got.LogoGradientTo != "" || got.LogoGradientAngle != 135 {
		t.Fatalf("cleared custom gradient = %#v", got)
	}

	from := "#112233"
	to := "#445566"
	nextAngle := 225
	updated, err := svc.UpdateProject(ctx, p.ID, UpdateProjectInput{
		LogoGradientFrom:  &from,
		LogoGradientTo:    &to,
		LogoGradientAngle: &nextAngle,
	})
	if err != nil {
		t.Fatalf("update custom gradient: %v", err)
	}
	if updated.LogoGradientFrom != from || updated.LogoGradientTo != to || updated.LogoGradientAngle != nextAngle {
		t.Fatalf("updated custom gradient = %#v", updated)
	}
	got, err = svc.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("re-read updated project: %v", err)
	}
	if got.LogoGradientFrom != from || got.LogoGradientTo != to || got.LogoGradientAngle != nextAngle {
		t.Fatalf("persisted custom gradient = %#v", got)
	}

	partialFrom := "#778899"
	partial, err := svc.UpdateProject(ctx, p.ID, UpdateProjectInput{LogoGradientFrom: &partialFrom})
	if err != nil {
		t.Fatalf("partial custom gradient update: %v", err)
	}
	if partial.LogoGradientFrom != partialFrom || partial.LogoGradientTo != to {
		t.Fatalf("partial update lost the end colour: %#v", partial)
	}
}

func TestProjectCustomGradientValidation(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Name: "Validation"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	empty := ""
	badAngle := 361
	tests := []struct {
		name  string
		field string
		run   func() error
	}{
		{
			name:  "create short colour",
			field: "logo_gradient_from",
			run: func() error {
				_, err := svc.CreateProject(ctx, CreateProjectInput{
					Name:             "Bad colour",
					LogoGradientFrom: "#fff",
					LogoGradientTo:   "#000000",
				})
				return err
			},
		},
		{
			name:  "create start only",
			field: "logo_gradient_to",
			run: func() error {
				_, err := svc.CreateProject(ctx, CreateProjectInput{
					Name:             "Start only",
					LogoGradientFrom: "#aabbcc",
				})
				return err
			},
		},
		{
			name:  "create end only",
			field: "logo_gradient_from",
			run: func() error {
				_, err := svc.CreateProject(ctx, CreateProjectInput{
					Name:           "End only",
					LogoGradientTo: "#aabbcc",
				})
				return err
			},
		},
		{
			name:  "create bad angle",
			field: "logo_gradient_angle",
			run: func() error {
				_, err := svc.CreateProject(ctx, CreateProjectInput{
					Name:              "Bad angle",
					LogoGradientFrom:  "#aabbcc",
					LogoGradientTo:    "#ddeeff",
					LogoGradientAngle: &badAngle,
				})
				return err
			},
		},
		{
			name:  "update one-sided",
			field: "logo_gradient_to",
			run: func() error {
				from := "#aabbcc"
				_, err := svc.UpdateProject(ctx, p.ID, UpdateProjectInput{
					LogoGradientFrom: &from,
					LogoGradientTo:   &empty,
				})
				return err
			},
		},
		{
			name:  "update bad angle",
			field: "logo_gradient_angle",
			run: func() error {
				_, err := svc.UpdateProject(ctx, p.ID, UpdateProjectInput{LogoGradientAngle: &badAngle})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			ve, ok := domain.AsValidation(err)
			if !ok {
				t.Fatalf("error = %v, want validation error", err)
			}
			if ve.Field != test.field {
				t.Fatalf("field = %q, want %q", ve.Field, test.field)
			}
		})
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

func TestCommentsValidationIdempotencyAndUpdates(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	p := mustProject(t, svc, "Comments", "CMT")
	issue, err := svc.CreateIssue(ctx, p.ID, CreateIssueInput{Title: "Issue"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	if _, err := svc.CreateComment(ctx, issue.ID, CreateCommentInput{Content: "   "}); err == nil {
		t.Fatal("empty comment accepted")
	}
	long := make([]byte, 20001)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := svc.CreateComment(ctx, issue.ID, CreateCommentInput{Content: string(long)}); err == nil {
		t.Fatal("oversized comment accepted")
	}

	agent := "agent-a"
	first, err := svc.CreateComment(ctx, issue.Identifier, CreateCommentInput{
		Content: "hello **world**",
		AgentID: &agent,
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if first.AgentID == nil || *first.AgentID != agent {
		t.Fatalf("agent = %v", first.AgentID)
	}

	replayed, wasReplay, err := svc.CreateCommentIdempotent(ctx, issue.Identifier, CreateCommentInput{
		Content: "hello **world**",
		AgentID: &agent,
	}, "key-1", "hash-1")
	if err != nil || wasReplay {
		t.Fatalf("first idempotent create: replayed=%v err=%v", wasReplay, err)
	}
	again, wasReplay, err := svc.CreateCommentIdempotent(ctx, issue.Identifier, CreateCommentInput{
		Content: "hello **world**",
		AgentID: &agent,
	}, "key-1", "hash-1")
	if err != nil || !wasReplay {
		t.Fatalf("replay: replayed=%v err=%v", wasReplay, err)
	}
	if again.ID != replayed.ID {
		t.Fatalf("replayed id = %q, want %q", again.ID, replayed.ID)
	}

	edited := "updated"
	updated, err := svc.UpdateComment(ctx, first.ID, UpdateCommentInput{Content: &edited})
	if err != nil {
		t.Fatalf("update comment: %v", err)
	}
	if updated.Content != edited {
		t.Fatalf("content = %q", updated.Content)
	}
	if !updated.UpdatedAt.After(first.UpdatedAt) {
		t.Fatalf("updated_at did not advance: %v <= %v", updated.UpdatedAt, first.UpdatedAt)
	}

	page, err := svc.ListComments(ctx, issue.Identifier, 0, "")
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("comments = %d, want 2", len(page.Items))
	}

	if err := svc.DeleteComment(ctx, first.ID); err != nil {
		t.Fatalf("delete comment: %v", err)
	}
	if _, err := svc.UpdateComment(ctx, first.ID, UpdateCommentInput{Content: &edited}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("update deleted comment error = %v", err)
	}
}

func TestIdempotencyKeyValidation(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	p := mustProject(t, svc, "Keys", "KEY")
	if _, _, err := svc.CreateIssueIdempotent(ctx, p.ID, CreateIssueInput{Title: "X"}, "bad key", "h"); err == nil {
		t.Fatal("space in idempotency key accepted")
	}
	long := make([]byte, 129)
	for i := range long {
		long[i] = 'k'
	}
	if _, _, err := svc.CreateIssueIdempotent(ctx, p.ID, CreateIssueInput{Title: "X"}, string(long), "h"); err == nil {
		t.Fatal("oversized idempotency key accepted")
	}
}

type countingPublisher struct {
	mu     sync.Mutex
	events []events.Event
}

func (p *countingPublisher) Publish(event events.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, event)
}

func (p *countingPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.events)
}

func TestIdempotentReplayDoesNotPublishDuplicateEvents(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	publisher := &countingPublisher{}
	svc := New(st, publisher)
	ctx := context.Background()
	project, err := svc.CreateProject(ctx, CreateProjectInput{Name: "Events", Identifier: "EVS"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	issue, replayed, err := svc.CreateIssueIdempotent(
		ctx, project.ID, CreateIssueInput{Title: "Idempotent issue"}, "issue-key", "issue-hash",
	)
	if err != nil || replayed {
		t.Fatalf("create issue: replayed=%v err=%v", replayed, err)
	}
	if publisher.count() != 1 {
		t.Fatalf("events after create = %d, want 1", publisher.count())
	}
	if _, replayed, err := svc.CreateIssueIdempotent(
		ctx, project.ID, CreateIssueInput{Title: "Idempotent issue"}, "issue-key", "issue-hash",
	); err != nil || !replayed {
		t.Fatalf("replay issue: replayed=%v err=%v", replayed, err)
	}
	if publisher.count() != 1 {
		t.Fatalf("events after issue replay = %d, want 1", publisher.count())
	}

	if _, replayed, err := svc.CreateCommentIdempotent(
		ctx, issue.Identifier, CreateCommentInput{Content: "once"}, "comment-key", "comment-hash",
	); err != nil || replayed {
		t.Fatalf("create comment: replayed=%v err=%v", replayed, err)
	}
	if publisher.count() != 2 {
		t.Fatalf("events after comment create = %d, want 2", publisher.count())
	}
	if _, replayed, err := svc.CreateCommentIdempotent(
		ctx, issue.Identifier, CreateCommentInput{Content: "once"}, "comment-key", "comment-hash",
	); err != nil || !replayed {
		t.Fatalf("replay comment: replayed=%v err=%v", replayed, err)
	}
	if publisher.count() != 2 {
		t.Fatalf("events after comment replay = %d, want 2", publisher.count())
	}
}
