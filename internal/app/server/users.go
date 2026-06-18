package server

import (
	"context"
	"fmt"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/domain/projects"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Users struct {
	projectRepo projects.Repository
	docDB       databases.DocumentDB
}

func NewUsers(projectRepo projects.Repository, docDB databases.DocumentDB) *Users {
	return &Users{projectRepo: projectRepo, docDB: docDB}
}

func (u *Users) resolveProject(ctx context.Context, projectID string) (*projects.Project, error) {
	p, err := u.projectRepo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, status.Error(codes.NotFound, "project not found")
	}
	if err := u.docDB.EnsureSystemCollections(ctx, p.ID, p.InternalID); err != nil {
		return nil, err
	}
	return p, nil
}

func (u *Users) ListUsers(ctx context.Context, projectID string, q databases.Query, roles []string) ([]databases.Document, int64, string, error) {
	if _, err := u.resolveProject(ctx, projectID); err != nil {
		return nil, 0, "", err
	}
	list, err := u.docDB.ListDocuments(ctx, projectID, "default", "users", q, roles)
	if err != nil {
		return nil, 0, "", err
	}
	return list.Documents, list.TotalCount, list.NextPageToken, nil
}

func (u *Users) GetUser(ctx context.Context, projectID, userID string, roles []string) (*databases.Document, error) {
	if _, err := u.resolveProject(ctx, projectID); err != nil {
		return nil, err
	}
	return u.docDB.GetDocument(ctx, projectID, "default", "users", userID, roles)
}

func (u *Users) UpdateUser(ctx context.Context, projectID, userID string, updates map[string]any, roles []string) (*databases.Document, error) {
	if _, err := u.resolveProject(ctx, projectID); err != nil {
		return nil, err
	}
	doc := databases.Document{ID: userID, Data: updates}
	updated, err := u.docDB.UpdateDocument(ctx, projectID, "default", "users", doc, nil, roles)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &updated, nil
}

func (u *Users) DeleteUser(ctx context.Context, projectID, userID string, roles []string) error {
	if _, err := u.resolveProject(ctx, projectID); err != nil {
		return err
	}
	return u.docDB.DeleteDocument(ctx, projectID, "default", "users", userID, roles)
}

func (u *Users) UpdateUserStatus(ctx context.Context, projectID, userID, status string, roles []string) (*databases.Document, error) {
	if status == "" {
		status = "active"
	}
	return u.UpdateUser(ctx, projectID, userID, map[string]any{"status": status, "updated_at": time.Now().Format(time.RFC3339Nano)}, roles)
}
