package links

import (
	"time"

	"github.com/google/uuid"
)

type CreateLinkRequest struct {
	URL          string     `form:"url" binding:"required,url"`
	CollectionID *uuid.UUID `form:"collection_id"`
	Title        string     `form:"title"`
}

type DueLink struct {
	ID         uuid.UUID
	URL        string
	Host       string
	LeaseUntil time.Time
}
