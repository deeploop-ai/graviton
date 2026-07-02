package auth

import (
	"context"
	"time"
)

// OAuthState captures pending OAuth2 authorization context stored server-side.
type OAuthState struct {
	StateID      string
	ProjectID    string
	Provider     string
	SuccessURL   string
	FailureURL   string
	PKCEVerifier string
	// LinkUserID, when set, binds the OAuth identity to an existing authenticated user.
	LinkUserID string
}

// OAuthStateStore persists OAuth2 state and PKCE verifiers until callback.
type OAuthStateStore interface {
	Save(ctx context.Context, state OAuthState, ttl time.Duration) error
	Get(ctx context.Context, stateID string) (*OAuthState, error)
	Delete(ctx context.Context, stateID string) error
}

// OAuthUserInfo is normalized profile data from an OAuth2 provider.
type OAuthUserInfo struct {
	ProviderUID string
	UnionID     string
	OpenID      string
	Email       string
	Name        string
	AvatarURL   string
	Raw         map[string]any
}

// OAuthAuthenticator builds authorize URLs and exchanges authorization codes.
type OAuthAuthenticator interface {
	AuthorizeURL(stateID, pkceChallenge string) string
	Exchange(ctx context.Context, code, pkceVerifier string) (*OAuthUserInfo, error)
}
