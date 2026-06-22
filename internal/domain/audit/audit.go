package audit

import (
	"context"
	"time"
)

type Entry struct {
	ID         string
	ProjectID  string
	ActorID    string
	ActorKind  string
	Action     string
	ResourceID string
	Status     string
	IP         string
	UserAgent  string
	Metadata   map[string]any
	CreatedAt  time.Time
}

type Repository interface {
	Insert(ctx context.Context, entry *Entry) error
}
