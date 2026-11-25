package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pablogtrz/tablon/internal/domain"
)

func (s *Store) IdentifierExists(ctx context.Context, identifier string) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM project_identifiers WHERE identifier = ?`, identifier,
	).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) CreateProject(ctx context.Context, p domain.Project) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO project_identifiers (identifier, created_at) VALUES (?, ?)`,
		p.Identifier, formatTime(p.CreatedAt),
	); err != nil {
		if isUnique(err) {
			return domain.ErrConflict
		}
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO projects (
			id, identifier, name, description, status, logo_gradient, logo_emoji,
			logo_gradient_from, logo_gradient_to, logo_gradient_angle,
			issue_counter, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
	`,
		p.ID, p.Identifier, p.Name, p.Description, p.Status, p.LogoGradient,
		p.LogoEmoji, p.LogoGradientFrom, p.LogoGradientTo, p.LogoGradientAngle,
		formatTime(p.CreatedAt), formatTime(p.UpdatedAt),
	); err != nil {
		if isUnique(err) {
			return domain.ErrConflict
		}
		return err
	}

	return tx.Commit()
}

func (s *Store) GetProject(ctx context.Context, ref string) (domain.Project, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, identifier, name, description, status, logo_gradient, logo_emoji,
		       logo_gradient_from, logo_gradient_to, logo_gradient_angle,
		       created_at, updated_at
		FROM projects
		WHERE id = ? OR identifier = ?
	`, ref, ref)
	p, err := scanProject(row)
	if err != nil {
		return domain.Project{}, notFound(err)
	}
	return p, nil
}

func (s *Store) ListProjects(ctx context.Context, includeArchived bool) ([]domain.ProjectListItem, error) {
	query := `
		SELECT p.id, p.identifier, p.name, p.description, p.status, p.logo_gradient,
		       p.logo_emoji, p.logo_gradient_from, p.logo_gradient_to, p.logo_gradient_angle,
		       p.created_at, p.updated_at,
		       (SELECT COUNT(*) FROM epics e WHERE e.project_id = p.id),
		       (SELECT COUNT(*) FROM issues i WHERE i.project_id = p.id),
		       (SELECT COUNT(*) FROM issues i WHERE i.project_id = p.id AND i.status = 'done')
		FROM projects p`
	if !includeArchived {
		query += ` WHERE p.status != 'archived'`
	}
	query += ` ORDER BY p.created_at DESC, p.id DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.ProjectListItem, 0)
	for rows.Next() {
		item, err := scanProjectListItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) UpdateProject(ctx context.Context, p domain.Project) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE projects
		SET name = ?, description = ?, status = ?, logo_gradient = ?, logo_emoji = ?,
		    logo_gradient_from = ?, logo_gradient_to = ?, logo_gradient_angle = ?, updated_at = ?
		WHERE id = ?
	`,
		p.Name, p.Description, p.Status, p.LogoGradient, p.LogoEmoji,
		p.LogoGradientFrom, p.LogoGradientTo, p.LogoGradientAngle,
		formatTime(p.UpdatedAt), p.ID,
	)
	if err != nil {
		return err
	}
	return requireOne(res)
}

func (s *Store) DeleteProject(ctx context.Context, ref string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM projects WHERE id = ? OR identifier = ?`, ref, ref,
	)
	if err != nil {
		return err
	}
	return requireOne(res)
}

func requireOne(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) CreateEpic(ctx context.Context, e domain.Epic) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO epics (id, project_id, title, description, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		e.ID, e.ProjectID, e.Title, e.Description, e.Status,
		formatTime(e.CreatedAt), formatTime(e.UpdatedAt),
	)
	if isForeignKey(err) {
		return domain.ErrNotFound
	}
	return err
}

func (s *Store) ListEpics(ctx context.Context, projectID string) ([]domain.Epic, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM epics
		WHERE project_id = ?
		ORDER BY created_at DESC, id DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Epic, 0)
	for rows.Next() {
		e, err := scanEpic(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (s *Store) GetEpic(ctx context.Context, ref string) (domain.Epic, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM epics
		WHERE id = ?
	`, ref)
	e, err := scanEpic(row)
	if err != nil {
		return domain.Epic{}, notFound(err)
	}
	return e, nil
}

func (s *Store) UpdateEpic(ctx context.Context, e domain.Epic) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE epics
		SET title = ?, description = ?, status = ?, updated_at = ?
		WHERE id = ?
	`, e.Title, e.Description, e.Status, formatTime(e.UpdatedAt), e.ID)
	if err != nil {
		return err
	}
	return requireOne(res)
}

func (s *Store) DeleteEpic(ctx context.Context, ref string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM epics WHERE id = ?`, ref)
	if err != nil {
		return err
	}
	return requireOne(res)
}

func (s *Store) CreateIssue(ctx context.Context, i domain.Issue) (domain.Issue, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Issue{}, err
	}
	defer tx.Rollback()

	created, err := s.createIssueTx(ctx, tx, i)
	if err != nil {
		return domain.Issue{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Issue{}, err
	}
	return created, nil
}

func (s *Store) CreateIssueIdempotent(
	ctx context.Context,
	i domain.Issue,
	key, requestHash string,
) (domain.Issue, bool, error) {
	return idempotent(ctx, s.db, key, requestHash, 201, func(tx *sql.Tx) (domain.Issue, error) {
		return s.createIssueTx(ctx, tx, i)
	})
}

