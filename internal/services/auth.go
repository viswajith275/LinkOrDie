package services

import (
	"context"
	"errors"
	"fmt"
	"link_or_die/internal/apperror"
	"link_or_die/internal/domains/auth"
	"link_or_die/internal/domains/user"
	"link_or_die/internal/store"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	r *store.UserRepository
}

func CreateAuthService(userRepo *store.UserRepository) *AuthService {
	return &AuthService{
		r: userRepo,
	}
}

// just to make the waiting time equal for not found and wrong pass (by just becrypting a dummy hash)
var dummyHash []byte

func init() {
	var err error
	dummyHash, err = bcrypt.GenerateFromPassword([]byte("dummy-random-string"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
}

func (s *AuthService) Register(ctx context.Context, registerReq user.RegisterRequest) (*user.User, error) {
	username := strings.TrimSpace(registerReq.Username)
	email := strings.ToLower(strings.TrimSpace(registerReq.Email))

	if len(registerReq.Password) > 72 {
		return nil, apperror.NewValidation("password", "Password must be 72 bytes or fewer!!")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(registerReq.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	u, err := s.r.CreateUser(ctx, username, email, string(hash))
	if errors.Is(err, apperror.ErrConflict) {

		return nil, apperror.NewValidation("username", "Username or email already taken")
	}
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return u, nil
}

func (s *AuthService) LoginWithPassword(ctx context.Context, username, password string) (*user.User, error) {
	u, hash, err := s.r.GetUserByUserName(ctx, username)

	if errors.Is(err, apperror.ErrNotFound) {
		// equalising time (hashing dummy)
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, apperror.ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	return u, nil
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
