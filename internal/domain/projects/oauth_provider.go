package projects

import (
	"context"
	"time"
)

// OAuthProvider holds per-project OAuth2 client credentials.
type OAuthProvider struct {
	ProjectID    string
	Provider     string
	Enabled      bool
	ClientID     string
	ClientSecret string
	Scopes       []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// OAuthProviderRepository manages per-project OAuth provider configuration.
type OAuthProviderRepository interface {
	GetOAuthProvider(ctx context.Context, projectID, provider string) (*OAuthProvider, error)
	ListOAuthProviders(ctx context.Context, projectID string) ([]OAuthProvider, error)
	UpsertOAuthProvider(ctx context.Context, cfg *OAuthProvider) error
	DeleteOAuthProvider(ctx context.Context, projectID, provider string) error
}
