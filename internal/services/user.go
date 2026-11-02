package services

import (
	"context"
	"link_or_die/internal/domains/user"
	"link_or_die/internal/store"

	"github.com/google/uuid"
)

type UserService struct {
	r *store.UserRepository
}

func CreateUserService(userRepo *store.UserRepository) *UserService {
	return &UserService{
		r: userRepo,
	}
}

func (s *UserService) GetUserById(c context.Context, userID uuid.UUID) (*user.User, error) {
	return s.r.GetUserByID(c, userID)
}
