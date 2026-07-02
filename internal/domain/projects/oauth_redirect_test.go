package projects

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchRedirectURL(t *testing.T) {
	t.Parallel()
	allowed := []string{
		"https://app.example.com/callback",
		"http://localhost:5173",
	}
	require.True(t, MatchRedirectURL("https://app.example.com/callback?x=1", allowed))
	require.True(t, MatchRedirectURL("http://localhost:5173/oauth", allowed))
	require.False(t, MatchRedirectURL("https://evil.example.com/callback", allowed))
}

func TestOAuthAllowedRedirectURLs(t *testing.T) {
	t.Parallel()
	settings := map[string]any{
		SettingsKeyOAuthAllowedRedirectURLs: []any{"https://app.example.com"},
	}
	require.Equal(t, []string{"https://app.example.com"}, OAuthAllowedRedirectURLs(settings))
}
