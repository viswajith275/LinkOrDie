package services

import (
	"context"
	"link_or_die/internal/apperror"
	"link_or_die/internal/domains/links"
	"link_or_die/internal/store"

	"github.com/google/uuid"
)

type LinkService struct {
	r  *store.LinkRepository
	cr *store.CollectionRepository
}

func CreateLinkService(r *store.LinkRepository, cr *store.CollectionRepository) *LinkService {
	return &LinkService{r: r,
		cr: cr}
}

func (s *LinkService) AddLink(c context.Context, userID uuid.UUID, req links.CreateLinkRequest) (*links.Link, error) {

	normalised_url, host, hash, err := links.Normalize(req.URL)

	if err != nil {
		return nil, err
	}

	link := &links.Link{
		UserID:        userID,
		URL:           req.URL,
		NormalizedURL: normalised_url,
		Host:          host,
		URLHash:       hash,
	}

	if req.Title != "" {
		title := req.Title
		link.Title = &title
	}

	if req.CollectionID != nil {
		ok, err := s.cr.ExistsForUser(c, userID, *req.CollectionID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, apperror.NewValidation("collection_id", "collection not found")
		}
		link.CollectionID = req.CollectionID
	}

	return s.r.Create(c, link)

}
