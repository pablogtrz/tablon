package store

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pablogtrz/tablon/internal/domain"
)

func seedBench(tb testing.TB, issueCount, commentCount int) (*Store, domain.Project, domain.Issue) {
	tb.Helper()
	st, err := Open(tb.TempDir())
	if err != nil {
		tb.Fatalf("open store: %v", err)
	}
	tb.Cleanup(func() { _ = st.Close() })

	ctx := context.Background()
	now := time.Now().UTC()
	project := domain.Project{
		ID:         "bench-project",
		Identifier: "BENCH",
		Name:       "Benchmark",
		Status:     domain.ProjectActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := st.CreateProject(ctx, project); err != nil {
		tb.Fatalf("create project: %v", err)
	}

	first := domain.Issue{
		ID:         "issue-00000",
		Identifier: "BENCH-1",
		ProjectID:  project.ID,
		Title:      "Issue 0",
		Status:     domain.IssueBacklog,
		Priority:   domain.PriorityNone,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if issueCount == 0 {
		return st, project, first
	}

	statuses := []domain.IssueStatus{
		domain.IssueBacklog, domain.IssueTodo, domain.IssueInProgress,
		domain.IssueInReview, domain.IssueDone, domain.IssueCanceled,
	}
	priorities := []domain.Priority{
		domain.PriorityNone, domain.PriorityLow, domain.PriorityMedium,
		domain.PriorityHigh, domain.PriorityUrgent,
	}

	tx, err := st.db.Begin()
	if err != nil {
		tb.Fatalf("begin seed: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO issues (
			id, identifier, project_id, epic_id, title, description, status,
			priority, agent_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tb.Fatalf("prepare issues: %v", err)
	}
	defer stmt.Close()

	for i := 0; i < issueCount; i++ {
		created := now.Add(time.Duration(i) * time.Millisecond)
		issue := domain.Issue{
			ID:         fmt.Sprintf("issue-%05d", i),
			Identifier: fmt.Sprintf("BENCH-%d", i+1),
			ProjectID:  project.ID,
			Title:      fmt.Sprintf("Issue %d", i),
			Status:     statuses[i%len(statuses)],
			Priority:   priorities[i%len(priorities)],
			CreatedAt:  created,
			UpdatedAt:  created,
		}
		if i == 0 {
			first = issue
		}
		if _, err := stmt.Exec(
			issue.ID, issue.Identifier, issue.ProjectID, nil, issue.Title, "",
			issue.Status, issue.Priority, nil, formatTime(issue.CreatedAt), formatTime(issue.UpdatedAt),
		); err != nil {
			tb.Fatalf("seed issue %d: %v", i, err)
		}
	}

	if _, err := tx.Exec(`UPDATE projects SET issue_counter = ? WHERE id = ?`, issueCount, project.ID); err != nil {
		tb.Fatalf("update issue counter: %v", err)
	}

	if commentCount > 0 {
		cstmt, err := tx.Prepare(`
			INSERT INTO comments (id, issue_id, content, agent_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			tb.Fatalf("prepare comments: %v", err)
		}
		defer cstmt.Close()
		for i := 0; i < commentCount; i++ {
			created := now.Add(time.Duration(i) * time.Millisecond)
			if _, err := cstmt.Exec(
				fmt.Sprintf("comment-%05d", i), first.ID, fmt.Sprintf("Comment %d", i), nil,
				formatTime(created), formatTime(created),
			); err != nil {
				tb.Fatalf("seed comment %d: %v", i, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		tb.Fatalf("commit seed: %v", err)
	}
	return st, project, first
}

func BenchmarkListIssues(b *testing.B) {
	st, project, _ := seedBench(b, 10000, 0)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := st.ListIssues(ctx, project.ID, domain.IssueQuery{Limit: 50}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListIssuesFiltered(b *testing.B) {
	st, project, _ := seedBench(b, 10000, 0)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := st.ListIssues(ctx, project.ID, domain.IssueQuery{
			Status: "in-progress",
			Limit:  50,
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListIssuesDeepCursor(b *testing.B) {
	st, project, _ := seedBench(b, 10000, 0)
	ctx := context.Background()
	page, err := st.ListIssues(ctx, project.ID, domain.IssueQuery{Limit: 50})
	if err != nil {
		b.Fatal(err)
	}
	q := domain.IssueQuery{Limit: 50, Cursor: page.NextCursor}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := st.ListIssues(ctx, project.ID, q); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetIssueByIdentifier(b *testing.B) {
	st, _, _ := seedBench(b, 10000, 0)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := st.GetIssue(ctx, "BENCH-5000"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateIssue(b *testing.B) {
	st, project, _ := seedBench(b, 1, 0)
	ctx := context.Background()
	now := time.Now().UTC()
	var counter atomic.Int64
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := counter.Add(1)
		if _, err := st.CreateIssue(ctx, domain.Issue{
			ID:        fmt.Sprintf("new-issue-%d", n),
			ProjectID: project.ID,
			Title:     "New issue",
			Status:    domain.IssueTodo,
			Priority:  domain.PriorityMedium,
			CreatedAt: now,
			UpdatedAt: now,
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateIssueStatus(b *testing.B) {
	st, _, issue := seedBench(b, 1, 0)
	ctx := context.Background()
	issue, err := st.GetIssue(ctx, issue.ID)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			issue.Status = domain.IssueInProgress
		} else {
			issue.Status = domain.IssueDone
		}
		issue.UpdatedAt = time.Now().UTC()
		if err := st.UpdateIssue(ctx, issue); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClaimIssue(b *testing.B) {
	st, _, issue := seedBench(b, 1, 0)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := st.ClaimIssue(ctx, issue.ID, "bench-agent", time.Now().UTC()); err != nil {
			b.Fatal(err)
		}
		if _, err := st.ReleaseIssue(ctx, issue.ID, "bench-agent", time.Now().UTC()); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClaimIssueContended(b *testing.B) {
	st, _, issue := seedBench(b, 1, 0)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := st.ClaimIssue(ctx, issue.ID, "bench-agent", time.Now().UTC()); err != nil {
				if !errors.Is(err, domain.ErrConflict) {
					b.Errorf("claim: %v", err)
					return
				}
				continue
			}
			if _, err := st.ReleaseIssue(ctx, issue.ID, "bench-agent", time.Now().UTC()); err != nil {
				b.Errorf("release: %v", err)
				return
			}
		}
	})
}

func BenchmarkListComments(b *testing.B) {
	st, _, issue := seedBench(b, 1, 5000)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := st.ListComments(ctx, issue.ID, 50, ""); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetProject(b *testing.B) {
	st, project, _ := seedBench(b, 1, 0)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := st.GetProject(ctx, project.ID); err != nil {
			b.Fatal(err)
		}
	}
}
