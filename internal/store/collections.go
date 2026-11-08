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

func (r *CollectionRepository) Create(c context.Context, userID uuid.UUID, name string) (*collections.Collection, error) {
	var collection collections.Collection

	err := r.db.QueryRow(c, `
		INSERT INTO collections (user_id, name)
	 	VALUES ($1, $2)
		RETURNING id, user_id, name, created_at
		`, userID, name).Scan(&collection.ID, &collection.UserID, &collection.Name, &collection.CreatedAt)

	if err != nil {
		return nil, translatePgError(err, "creating collection")
	}

	return &collection, nil
}

func (r *CollectionRepository) Update(c context.Context, id uuid.UUID, userID uuid.UUID, name string) (*collections.Collection, error) {
	var collection collections.Collection

	err := r.db.QueryRow(c, `
		UPDATE collections SET name = $1
	 	WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, name, created_at
		`, name, id, userID).Scan(&collection.ID, &collection.UserID, &collection.Name, &collection.CreatedAt)

	if err != nil {
		return nil, translatePgError(err, "updating collection")
	}

	return &collection, nil
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

func (r *CollectionRepository) FindAllCollectionOfUser(c context.Context, userID uuid.UUID) ([]collections.Collection, error) {
	var out []collections.Collection

	rows, err := r.db.Query(c, `
		SELECT id, user_id, name, created_at
		FROM collections
		WHERE user_id = $1
		`, userID)

	if err != nil {
		return nil, translatePgError(err, "listing links")
	}
	defer rows.Close()

	for rows.Next() {
		var c collections.Collection
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
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
