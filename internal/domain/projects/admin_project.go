package projects

import "context"

type ConsoleAdminProjectRepository interface {
	HasProjectAccess(ctx context.Context, adminID, projectID string) (bool, error)
	GrantProjectAccess(ctx context.Context, adminID, projectID string) error
}