func (s *Store) createIssueTx(ctx context.Context, tx *sql.Tx, i domain.Issue) (domain.Issue, error) {
	var identifier string
	var counter int
	err := tx.QueryRowContext(ctx, `
		UPDATE projects
		SET issue_counter = issue_counter + 1, updated_at = ?
		WHERE id = ?
		RETURNING identifier, issue_counter
	`, formatTime(i.CreatedAt), i.ProjectID).Scan(&identifier, &counter)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Issue{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Issue{}, err
	}

	i.Identifier = fmt.Sprintf("%s-%d", identifier, counter)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO issues (
			id, identifier, project_id, epic_id, title, description, status,
			priority, agent_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		i.ID, i.Identifier, i.ProjectID, i.EpicID, i.Title, i.Description,
		i.Status, i.Priority, i.AgentID, formatTime(i.CreatedAt), formatTime(i.UpdatedAt),
	)
	if err != nil {
		if isForeignKey(err) {
			return domain.Issue{}, domain.ErrNotFound
		}
		return domain.Issue{}, err
	}
	return i, nil
}

func (s *Store) GetIssue(ctx context.Context, ref string) (domain.Issue, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, identifier, project_id, epic_id, title, description, status,
		       priority, agent_id, created_at, updated_at
		FROM issues
		WHERE id = ? OR identifier = ?
	`, ref, ref)
	i, err := scanIssue(row)
	if err != nil {
		return domain.Issue{}, notFound(err)
	}
	return i, nil
}

func (s *Store) ListIssues(ctx context.Context, projectID string, q domain.IssueQuery) (domain.Page[domain.Issue], error) {
	sortExpr, sortIsPriority, err := issueSortExpr(q.Sort)
	if err != nil {
		return domain.Page[domain.Issue]{}, err
	}
	dir := "DESC"
	if q.Order == "asc" {
		dir = "ASC"
	}

	where := []string{"project_id = ?"}
	args := []any{projectID}

	if q.Status != "" {
		where = append(where, "status = ?")
		args = append(args, q.Status)
	}
	if q.Priority != "" {
		where = append(where, "priority = ?")
		args = append(args, q.Priority)
	}
	if q.EpicID != "" {
		where = append(where, "epic_id = ?")
		args = append(args, q.EpicID)
	}
	if q.Search != "" {
		where = append(where, "(title LIKE ? OR description LIKE ?)")
		like := "%" + q.Search + "%"
		args = append(args, like, like)
	}
	if q.Cursor != "" {
		cur, err := decodeIssueCursor(q.Cursor)
		if err != nil {
			return domain.Page[domain.Issue]{}, err
		}
		op := "<"
		if dir == "ASC" {
			op = ">"
		}
		if sortIsPriority {
			where = append(where, fmt.Sprintf("(%s, id) %s (CAST(? AS INTEGER), ?)", sortExpr, op))
		} else {
			where = append(where, fmt.Sprintf("(%s, id) %s (?, ?)", sortExpr, op))
		}
		args = append(args, cur.Value, cur.ID)
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	args = append(args, limit+1)

	query := fmt.Sprintf(`
		SELECT id, identifier, project_id, epic_id, title, description, status,
		       priority, agent_id, created_at, updated_at
		FROM issues
		WHERE %s
		ORDER BY %s %s, id %s
		LIMIT ?
	`, join(where, " AND "), sortExpr, dir, dir)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.Page[domain.Issue]{}, err
	}
	defer rows.Close()

	items := make([]domain.Issue, 0, limit+1)
	for rows.Next() {
		i, err := scanIssue(rows)
		if err != nil {
			return domain.Page[domain.Issue]{}, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return domain.Page[domain.Issue]{}, err
	}

	page := domain.Page[domain.Issue]{Items: items}
	if len(items) > limit {
		last := items[limit-1]
		page.Items = items[:limit]
		page.NextCursor = encodeIssueCursor(issueCursor{
			Value: issueSortValue(last, q.Sort),
			ID:    last.ID,
		})
	}
	return page, nil
}

func (s *Store) UpdateIssue(ctx context.Context, i domain.Issue) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE issues
		SET title = ?, description = ?, status = ?, priority = ?, epic_id = ?, updated_at = ?
		WHERE id = ?
	`, i.Title, i.Description, i.Status, i.Priority, i.EpicID, formatTime(i.UpdatedAt), i.ID)
	if err != nil {
		if isForeignKey(err) {
			return domain.ErrNotFound
		}
		return err
	}
	return requireOne(res)
}

func (s *Store) DeleteIssue(ctx context.Context, ref string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM issues WHERE id = ? OR identifier = ?`, ref, ref,
	)
	if err != nil {
		return err
	}
	return requireOne(res)
}

func (s *Store) ClaimIssue(ctx context.Context, ref, agentID string, at time.Time) (domain.Issue, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE issues
		SET agent_id = ?, updated_at = ?
		WHERE (id = ? OR identifier = ?)
		  AND (agent_id IS NULL OR agent_id = '')
	`, agentID, formatTime(at), ref, ref)
	if err != nil {
		return domain.Issue{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Issue{}, err
	}
	if n == 0 {
		if _, err := s.GetIssue(ctx, ref); err != nil {
			return domain.Issue{}, err
		}
		return domain.Issue{}, domain.ErrConflict
	}
	return s.GetIssue(ctx, ref)
}

func (s *Store) ReleaseIssue(ctx context.Context, ref, agentID string, at time.Time) (domain.Issue, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE issues
		SET agent_id = NULL, updated_at = ?
		WHERE (id = ? OR identifier = ?)
		  AND agent_id = ?
	`, formatTime(at), ref, ref, agentID)
	if err != nil {
		return domain.Issue{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Issue{}, err
	}
	if n == 0 {
		if _, err := s.GetIssue(ctx, ref); err != nil {
			return domain.Issue{}, err
		}
		return domain.Issue{}, domain.ErrConflict
	}
	return s.GetIssue(ctx, ref)
}
