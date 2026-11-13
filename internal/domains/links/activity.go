package links

import (
	"time"

	"github.com/google/uuid"
)

type Check struct {
	ID         int64
	LinkID     uuid.UUID
	CheckedAt  time.Time
	Outcome    string
	StatusCode *int32
	FinalURL   *string
	Method     string
	ElapsedMS  int32
	Error      *string
}

type Transition struct {
	ID         int64
	LinkID     uuid.UUID
	FromState  string
	ToState    string
	Outcome    *string
	Cause      *string
	OccurredAt time.Time
	AcknowledgedAt time.Time
}

type View struct {
	Link           Link
	CollectionName string
}

type Detail struct {
	Link           Link
	CollectionName string
	Checks         []Check
	Transitions    []Transition
}
