package store

import (
	"context"
	"link_or_die/internal/apperror"
	"link_or_die/internal/domains/links"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkRepository struct {
	db *pgxpool.Pool
}

func CreateLinkRepo(db *pgxpool.Pool) *LinkRepository {
	return &LinkRepository{
		db: db,
	}
}

func (r *LinkRepository) Create(ctx context.Context, l *links.Link) (*links.Link, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO links (user_id, collection_id, url, normalized_url, url_hash, host, title)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, state, consecutive_failures, total_failures, resurrected_count, next_check_at, created_at, updated_at
	`,
		l.UserID, l.CollectionID, l.URL, l.NormalizedURL, l.URLHash, l.Host, l.Title,
	).Scan(&l.ID, &l.State, &l.ConsecutiveFailures, &l.TotalFailures, &l.ResurrectedCount,
		&l.NextCheckAt, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, translatePgError(err, "creating link")
	}
	return l, nil
}

func (r *LinkRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]links.Link, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, collection_id, url, normalized_url, url_hash, host, title,
		       state, last_outcome, last_status_code, last_error, cause_of_death,
		       consecutive_failures, total_failures, resurrected_count,
		       last_checked_at, last_ok_at, died_at, next_check_at, created_at, updated_at
		FROM links
		WHERE user_id = $1
		ORDER BY collection_id NULLS LAST, created_at DESC
	`, userID)
	if err != nil {
		return nil, translatePgError(err, "listing links")
	}
	defer rows.Close()

	var out []links.Link
	for rows.Next() {
		var l links.Link
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.CollectionID, &l.URL, &l.NormalizedURL, &l.URLHash,
			&l.Host, &l.Title, &l.State, &l.LastOutcome, &l.LastStatusCode, &l.LastError,
			&l.CauseOfDeath, &l.ConsecutiveFailures, &l.TotalFailures, &l.ResurrectedCount,
			&l.LastCheckedAt, &l.LastOKAt, &l.DiedAt, &l.NextCheckAt, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *LinkRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM links WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return translatePgError(err, "deleting link")
	}
	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
