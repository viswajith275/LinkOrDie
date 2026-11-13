package store

import (
	"context"
	"link_or_die/internal/apperror"
	"link_or_die/internal/domains/links"
	"time"

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

func (r *LinkRepository) ListByCollection(ctx context.Context, userID uuid.UUID, collectionID uuid.UUID) ([]links.Link, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, collection_id, url, normalized_url, url_hash, host, title,
		       state, last_outcome, last_status_code, last_error, cause_of_death,
		       consecutive_failures, total_failures, resurrected_count,
		       last_checked_at, last_ok_at, died_at, next_check_at, created_at, updated_at
		FROM links
		WHERE collection_id = $1 AND user_id = $2
		ORDER BY created_at DESC
	`, collectionID, userID)
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

func (r *LinkRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*links.Link, error) {
	var l *links.Link;

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, collection_id, url, normalized_url, url_hash, host, title,
		       state, last_outcome, last_status_code, last_error, cause_of_death,
		       consecutive_failures, total_failures, resurrected_count,
		       last_checked_at, last_ok_at, died_at, next_check_at, created_at, updated_at
		FROM links WHERE id = $1 AND user_id = $2
		`, id, userID).Scan(&l.ID, &l.UserID, &l.CollectionID, &l.URL, &l.NormalizedURL, &l.URLHash,
		&l.Host, &l.Title, &l.State, &l.LastOutcome, &l.LastStatusCode, &l.LastError,
		&l.CauseOfDeath, &l.ConsecutiveFailures, &l.TotalFailures, &l.ResurrectedCount,
		&l.LastCheckedAt, &l.LastOKAt, &l.DiedAt, &l.NextCheckAt, &l.CreatedAt, &l.UpdatedAt,)

	if err != nil {
		return nil, translatePgError(err, "finding link by id: ")
	}

	return l, err
}


// ClaimDue atomically leases a batch of due links
func (r *LinkRepository) ClaimDue(ctx context.Context, limit int, lease time.Duration) ([]links.DueLink, error) {

	rows, err := r.db.Query(ctx, `
		WITH due AS (
			SELECT id
			FROM links
			WHERE next_check_at <= now()
			ORDER BY next_check_at
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE links l
		SET next_check_at = now() + make_interval(secs => $2)
		FROM due
		WHERE l.id = due.id
		RETURNING l.id, l.normalized_url, l.host, l.next_check_at
	`, limit, lease.Seconds())
	if err != nil {
		return nil, translatePgError(err, "claiming due links")
	}
	defer rows.Close()

	var out []links.DueLink
	for rows.Next() {
		var link links.DueLink
		if err := rows.Scan(&link.ID, &link.URL, &link.Host, &link.LeaseUntil); err != nil {
			return nil, err
		}
		out = append(out, link)
	}

	return out, rows.Err()
}

// Requeue moves a link forward only if this worker still owns the claim
func (r *LinkRepository) Requeue(ctx context.Context, id uuid.UUID, leaseUntil, at time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE links
		SET next_check_at = $3
		WHERE id = $1 AND next_check_at = $2
	`, id, leaseUntil, at)
	if err != nil {
		return translatePgError(err, "requeuing link")
	}
	return nil
}
