package bunrepo_test

import (
	"context"
	"testing"

	domainauth "github.com/deeploop-ai/graviton/internal/domain/auth"
	"github.com/deeploop-ai/graviton/internal/domain/projects"
	"github.com/deeploop-ai/graviton/internal/infra/bun/bunrepo"
	"github.com/deeploop-ai/graviton/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestOAuthProviderRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	projectID, _, cleanup := testutil.CreateTestProject(ctx, db)
	defer cleanup()

	repo := bunrepo.NewOAuthProviderRepository(db)

	list, err := repo.ListOAuthProviders(ctx, projectID)
	require.NoError(t, err)
	require.Empty(t, list)

	cfg := &projects.OAuthProvider{
		ProjectID:    projectID,
		Provider:     domainauth.ProviderGoogle,
		Enabled:      true,
		ClientID:     "id-1",
		ClientSecret: "secret-1",
		Scopes:       []string{"openid", "email"},
	}
	require.NoError(t, repo.UpsertOAuthProvider(ctx, cfg))

	stored, err := repo.GetOAuthProvider(ctx, projectID, domainauth.ProviderGoogle)
	require.NoError(t, err)
	require.NotNil(t, stored)
	require.Equal(t, "id-1", stored.ClientID)

	cfg.ClientID = "id-2"
	cfg.ClientSecret = "secret-2"
	cfg.Scopes = []string{"openid"}
	require.NoError(t, repo.UpsertOAuthProvider(ctx, cfg))

	stored, err = repo.GetOAuthProvider(ctx, projectID, domainauth.ProviderGoogle)
	require.NoError(t, err)
	require.Equal(t, "id-2", stored.ClientID)

	list, err = repo.ListOAuthProviders(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, repo.DeleteOAuthProvider(ctx, projectID, domainauth.ProviderGoogle))

	stored, err = repo.GetOAuthProvider(ctx, projectID, domainauth.ProviderGoogle)
	require.NoError(t, err)
	require.Nil(t, stored)
}
