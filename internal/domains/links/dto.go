package links

import "github.com/google/uuid"

type CreateLinkRequest struct {
	URL          string     `form:"url" binding:"required,url"`
	CollectionID *uuid.UUID `form:"collection_id"`
	Title        string     `form:"title"`
}
