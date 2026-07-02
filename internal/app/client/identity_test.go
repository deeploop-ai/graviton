package client

import (
	"context"
	"testing"

	domainauth "github.com/deeploop-ai/graviton/internal/domain/auth"
	"github.com/deeploop-ai/graviton/internal/infra/bun/bunrepo"
	"github.com/deeploop-ai/graviton/internal/infra/documentdb"
	"github.com/deeploop-ai/graviton/internal/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestResolveOAuthUser_RejectsExistingEmailWithoutIdentity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	projectID, _, cleanup := testutil.CreateTestProject(ctx, db)
	defer cleanup()

	cfg := buildTestConfig()
	projectRepo := bunrepo.NewProjectRepository(db)
	docDB := documentdb.NewPostgresDocumentDB(db)
	account := NewTestAccount(cfg, projectRepo, docDB)

	_, _, _, err := account.SignUp(ctx, SignUpCommand{
		ProjectID: projectID,
		Email:     "existing@graviton.local",
		Password:  "User@123",
		Name:      "Existing",
	})
	require.NoError(t, err)

	_, err = account.resolveOAuthUser(ctx, projectID, domainauth.ProviderGoogle, &domainauth.OAuthUserInfo{
		ProviderUID: "google-123",
		Email:       "existing@graviton.local",
		Name:        "OAuth User",
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}
