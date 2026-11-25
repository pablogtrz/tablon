package store

import (
	"context"
	"testing"
	"time"

	"github.com/pablogtrz/tablon/internal/domain"
	"github.com/pablogtrz/tablon/migrations"
)

func TestFixedWidthTimestampsSortChronologically(t *testing.T) {
	early := time.Date(2026, 11, 24, 10, 0, 0, 100_000_000, time.UTC)
	late := time.Date(2026, 11, 24, 10, 0, 0, 120_000_000, time.UTC)

	if got := formatTime(early); got != "2026-11-24T10:00:00.100000000Z" {
		t.Fatalf("early format = %q", got)
	}
	if got := formatTime(late); got != "2026-11-24T10:00:00.120000000Z" {
		t.Fatalf("late format = %q", got)
	}
	if formatTime(early) >= formatTime(late) {
		t.Fatalf("timestamps are not ordered: %q >= %q", formatTime(early), formatTime(late))
	}
	if parsed := parseTime(formatTime(late)); !parsed.Equal(late) {
		t.Fatalf("round trip = %s, want %s", parsed, late)
	}
}

func TestProjectListOrdersFixedWidthTimestamps(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	early := time.Date(2026, 11, 24, 10, 0, 0, 100_000_000, time.UTC)
	late := time.Date(2026, 11, 24, 10, 0, 0, 120_000_000, time.UTC)
	for _, p := range []domain.Project{
		{ID: "early", Identifier: "EARLY", Name: "Early", Status: domain.ProjectPlanning, CreatedAt: early, UpdatedAt: early},
		{ID: "late", Identifier: "LATE", Name: "Late", Status: domain.ProjectPlanning, CreatedAt: late, UpdatedAt: late},
	} {
		if err := st.CreateProject(context.Background(), p); err != nil {
			t.Fatalf("create %s: %v", p.Identifier, err)
		}
	}

	items, err := st.ListProjects(context.Background(), true)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(items) != 2 || items[0].ID != "late" || items[1].ID != "early" {
		t.Fatalf("order = %+v, want late then early", items)
	}
}

func TestTimestampMigrationNormalizesLegacyRows(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if _, err := st.db.Exec(`
		INSERT INTO projects (
			id, identifier, name, description, status, logo_gradient, logo_emoji,
			issue_counter, created_at, updated_at
		) VALUES (?, ?, ?, '', 'planning', 0, '', 0, ?, ?)
	`, "legacy", "LEGACY", "Legacy", "2026-11-24T10:00:00.1Z", "2026-11-24T10:00:00.12Z"); err != nil {
		t.Fatalf("insert legacy project: %v", err)
	}

	body, err := migrations.FS.ReadFile("0003_fixed_width_timestamps.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := st.db.Exec(string(body)); err != nil {
		t.Fatalf("apply normalization: %v", err)
	}

	var created, updated string
	if err := st.db.QueryRow(
		`SELECT created_at, updated_at FROM projects WHERE id = 'legacy'`,
	).Scan(&created, &updated); err != nil {
		t.Fatalf("read timestamps: %v", err)
	}
	if created != "2026-11-24T10:00:00.100000000Z" || updated != "2026-11-24T10:00:00.120000000Z" {
		t.Fatalf("timestamps = %q, %q", created, updated)
	}
}
