package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/pablogtrz/tablon/migrations"
)

func TestCustomGradientMigrationUpgradesExistingProjects(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open raw sqlite: %v", err)
	}

	for _, name := range []string{
		"0001_init.sql",
		"0002_comments_and_idempotency.sql",
		"0003_fixed_width_timestamps.sql",
	} {
		body, err := migrations.FS.ReadFile(name)
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		if _, err := db.Exec(string(body)); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}

	if _, err := db.Exec(`
		CREATE TABLE schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		)
	`); err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}
	for _, version := range []int{1, 2, 3} {
		if _, err := db.Exec(
			`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`,
			version, "2026-11-24T10:00:00.000000000Z",
		); err != nil {
			t.Fatalf("record migration %d: %v", version, err)
		}
	}
	if _, err := db.Exec(`
		INSERT INTO projects (
			id, identifier, name, description, status, logo_gradient, logo_emoji,
			issue_counter, created_at, updated_at
		) VALUES ('legacy', 'LEGACY', 'Legacy', '', 'planning', 2, '🧭', 0, ?, ?)
	`, "2026-11-24T10:00:00.000000000Z", "2026-11-24T10:00:00.000000000Z"); err != nil {
		t.Fatalf("insert legacy project: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close raw sqlite: %v", err)
	}

	st, err := Open(dir)
	if err != nil {
		t.Fatalf("open migrated store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	ctx := context.Background()
	project, err := st.GetProject(ctx, "legacy")
	if err != nil {
		t.Fatalf("get legacy project: %v", err)
	}
	if project.LogoGradient != 2 || project.LogoGradientFrom != "" || project.LogoGradientTo != "" || project.LogoGradientAngle != 135 {
		t.Fatalf("legacy project gradient = %#v", project)
	}

	items, err := st.ListProjects(ctx, true)
	if err != nil {
		t.Fatalf("list legacy projects: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("projects = %d, want 1", len(items))
	}
	if items[0].LogoGradientFrom != "" || items[0].LogoGradientTo != "" || items[0].LogoGradientAngle != 135 {
		t.Fatalf("legacy list gradient = %#v", items[0])
	}
}
