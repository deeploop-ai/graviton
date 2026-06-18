package documentdb

import (
	"context"
	"testing"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestPostgresDocumentDatabase_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	projectID, _, cleanup := testutil.CreateTestProject(ctx, db)
	defer cleanup()

	docDB := NewPostgresDocumentDatabase(db)
	require.NoError(t, docDB.EnsureSystemCollections(ctx, projectID, 0))

	// Create a custom database and collection.
	require.NoError(t, docDB.CreateDatabase(ctx, projectID, "app", "Application DB"))
	require.NoError(t, docDB.CreateCollection(ctx, projectID, "app", "posts", "Posts", []databases.Attribute{
		{ID: "title", Key: "title", Type: "string", Size: 256},
		{ID: "views", Key: "views", Type: "integer"},
	}, []databases.Index{
		{ID: "title_key", Type: "key", Attributes: []string{"title"}},
	}))

	// Create document.
	created, err := docDB.CreateDocument(ctx, projectID, "app", "posts", databases.Document{
		Data: map[string]any{
			"title": "Hello World",
			"views": 42,
		},
	}, []databases.Permission{
		{Type: "read", Role: "any"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	// Get document.
	got, err := docDB.GetDocument(ctx, projectID, "app", "posts", created.ID)
	require.NoError(t, err)
	require.Equal(t, "Hello World", got.Data["title"])

	// Update document.
	updated, err := docDB.UpdateDocument(ctx, projectID, "app", "posts", databases.Document{
		ID: got.ID,
		Data: map[string]any{
			"views": 100,
		},
	}, nil)
	require.NoError(t, err)
	require.Equal(t, float64(100), updated.Data["views"])

	// List with Appwrite-style query.
	list, err := docDB.ListDocuments(ctx, projectID, "app", "posts", databases.Query{
		Queries: []string{`greaterThan("views",50)`, `orderDesc("$createdAt")`},
	}, []string{"any"})
	require.NoError(t, err)
	require.Len(t, list.Documents, 1)
	require.Equal(t, int64(1), list.TotalCount)

	// Count.
	count, err := docDB.CountDocuments(ctx, projectID, "app", "posts", []string{`equal("title","Hello World")`}, []string{"any"})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Delete.
	require.NoError(t, docDB.DeleteDocument(ctx, projectID, "app", "posts", created.ID))
	got2, err := docDB.GetDocument(ctx, projectID, "app", "posts", created.ID)
	require.NoError(t, err)
	require.Nil(t, got2)
}

func TestPostgresDocumentDatabase_Permissions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	projectID, _, cleanup := testutil.CreateTestProject(ctx, db)
	defer cleanup()

	docDB := NewPostgresDocumentDatabase(db)
	require.NoError(t, docDB.EnsureSystemCollections(ctx, projectID, 0))

	created, err := docDB.CreateDocument(ctx, projectID, "default", "users", databases.Document{
		Data: map[string]any{
			"email": "perm@fleet.local",
			"name":  "Permission Test",
		},
	}, []databases.Permission{
		{Type: "read", Role: "user:alice"},
	})
	require.NoError(t, err)

	// User without permission cannot read.
	list, err := docDB.ListDocuments(ctx, projectID, "default", "users", databases.Query{
		Queries: []string{`equal("$id","` + created.ID + `")`},
	}, []string{"user:bob"})
	require.NoError(t, err)
	require.Len(t, list.Documents, 0)

	// User with permission can read.
	list, err = docDB.ListDocuments(ctx, projectID, "default", "users", databases.Query{
		Queries: []string{`equal("$id","` + created.ID + `")`},
	}, []string{"user:alice"})
	require.NoError(t, err)
	require.Len(t, list.Documents, 1)

	// Admin bypasses permissions.
	list, err = docDB.ListDocuments(ctx, projectID, "default", "users", databases.Query{}, []string{"admin"})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(list.Documents), 1)

	_ = time.Now()
}
