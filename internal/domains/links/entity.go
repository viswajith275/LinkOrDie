package links

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CollectionID *uuid.UUID

	URL           string
	NormalizedURL string
	URLHash       []byte
	Host          string
	Title         *string

	State          string
	LastOutcome    *string
	LastStatusCode *int32
	LastError      *string
	CauseOfDeath   *string

	ConsecutiveFailures int
	TotalFailures       int
	ResurrectedCount    int

	LastCheckedAt *time.Time
	LastOKAt      *time.Time
	DiedAt        *time.Time
	NextCheckAt   *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (l Link) TitleOrHost() string {
	if l.Title != nil && *l.Title != "" {
		return *l.Title
	}

	return l.Host
}
