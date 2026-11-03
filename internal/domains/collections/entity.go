package collections

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	CreatedAt time.Time
}
