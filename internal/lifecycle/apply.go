package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"time"

	"link_or_die/internal/apperror"
	"link_or_die/internal/checker"
	"link_or_die/internal/domains/links"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrStaleClaim = errors.New("stale check claim")

func Apply(
	ctx context.Context,
	tx pgx.Tx,
	linkID uuid.UUID,
	result checker.Result,
	checkedAt time.Time,
	leaseUntil time.Time,
) (Decision, error) {
	if result.Canceled {
		return Decision{}, context.Canceled
	}
	if !result.Outcome.IsValid() {
		return Decision{}, fmt.Errorf("invalid check outcome %q", result.Outcome)
	}
	if checkedAt.IsZero() {
		checkedAt = time.Now()
	}

	var currentState string
	var consecutiveFailures int

	query := `
		SELECT state::text, consecutive_failures
		FROM links
		WHERE id = $1`
	args := []any{linkID}

	if !leaseUntil.IsZero() {
		query += ` AND next_check_at = $2`
		args = append(args, leaseUntil)
	}
	query += ` FOR UPDATE`

	// fetching current state and consecutive failures (Also checks if the link exists)
	err := tx.QueryRow(ctx, query, args...).Scan(&currentState, &consecutiveFailures)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if !leaseUntil.IsZero() {
				return Decision{}, ErrStaleClaim
			}
			return Decision{}, fmt.Errorf("apply check: %w", apperror.ErrNotFound)
		}
		return Decision{}, fmt.Errorf("lock link: %w", err)
	}

	// decide next state and other based on it
	decision := Decide(links.State(currentState), consecutiveFailures, result.Outcome)

	var finalURL *string
	if result.FinalURL != "" {
		finalURL = &result.FinalURL
	}

	var checkError *string
	if result.Error != "" {
		checkError = &result.Error
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO checks (
			link_id, checked_at, outcome, status_code, final_url, method, elapsed_ms, error
		)
		VALUES ($1, $2, $3::check_outcome, $4, $5, $6, $7, $8)
	`,
		linkID,
		checkedAt,
		string(result.Outcome),
		result.StatusCode,
		finalURL,
		result.Method,
		int32(result.Elapsed/time.Millisecond),
		checkError,
	)
	if err != nil {
		return decision, fmt.Errorf("insert check: %w", err)
	}

	nextCheckAt := checkedAt.Add(NextDelay(decision.NextState, result.Outcome))

	_, err = tx.Exec(ctx, `
		UPDATE links
		SET
			state = $2::link_state,
			last_outcome = $3::check_outcome,
			last_status_code = $4,
			last_error = CASE WHEN $5::bool THEN NULL ELSE $6::text END,
			cause_of_death = CASE
				WHEN $7::bool THEN $3::check_outcome
				WHEN $8::bool THEN NULL
				ELSE cause_of_death
			END,
			consecutive_failures = $9,
			total_failures = total_failures + CASE WHEN $10::bool THEN 1 ELSE 0 END,
			resurrected_count = resurrected_count + CASE WHEN $8::bool THEN 1 ELSE 0 END,
			last_checked_at = $11,
			last_ok_at = CASE WHEN $5::bool THEN $11 ELSE last_ok_at END,
			died_at = CASE
				WHEN $7::bool THEN $11
				WHEN $8::bool THEN NULL
				ELSE died_at
			END,
			next_check_at = $12
		WHERE id = $1
	`,
		linkID,
		string(decision.NextState),
		string(result.Outcome),
		result.StatusCode,
		result.Outcome.IsSuccess(),
		checkError,
		decision.Died,
		decision.Resurrected,
		decision.ConsecutiveFailures,
		decision.Failed,
		checkedAt,
		nextCheckAt,
	)
	if err != nil {
		return decision, fmt.Errorf("update link: %w", err)
	}

	if decision.Changed {
		var cause *string
		if decision.Cause != nil {
			v := string(*decision.Cause)
			cause = &v
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO transitions (
				link_id, from_state, to_state, outcome, cause, occurred_at
			)
			VALUES ($1, $2::link_state, $3::link_state, $4::check_outcome, $5::check_outcome, $6)
		`,
			linkID,
			currentState,
			string(decision.NextState),
			string(result.Outcome),
			cause,
			checkedAt,
		)
		if err != nil {
			return decision, fmt.Errorf("insert transition: %w", err)
		}
	}

	return decision, nil
}
