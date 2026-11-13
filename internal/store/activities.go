package store

import (
	"context"
	"link_or_die/internal/domains/links"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// repo for checks, transitions
type ActivityRepository struct {
	db *pgxpool.Pool
}

func CreateActivityRepo(db *pgxpool.Pool) *LinkRepository {
	return &LinkRepository{
		db: db,
	}
}

func (r *ActivityRepository) ListChecks(ctx context.Context, userID uuid.UUID, limit int) ([]links.Check, error) {

	rows, err := r.db.Query(ctx, `
			SELECT c.id, c.link_id, c.checked_at, c.outcome, c.status_code, c.final_url, c.methode, c.elapsed_ms, c.error
			FROM checks c JOIN links l ON c.link_id = l.id AND l.user_id = $1 ORDER BY c.checked_at DESC LIMIT $2
		`, userID, limit)

	if err != nil {
		return nil, translatePgError(err, "claiming due links")
	}
	defer rows.Close()

	var out []links.Check
	for rows.Next() {
		var check links.Check
		if err := rows.Scan(&check.ID, &check.LinkID, &check.CheckedAt, &check.Outcome, &check.StatusCode, &check.FinalURL, &check.Method, &check.ElapsedMS, &check.Error); err != nil {
			return nil, err
		}
		out = append(out, check)
	}

	return out, rows.Err()
}

func (r *ActivityRepository) ListTransitions(ctx context.Context, userID uuid.UUID, limit int) ([]links.Transition, error) {

	rows, err := r.db.Query(ctx, `
			SELECT t.id, t.link_id, t.from_state, t.to_state, t.outcome, t.cause, t.occured_at, t.acknowledged_at
			FROM transitions t JOIN links l ON t.link_id = l.id AND l.user_id = $1 ORDER BY c.occured_at DESC LIMIT $2
		`, userID, limit)

	if err != nil {
		return nil, translatePgError(err, "claiming due links")
	}
	defer rows.Close()

	var out []links.Transition
	for rows.Next() {
		var transition links.Transition
		if err := rows.Scan(&transition.ID, &transition.LinkID, &transition.FromState, &transition.ToState, &transition.Outcome, &transition.Cause, &transition.OccurredAt, &transition.AcknowledgedAt); err != nil {
			return nil, err
		}
		out = append(out, transition)
	}

	return out, rows.Err()
}
