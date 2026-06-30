package projects

import "context"

// OAuthProvider holds per-project OAuth2 client credentials.
type OAuthProvider struct {
	ProjectID    string
	Provider     string
	Enabled      bool
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// OAuthProviderRepository loads OAuth provider configuration for a project.
type OAuthProviderRepository interface {
	GetOAuthProvider(ctx context.Context, projectID, provider string) (*OAuthProvider, error)
}
