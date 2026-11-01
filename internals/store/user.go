package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"link_or_die/internals/domains/user"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func CreateUserRepo(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.QueryRow(ctx, `
		SELECT id, username, email
		FROM users WHERE id = $1
		`, userID).Scan(&u.ID, &u.Username, &u.Email)
	if err != nil {
		return nil, translatePgError(err, "finding user by userid")
	}
	return &u, nil
}

func (r *UserRepository) FindUserByOAuth(ctx context.Context, provider, providerUID string) (*user.User, error) {
	var u user.User
	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.username, u.email
		FROM user_oauth o
		JOIN users u ON u.id = o.user_id
		WHERE o.provider = $1 AND o.provider_uid = $2`,
		provider, providerUID,
	).Scan(&u.ID, &u.Username, &u.Email)

	if err != nil {
		return nil, translatePgError(err, "finding user by oauth")
	}
	return &u, nil
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db.QueryRow(ctx, `
		SELECT id, username, email
		FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Username, &u.Email)

	if err != nil {
		return nil, translatePgError(err, "finding user by email")
	}
	return &u, nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (r *UserRepository) LinkOAuthAndCreateUser(
	ctx context.Context,
	provider, providerUID, email, name string,
) (*user.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Check if OAuth identity already exists
	var existingID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT user_id FROM user_oauth
		WHERE provider = $1 AND provider_uid = $2`,
		provider, providerUID,
	).Scan(&existingID)

	if err == nil {
		_ = tx.Commit(ctx)
		return r.GetUserByID(ctx, existingID)
	}

	// Check if a user with this email already exists (Account Linking)
	var existingEmailUserID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM users
		WHERE email = $1`,
		email,
	).Scan(&existingEmailUserID)

	if err == nil {
		// Link OAuth identity to this existing user
		_, err = tx.Exec(ctx, `
			INSERT INTO user_oauth (user_id, provider, provider_uid, email)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (provider, provider_uid) DO NOTHING`,
			existingEmailUserID, provider, providerUID, email,
		)
		if err != nil {
			return nil, translatePgError(err, "linking oauth to existing user by email")
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return r.GetUserByID(ctx, existingEmailUserID)
	}

	// User does not exist, create new user
	baseUsername := strings.TrimSpace(name)
	if baseUsername == "" {
		if parts := strings.Split(email, "@"); len(parts) > 0 && parts[0] != "" {
			baseUsername = parts[0]
		} else {
			baseUsername = "user"
		}
	}

	var u user.User
	var lastErr error
	for attempts := 0; attempts < 5; attempts++ {
		candidate := baseUsername
		if attempts > 0 {
			candidate = fmt.Sprintf("%s_%s", baseUsername, randomHex(3))
		}

		err = tx.QueryRow(ctx, `
			INSERT INTO users (username, email, role)
			VALUES ($1, $2, 'user')
			RETURNING id, username, email`,
			candidate, email,
		).Scan(&u.ID, &u.Username, &u.Email)

		if err == nil {
			lastErr = nil
			break
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// If unique violation on username, retry with random suffix
			lastErr = err
			continue
		}

		return nil, translatePgError(err, "creating oauth user")
	}

	if lastErr != nil {
		return nil, translatePgError(lastErr, "creating oauth user after username retries")
	}

	// Link OAuth identity
	_, err = tx.Exec(ctx, `
		INSERT INTO user_oauth (user_id, provider, provider_uid, email)
		VALUES ($1, $2, $3, $4)`,
		u.ID, provider, providerUID, email,
	)
	if err != nil {
		return nil, translatePgError(err, "linking oauth identity")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &u, nil
}
