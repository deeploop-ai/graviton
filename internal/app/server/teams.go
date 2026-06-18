package server

import (
	"context"
	"fmt"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/domain/projects"
	"github.com/deeploop-ai/fleet/pkg/idgen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Teams struct {
	projectRepo projects.Repository
	docDB       databases.DocumentDB
}

func NewTeams(projectRepo projects.Repository, docDB databases.DocumentDB) *Teams {
	return &Teams{projectRepo: projectRepo, docDB: docDB}
}

func (t *Teams) resolveProject(ctx context.Context, projectID string) (*projects.Project, error) {
	p, err := t.projectRepo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, status.Error(codes.NotFound, "project not found")
	}
	if err := t.docDB.EnsureSystemCollections(ctx, p.ID, p.InternalID); err != nil {
		return nil, err
	}
	return p, nil
}

func (t *Teams) CreateTeam(ctx context.Context, projectID, name string, perms []string) (*databases.Document, error) {
	if _, err := t.resolveProject(ctx, projectID); err != nil {
		return nil, err
	}
	teamID := idgen.UUID().String()
	doc := databases.Document{
		ID: teamID,
		Data: map[string]any{
			"name":        name,
			"permissions": perms,
			"total":       0,
		},
	}
	if _, err := t.docDB.CreateDocument(ctx, projectID, "default", "teams", doc, defaultTeamPermissions(teamID, perms)); err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}
	return t.docDB.GetDocument(ctx, projectID, "default", "teams", teamID)
}

func (t *Teams) ListTeams(ctx context.Context, projectID string, q databases.Query, roles []string) ([]databases.Document, int64, string, error) {
	if _, err := t.resolveProject(ctx, projectID); err != nil {
		return nil, 0, "", err
	}
	list, err := t.docDB.ListDocuments(ctx, projectID, "default", "teams", q, roles)
	if err != nil {
		return nil, 0, "", err
	}
	return list.Documents, list.TotalCount, list.NextPageToken, nil
}

func (t *Teams) GetTeam(ctx context.Context, projectID, teamID string) (*databases.Document, error) {
	if _, err := t.resolveProject(ctx, projectID); err != nil {
		return nil, err
	}
	return t.docDB.GetDocument(ctx, projectID, "default", "teams", teamID)
}

func (t *Teams) DeleteTeam(ctx context.Context, projectID, teamID string) error {
	if _, err := t.resolveProject(ctx, projectID); err != nil {
		return err
	}
	return t.docDB.DeleteDocument(ctx, projectID, "default", "teams", teamID)
}

func defaultTeamPermissions(teamID string, explicit []string) []databases.Permission {
	var perms []databases.Permission
	if len(explicit) > 0 {
		for _, r := range explicit {
			parts := splitPermission(r)
			if len(parts) == 2 {
				perms = append(perms, databases.Permission{Type: parts[0], Role: parts[1]})
			}
		}
		return perms
	}
	return []databases.Permission{
		{Type: "read", Role: fmt.Sprintf("team:%s", teamID)},
		{Type: "update", Role: fmt.Sprintf("team:%s", teamID)},
		{Type: "delete", Role: fmt.Sprintf("team:%s", teamID)},
		{Type: "read", Role: "admin"},
		{Type: "update", Role: "admin"},
		{Type: "delete", Role: "admin"},
	}
}

func splitPermission(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}

func teamUpdatedAt() time.Time { return time.Now() }
