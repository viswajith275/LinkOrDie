package store

import (
	"context"
	"link_or_die/internal/domains/collections"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CollectionRepository struct {
	db *pgxpool.Pool
}

func CreateCollectionRepo(db *pgxpool.Pool) *CollectionRepository {
	return &CollectionRepository{
		db: db,
	}
}

func (r *CollectionRepository) FindCollectionByID(c context.Context, userID uuid.UUID, id uuid.UUID) (*collections.Collection, error) {
	var collection collections.Collection

	err := r.db.QueryRow(c, `
		SELECT id, user_id, name, created_at
		FROM collections
		WHERE id = $1 AND user_id = $2
		`, id, userID).Scan(&collection.ID, &collection.UserID, &collection.Name, &collection.CreatedAt)

	if err != nil {
		return nil, translatePgError(err, "finding collection by id")
	}

	return &collection, nil
}

func (r *CollectionRepository) ExistsForUser(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM collections WHERE id = $1 AND user_id = $2)`,
		id, userID,
	).Scan(&exists)
	if err != nil {
		return false, translatePgError(err, "checking collection ownership")
	}
	return exists, nil
}
