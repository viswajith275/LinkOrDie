package store

import (
	"errors"
	"fmt"
	"link_or_die/internals/apperror"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func translatePgError(err error, context string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", context, apperror.ErrNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%s: %w", context, apperror.ErrConflict)
	}

	return fmt.Errorf("%s: %v", context, err)
}
