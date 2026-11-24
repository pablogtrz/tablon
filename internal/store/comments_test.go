package store

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pablogtrz/tablon/internal/domain"
)

func newTestIssue(t *testing.T, st *Store) domain.Issue {
	t.Helper()
	ctx := context.Background()
	now := time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)
	project := domain.Project{
		ID:           "project-1",
		Identifier:   "TST",
		Name:         "Test",
		Status:       domain.ProjectActive,
		LogoGradient: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := st.CreateProject(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}
	i, err := st.CreateIssue(ctx, domain.Issue{
		ID:        "issue-1",
		ProjectID: project.ID,
		Title:     "Issue",
		Status:    domain.IssueTodo,
		Priority:  domain.PriorityNone,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	return i
}

func TestCommentsCRUDPaginationAndCascade(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	issue := newTestIssue(t, st)

	agent := "agent-a"
	base := time.Date(2026, 11, 24, 11, 0, 0, 0, time.UTC)
	ids := []string{"c1", "c2", "c3"}
	for n, id := range ids {
		created := base.Add(time.Duration(n) * time.Minute)
		comment, err := st.CreateComment(ctx, domain.Comment{
			ID:        id,
			IssueID:   issue.ID,
			Content:   "body " + id,
			AgentID:   &agent,
			CreatedAt: created,
			UpdatedAt: created,
		})
		if err != nil {
			t.Fatalf("create comment %s: %v", id, err)
		}
		if comment.IssueID != issue.ID {
			t.Fatalf("issue id = %q", comment.IssueID)
		}
	}

	page, err := st.ListComments(ctx, issue.ID, 2, "")
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(page.Items) != 2 || page.Items[0].ID != "c1" || page.Items[1].ID != "c2" {
		t.Fatalf("unexpected first page: %+v", page.Items)
	}
	if page.NextCursor == "" {
		t.Fatal("missing next cursor")
	}

	next, err := st.ListComments(ctx, issue.ID, 2, page.NextCursor)
	if err != nil {
		t.Fatalf("list next page: %v", err)
	}
	if len(next.Items) != 1 || next.Items[0].ID != "c3" {
		t.Fatalf("unexpected next page: %+v", next.Items)
	}

	got, err := st.GetComment(ctx, "c1")
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}
	got.Content = "edited"
	got.UpdatedAt = base.Add(time.Hour)
	if err := st.UpdateComment(ctx, got); err != nil {
		t.Fatalf("update comment: %v", err)
	}
	again, err := st.GetComment(ctx, "c1")
	if err != nil {
		t.Fatalf("get edited comment: %v", err)
	}
	if again.Content != "edited" {
		t.Fatalf("content = %q", again.Content)
	}

	if err := st.DeleteComment(ctx, "c2"); err != nil {
		t.Fatalf("delete comment: %v", err)
	}
	if _, err := st.GetComment(ctx, "c2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleted comment error = %v", err)
	}

	if err := st.DeleteIssue(ctx, issue.ID); err != nil {
		t.Fatalf("delete issue: %v", err)
	}
	var count int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM comments WHERE issue_id = ?`, issue.ID).Scan(&count); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if count != 0 {
		t.Fatalf("comments = %d, want cascade delete", count)
	}
}

func TestIdempotentCreatesReplayAndConflict(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	issue := newTestIssue(t, st)

	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	comment := domain.Comment{
		ID:        "comment-idem",
		IssueID:   issue.ID,
		Content:   "once",
		CreatedAt: now,
		UpdatedAt: now,
	}
	first, replayed, err := st.CreateCommentIdempotent(ctx, comment, "key-1", "hash-1")
	if err != nil || replayed {
		t.Fatalf("first create: replayed=%v err=%v", replayed, err)
	}
	second, replayed, err := st.CreateCommentIdempotent(ctx, comment, "key-1", "hash-1")
	if err != nil || !replayed {
		t.Fatalf("replay: replayed=%v err=%v", replayed, err)
	}
	if second.ID != first.ID {
		t.Fatalf("replayed id = %q, want %q", second.ID, first.ID)
	}
	if _, _, err := st.CreateCommentIdempotent(ctx, comment, "key-1", "different"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("hash conflict error = %v", err)
	}

	var count int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM comments`).Scan(&count); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if count != 1 {
		t.Fatalf("comments = %d, want 1", count)
	}

	newIssue := domain.Issue{
		ID:        "issue-idem",
		ProjectID: issue.ProjectID,
		Title:     "Idempotent issue",
		Status:    domain.IssueTodo,
		Priority:  domain.PriorityNone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	created, replayed, err := st.CreateIssueIdempotent(ctx, newIssue, "key-2", "hash-2")
	if err != nil || replayed {
		t.Fatalf("first issue create: replayed=%v err=%v", replayed, err)
	}
	again, replayed, err := st.CreateIssueIdempotent(ctx, newIssue, "key-2", "hash-2")
	if err != nil || !replayed {
		t.Fatalf("issue replay: replayed=%v err=%v", replayed, err)
	}
	if again.Identifier != created.Identifier {
		t.Fatalf("replayed identifier = %q, want %q", again.Identifier, created.Identifier)
	}
}

func TestIdempotentConcurrentCreate(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	issue := newTestIssue(t, st)

	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	comment := domain.Comment{
		ID:        "comment-concurrent",
		IssueID:   issue.ID,
		Content:   "once",
		CreatedAt: now,
		UpdatedAt: now,
	}

	type result struct {
		id       string
		replayed bool
		err      error
	}
	const workers = 4
	results := make(chan result, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			created, replayed, err := st.CreateCommentIdempotent(
				context.Background(), comment, "key-concurrent", "hash-concurrent",
			)
			results <- result{id: created.ID, replayed: replayed, err: err}
		}()
	}
	wg.Wait()
	close(results)

	created, replayed := 0, 0
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent create: %v", result.err)
		}
		if result.id != comment.ID {
			t.Fatalf("id = %q, want %q", result.id, comment.ID)
		}
		if result.replayed {
			replayed++
		} else {
			created++
		}
	}
	if created != 1 || replayed != workers-1 {
		t.Fatalf("created=%d replayed=%d, want 1 and %d", created, replayed, workers-1)
	}

	var count int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM comments`).Scan(&count); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if count != 1 {
		t.Fatalf("comments = %d, want 1", count)
	}
}
