package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/pablogtrz/tablon/internal/domain"
)

type Store struct {
	db   *sql.DB
	path string
}

func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	path := filepath.Join(dataDir, "app.db")
	dsn := "file:" + filepath.ToSlash(path) +
		"?_pragma=busy_timeout(5000)" +
		"&_pragma=foreign_keys(ON)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=synchronous(NORMAL)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	_ = os.Chmod(path, 0o600)

	s := &Store{db: db, path: path}
	if err := s.Migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) Path() string {
	return s.path
}

// EnterIdle releases the SQLite page cache and the idle writer connection.
// The database stays open and becomes available again on the next query.
func (s *Store) EnterIdle(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `PRAGMA shrink_memory`); err != nil {
		return err
	}
	s.db.SetMaxIdleConns(0)
	return nil
}

// ExitIdle restores the single idle writer connection used by the store.
func (s *Store) ExitIdle() {
	s.db.SetMaxIdleConns(1)
}

// timestampFormat is fixed-width so TEXT columns sort chronologically.
// RFC3339Nano trims trailing zeros, which breaks lexicographic ordering when
// one fractional value is a prefix of another (for example .1Z vs .12Z).
const timestampFormat = "2006-01-02T15:04:05.000000000Z07:00"

func formatTime(t time.Time) string {
	return t.UTC().Format(timestampFormat)
}

func parseTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, value)
	}
	return t.UTC()
}

func isUnique(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "constraint failed: unique")
}

func isForeignKey(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "foreign key constraint failed")
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProject(row rowScanner) (domain.Project, error) {
	var p domain.Project
	var created, updated string
	if err := row.Scan(
		&p.ID, &p.Identifier, &p.Name, &p.Description, &p.Status,
		&p.LogoGradient, &p.LogoEmoji, &p.LogoGradientFrom, &p.LogoGradientTo,
		&p.LogoGradientAngle, &created, &updated,
	); err != nil {
		return domain.Project{}, err
	}
	p.CreatedAt = parseTime(created)
	p.UpdatedAt = parseTime(updated)
	return p, nil
}

func scanProjectListItem(row rowScanner) (domain.ProjectListItem, error) {
	var item domain.ProjectListItem
	var created, updated string
	if err := row.Scan(
		&item.ID, &item.Identifier, &item.Name, &item.Description, &item.Status,
		&item.LogoGradient, &item.LogoEmoji, &item.LogoGradientFrom, &item.LogoGradientTo,
		&item.LogoGradientAngle, &created, &updated,
		&item.EpicCount, &item.IssueCount, &item.DoneCount,
	); err != nil {
		return domain.ProjectListItem{}, err
	}
	item.CreatedAt = parseTime(created)
	item.UpdatedAt = parseTime(updated)
	return item, nil
}

func scanEpic(row rowScanner) (domain.Epic, error) {
	var e domain.Epic
	var created, updated string
	if err := row.Scan(
		&e.ID, &e.ProjectID, &e.Title, &e.Description, &e.Status, &created, &updated,
	); err != nil {
		return domain.Epic{}, err
	}
	e.CreatedAt = parseTime(created)
	e.UpdatedAt = parseTime(updated)
	return e, nil
}

func scanIssue(row rowScanner) (domain.Issue, error) {
	var i domain.Issue
	var epicID, agentID sql.NullString
	var created, updated string
	if err := row.Scan(
		&i.ID, &i.Identifier, &i.ProjectID, &epicID, &i.Title, &i.Description,
		&i.Status, &i.Priority, &agentID, &created, &updated,
	); err != nil {
		return domain.Issue{}, err
	}
	if epicID.Valid {
		i.EpicID = &epicID.String
	}
	if agentID.Valid {
		i.AgentID = &agentID.String
	}
	i.CreatedAt = parseTime(created)
	i.UpdatedAt = parseTime(updated)
	return i, nil
}

func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}
