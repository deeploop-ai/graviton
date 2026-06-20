package client

import (
	"context"
	"fmt"

	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/domain/projects"
	"github.com/deeploop-ai/fleet/internal/pkg/contexts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Databases struct {
	projectRepo projects.Repository
	docDB       databases.DocumentDB
}

func NewDatabases(projectRepo projects.Repository, docDB databases.DocumentDB) *Databases {
	return &Databases{projectRepo: projectRepo, docDB: docDB}
}

func (d *Databases) resolveProject(ctx context.Context) (*projects.Project, []string, error) {
	p, ok := contexts.Principal(ctx)
	if !ok || p.ProjectID == "" || p.UserID == "" {
		return nil, nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	project, err := d.projectRepo.GetProject(ctx, p.ProjectID)
	if err != nil {
		return nil, nil, err
	}
	if project == nil {
		return nil, nil, status.Error(codes.NotFound, "project not found")
	}
	if err := d.docDB.EnsureSystemCollections(ctx, project.ID, project.InternalID); err != nil {
		return nil, nil, fmt.Errorf("ensure system collections: %w", err)
	}
	return project, p.Roles, nil
}

func (d *Databases) ensureCollection(ctx context.Context, databaseID, collectionID string) (string, []string, error) {
	project, roles, err := d.resolveProject(ctx)
	if err != nil {
		return "", nil, err
	}
	col, err := d.docDB.GetCollection(ctx, project.ID, databaseID, collectionID)
	if err != nil {
		return "", nil, err
	}
	if col == nil {
		return "", nil, status.Error(codes.NotFound, "collection not found")
	}
	return project.ID, roles, nil
}

func (d *Databases) CreateDocument(
	ctx context.Context,
	databaseID, collectionID, documentID string,
	data map[string]any,
	perms []databases.Permission,
) (*databases.Document, error) {
	projectID, roles, err := d.ensureCollection(ctx, databaseID, collectionID)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}
	p, _ := contexts.Principal(ctx)
	if len(perms) == 0 {
		perms = ownerDocumentPermissions(p.UserID)
	}
	created, err := d.docDB.CreateDocument(ctx, projectID, databaseID, collectionID, databases.Document{
		ID:   documentID,
		Data: data,
	}, perms)
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}
	got, err := d.docDB.GetDocument(ctx, projectID, databaseID, collectionID, created.ID, roles)
	if err != nil {
		return nil, err
	}
	if got == nil {
		return nil, status.Error(codes.NotFound, "document not found after create")
	}
	return got, nil
}

func (d *Databases) ListDocuments(
	ctx context.Context,
	databaseID, collectionID string,
	q databases.Query,
) ([]databases.Document, int64, string, error) {
	projectID, roles, err := d.ensureCollection(ctx, databaseID, collectionID)
	if err != nil {
		return nil, 0, "", err
	}
	list, err := d.docDB.ListDocuments(ctx, projectID, databaseID, collectionID, q, roles)
	if err != nil {
		return nil, 0, "", err
	}
	return list.Documents, list.TotalCount, list.NextPageToken, nil
}

func (d *Databases) GetDocument(ctx context.Context, databaseID, collectionID, documentID string) (*databases.Document, error) {
	projectID, roles, err := d.ensureCollection(ctx, databaseID, collectionID)
	if err != nil {
		return nil, err
	}
	doc, err := d.docDB.GetDocument(ctx, projectID, databaseID, collectionID, documentID, roles)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, status.Error(codes.NotFound, "document not found")
	}
	return doc, nil
}

func (d *Databases) UpdateDocument(
	ctx context.Context,
	databaseID, collectionID, documentID string,
	data map[string]any,
) (*databases.Document, error) {
	projectID, roles, err := d.ensureCollection(ctx, databaseID, collectionID)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}
	updated, err := d.docDB.UpdateDocument(ctx, projectID, databaseID, collectionID, databases.Document{
		ID:   documentID,
		Data: data,
	}, nil, roles)
	if err != nil {
		return nil, fmt.Errorf("update document: %w", err)
	}
	return &updated, nil
}

func (d *Databases) DeleteDocument(ctx context.Context, databaseID, collectionID, documentID string) error {
	projectID, roles, err := d.ensureCollection(ctx, databaseID, collectionID)
	if err != nil {
		return err
	}
	return d.docDB.DeleteDocument(ctx, projectID, databaseID, collectionID, documentID, roles)
}

func (d *Databases) CountDocuments(ctx context.Context, databaseID, collectionID string, queries []string) (int64, error) {
	projectID, roles, err := d.ensureCollection(ctx, databaseID, collectionID)
	if err != nil {
		return 0, err
	}
	return d.docDB.CountDocuments(ctx, projectID, databaseID, collectionID, queries, roles)
}

func ownerDocumentPermissions(userID string) []databases.Permission {
	userRole := fmt.Sprintf("user:%s", userID)
	return []databases.Permission{
		{Type: "read", Role: userRole},
		{Type: "update", Role: userRole},
		{Type: "delete", Role: userRole},
	}
}
