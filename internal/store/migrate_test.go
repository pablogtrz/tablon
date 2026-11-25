package store

import "testing"

func TestMigrateIsIdempotentAndEnablesWAL(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if err := st.Migrate(); err != nil {
		t.Fatalf("second migrate: %v", err)
	}

	var count int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 4 {
		t.Fatalf("migrations = %d, want 4", count)
	}

	var mode string
	if err := st.db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatalf("journal mode: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("journal mode = %q, want wal", mode)
	}
}
