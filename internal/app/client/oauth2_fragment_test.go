package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendOAuthSPAFragment(t *testing.T) {
	t.Parallel()
	url := appendOAuthSPAFragment("https://app.example/login/callback", "user-1", &TokenBundle{
		AccessToken:  "at-1",
		RefreshToken: "rt-1",
	})
	require.Contains(t, url, "#")
	require.Contains(t, url, "access_token=at-1")
	require.NotContains(t, url, "refresh_token=")
	require.Contains(t, url, "userId=user-1")
}
