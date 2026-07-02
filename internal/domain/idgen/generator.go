package idgen

import "context"

// Resource identifies which entity class is receiving a new identifier.
type Resource string

const (
	ResourceUsers     Resource = "users"
	ResourceSessions  Resource = "sessions"
	ResourceDocuments Resource = "documents"
	ResourceDefault   Resource = "default"
)

// Generator creates unique string identifiers per project and resource type.
type Generator interface {
	NewID(ctx context.Context, projectID string, resource Resource) (string, error)
}
