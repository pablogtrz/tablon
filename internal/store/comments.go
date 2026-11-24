package store

import (
	"context"
	"database/sql"

	"github.com/pablogtrz/tablon/internal/domain"
)

func scanComment(row rowScanner) (domain.Comment, error) {
	var c domain.Comment
	var agentID sql.NullString
	var created, updated string
	if err := row.Scan(
		&c.ID, &c.IssueID, &c.Content, &agentID, &created, &updated,
	); err != nil {
		return domain.Comment{}, err
	}
	if agentID.Valid {
		c.AgentID = &agentID.String
	}
	c.CreatedAt = parseTime(created)
	c.UpdatedAt = parseTime(updated)
	return c, nil
}

func (s *Store) CreateComment(ctx context.Context, c domain.Comment) (domain.Comment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Comment{}, err
	}
	defer tx.Rollback()

	created, err := s.createCommentTx(ctx, tx, c)
	if err != nil {
		return domain.Comment{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Comment{}, err
	}
	return created, nil
}

func (s *Store) CreateCommentIdempotent(
	ctx context.Context,
	c domain.Comment,
	key, requestHash string,
) (domain.Comment, bool, error) {
	return idempotent(ctx, s.db, key, requestHash, 201, func(tx *sql.Tx) (domain.Comment, error) {
		return s.createCommentTx(ctx, tx, c)
	})
}

func (s *Store) createCommentTx(ctx context.Context, tx *sql.Tx, c domain.Comment) (domain.Comment, error) {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO comments (id, issue_id, content, agent_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		c.ID, c.IssueID, c.Content, c.AgentID, formatTime(c.CreatedAt), formatTime(c.UpdatedAt),
	)
	if err != nil {
		if isForeignKey(err) {
			return domain.Comment{}, domain.ErrNotFound
		}
		return domain.Comment{}, err
	}
	return c, nil
}

func (s *Store) GetComment(ctx context.Context, ref string) (domain.Comment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, issue_id, content, agent_id, created_at, updated_at
		FROM comments
		WHERE id = ?
	`, ref)
	c, err := scanComment(row)
	if err != nil {
		return domain.Comment{}, notFound(err)
	}
	return c, nil
}

func (s *Store) ListComments(
	ctx context.Context,
	issueID string,
	limit int,
	cursor string,
) (domain.Page[domain.Comment], error) {
	where := []string{"issue_id = ?"}
	args := []any{issueID}

	if cursor != "" {
		cur, err := decodeIssueCursor(cursor)
		if err != nil {
			return domain.Page[domain.Comment]{}, err
		}
		where = append(where, "(created_at, id) > (?, ?)")
		args = append(args, cur.Value, cur.ID)
	}

	if limit <= 0 {
		limit = 50
	}
	args = append(args, limit+1)

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, issue_id, content, agent_id, created_at, updated_at
		FROM comments
		WHERE `+join(where, " AND ")+`
		ORDER BY created_at ASC, id ASC
		LIMIT ?
	`, args...)
	if err != nil {
		return domain.Page[domain.Comment]{}, err
	}
	defer rows.Close()

	items := make([]domain.Comment, 0, limit+1)
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return domain.Page[domain.Comment]{}, err
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return domain.Page[domain.Comment]{}, err
	}

	page := domain.Page[domain.Comment]{Items: items}
	if len(items) > limit {
		last := items[limit-1]
		page.Items = items[:limit]
		page.NextCursor = encodeIssueCursor(issueCursor{
			Value: formatTime(last.CreatedAt),
			ID:    last.ID,
		})
	}
	return page, nil
}

func (s *Store) UpdateComment(ctx context.Context, c domain.Comment) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE comments
		SET content = ?, updated_at = ?
		WHERE id = ?
	`, c.Content, formatTime(c.UpdatedAt), c.ID)
	if err != nil {
		return err
	}
	return requireOne(res)
}

func (s *Store) DeleteComment(ctx context.Context, ref string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM comments WHERE id = ?`, ref)
	if err != nil {
		return err
	}
	return requireOne(res)
}
