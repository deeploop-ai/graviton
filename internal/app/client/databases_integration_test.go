package client

import (
	"context"
	"testing"

	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/domain/shared"
	appserver "github.com/deeploop-ai/fleet/internal/app/server"
	"github.com/deeploop-ai/fleet/internal/infra/bun/bunrepo"
	"github.com/deeploop-ai/fleet/internal/infra/documentdb"
	"github.com/deeploop-ai/fleet/internal/pkg/config"
	"github.com/deeploop-ai/fleet/internal/pkg/contexts"
	"github.com/deeploop-ai/fleet/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestClientDatabases_DocumentCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	projectID, internalID, cleanup := testutil.CreateTestProject(ctx, db)
	defer cleanup()

	docDB := documentdb.NewPostgresDocumentDB(db)
	require.NoError(t, docDB.EnsureSystemCollections(ctx, projectID, internalID))

	projectRepo := bunrepo.NewProjectRepository(db)
	account := NewAccount(testConfig(), projectRepo, docDB)
	user, _, _, err := account.SignUp(ctx, SignUpCommand{
		ProjectID: projectID,
		Email:     "client-docs@fleet.local",
		Password:  "User@123456",
		Name:      "Client Docs",
	})
	require.NoError(t, err)

	serverUC := appserver.NewDatabases(projectRepo, docDB)
	require.NoError(t, serverUC.CreateDatabase(ctx, projectID, "app", "Application DB"))
	require.NoError(t, serverUC.CreateCollection(ctx, projectID, "app", "notes", "Notes", []databases.Attribute{
		{ID: "title", Key: "title", Type: "string", Size: 256},
	}, nil, nil, true))

	userCtx := contexts.WithPrincipal(ctx, &shared.Principal{
		ProjectID: projectID,
		UserID:    user.ID,
		Roles:     []string{"users", "user:" + user.ID},
	})
	clientUC := NewDatabases(projectRepo, docDB)

	created, err := clientUC.CreateDocument(userCtx, "app", "notes", "", map[string]any{
		"title": "Client note",
	}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	got, err := clientUC.GetDocument(userCtx, "app", "notes", created.ID)
	require.NoError(t, err)
	require.Equal(t, "Client note", got.Data["title"])

	otherCtx := contexts.WithPrincipal(ctx, &shared.Principal{
		ProjectID: projectID,
		UserID:    "other-user",
		Roles:     []string{"users", "user:other-user"},
	})
	_, err = clientUC.GetDocument(otherCtx, "app", "notes", created.ID)
	require.Error(t, err)

	updated, err := clientUC.UpdateDocument(userCtx, "app", "notes", created.ID, map[string]any{
		"title": "Updated note",
	}, nil, nil)
	require.NoError(t, err)
	require.Equal(t, "Updated note", updated.Data["title"])

	list, total, _, err := clientUC.ListDocuments(userCtx, "app", "notes", databases.Query{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)

	require.NoError(t, clientUC.DeleteDocument(userCtx, "app", "notes", created.ID))
}

func testConfig() *config.AppConfig {
	return &config.AppConfig{}
}
