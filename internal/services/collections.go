package services

import (
	"context"
	"link_or_die/internal/domains/collections"
	"link_or_die/internal/store"

	"github.com/google/uuid"
)

type CollectionService struct {
	r  *store.CollectionRepository
}

func CreateCollectionService(r *store.CollectionRepository) *CollectionService {
	return &CollectionService{
		r: r,
		}
}

func (s *CollectionService) CreateCollection(c context.Context, userID uuid.UUID, req collections.CreateCollectionRequest) (*collections.Collection, error) {
	return s.r.Create(c, userID, req.Name)
}
