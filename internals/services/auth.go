package services

import (
	"context"
	"errors"
	"link_or_die/internals/apperror"
	"link_or_die/internals/domains/auth"
	"link_or_die/internals/domains/user"
	"link_or_die/internals/store"
)

type AuthService struct {
	r *store.UserRepository
}

func CreateAuthService(userRepo *store.UserRepository) *AuthService {
	return &AuthService{
		r: userRepo,
	}
}

func (s *AuthService) LoginWithOAuth(ctx context.Context, identity *auth.Identity) (*user.User, error) {
	u, err := s.r.FindUserByOAuth(ctx, string(identity.Provider), identity.ProviderUID)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	// First login with this provider → create user + link atomically (single transaction)
	return s.r.LinkOAuthAndCreateUser(
		ctx,
		string(identity.Provider), identity.ProviderUID,
		identity.Email, identity.Name,
	)
}
