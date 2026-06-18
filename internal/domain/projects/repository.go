package projects

import "context"

type Repository interface {
	CreateProject(ctx context.Context, p *Project) error
	GetProject(ctx context.Context, id string) (*Project, error)
	GetProjectByName(ctx context.Context, name string) (*Project, error)
	ListProjects(ctx context.Context) ([]Project, error)
	UpdateProject(ctx context.Context, p *Project) error
	DeleteProject(ctx context.Context, id string) error
}

type APIKeyRepository interface {
	CreateAPIKey(ctx context.Context, key *APIKey) error
	GetAPIKey(ctx context.Context, id string) (*APIKey, error)
	GetAPIKeyBySecretHash(ctx context.Context, hash string) (*APIKey, error)
	ListAPIKeys(ctx context.Context, projectID string) ([]APIKey, error)
	DeleteAPIKey(ctx context.Context, id string) error
}

type ConsoleAdminRepository interface {
	GetConsoleAdmin(ctx context.Context, id string) (*ConsoleAdmin, error)
	GetConsoleAdminByEmail(ctx context.Context, email string) (*ConsoleAdmin, error)
}

type ProjectResolver interface {
	InternalID(ctx context.Context, projectID string) (int64, error)
}
