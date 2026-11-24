package store

import (
	"context"
	"strings"
	"testing"

	"github.com/pablogtrz/tablon/internal/domain"
)

func TestListQueryPlansUseIndexes(t *testing.T) {
	st, project, issue := seedBench(t, 100, 10)
	ctx := context.Background()

	issuePlan := explain(t, st, `
		EXPLAIN QUERY PLAN
		SELECT id, identifier, project_id, epic_id, title, description, status,
		       priority, agent_id, created_at, updated_at
		FROM issues
		WHERE project_id = ? AND status = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, project.ID, "todo", 50)
	if !strings.Contains(issuePlan, "idx_issues") {
		t.Fatalf("issue plan does not use an issues index:\n%s", issuePlan)
	}

	commentPlan := explain(t, st, `
		EXPLAIN QUERY PLAN
		SELECT id, issue_id, content, agent_id, created_at, updated_at
		FROM comments
		WHERE issue_id = ? AND (created_at, id) > (?, ?)
		ORDER BY created_at ASC, id ASC
		LIMIT ?
	`, issue.ID, "2026-01-01T00:00:00Z", "comment-00000", 50)
	if !strings.Contains(commentPlan, "idx_comments_issue_id") {
		t.Fatalf("comment plan does not use idx_comments_issue_id:\n%s", commentPlan)
	}

	if _, err := st.ListIssues(ctx, project.ID, domain.IssueQuery{Limit: 50}); err != nil {
		t.Fatalf("list issues: %v", err)
	}
}

func explain(t *testing.T, st *Store, query string, args ...any) string {
	t.Helper()
	rows, err := st.db.Query(query, args...)
	if err != nil {
		t.Fatalf("explain: %v", err)
	}
	defer rows.Close()

	var parts []string
	for rows.Next() {
		var id, parent, notUsed int
		var detail string
		if err := rows.Scan(&id, &parent, &notUsed, &detail); err != nil {
			t.Fatalf("scan explain: %v", err)
		}
		parts = append(parts, detail)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("explain rows: %v", err)
	}
	return strings.Join(parts, "\n")
}
