package client

import (
	"context"

	"github.com/deeploop-ai/fleet/internal/app/server"
	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/domain/teams"
	"github.com/deeploop-ai/fleet/internal/pkg/contexts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Teams struct {
	teams *server.Teams
}

func NewTeams(teams *server.Teams) *Teams {
	return &Teams{teams: teams}
}

func (t *Teams) principal(ctx context.Context) (projectID, userID, email string, roles []string, err error) {
	p, ok := contexts.Principal(ctx)
	if !ok || p.ProjectID == "" || p.UserID == "" {
		return "", "", "", nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return p.ProjectID, p.UserID, p.Email, p.Roles, nil
}

func (t *Teams) CreateTeam(ctx context.Context, name string) (*databases.Document, error) {
	projectID, userID, email, roles, err := t.principal(ctx)
	if err != nil {
		return nil, err
	}
	team, _, err := t.teams.CreateTeamWithOwner(ctx, projectID, name, userID, email, roles)
	return team, err
}

func (t *Teams) ListTeams(ctx context.Context, q databases.Query) ([]databases.Document, int64, string, error) {
	projectID, _, _, roles, err := t.principal(ctx)
	if err != nil {
		return nil, 0, "", err
	}
	return t.teams.ListTeams(ctx, projectID, q, roles)
}

func (t *Teams) GetTeam(ctx context.Context, teamID string) (*databases.Document, error) {
	projectID, _, _, roles, err := t.principal(ctx)
	if err != nil {
		return nil, err
	}
	return t.teams.GetTeam(ctx, projectID, teamID, roles)
}

func (t *Teams) DeleteTeam(ctx context.Context, teamID string) error {
	projectID, _, _, roles, err := t.principal(ctx)
	if err != nil {
		return err
	}
	return t.teams.DeleteTeam(ctx, projectID, teamID, roles)
}

func (t *Teams) CreateMembership(ctx context.Context, teamID, inviteEmail, name string, roles []string) (*databases.Document, error) {
	projectID, _, _, principalRoles, err := t.principal(ctx)
	if err != nil {
		return nil, err
	}
	return t.teams.CreateMembership(ctx, projectID, server.CreateMembershipCommand{
		TeamID: teamID,
		Email:  inviteEmail,
		Name:   name,
		Roles:  roles,
		Status: teams.StatusPending,
	}, principalRoles)
}

func (t *Teams) ListMemberships(ctx context.Context, teamID string) ([]databases.Document, error) {
	projectID, _, _, roles, err := t.principal(ctx)
	if err != nil {
		return nil, err
	}
	docs, _, _, err := t.teams.ListMemberships(ctx, projectID, teamID, databases.Query{}, roles)
	return docs, err
}

func (t *Teams) UpdateMembershipStatus(ctx context.Context, teamID, membershipID, statusVal string) (*databases.Document, error) {
	projectID, userID, email, roles, err := t.principal(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := t.teams.GetMembership(ctx, projectID, teamID, membershipID, roles)
	if err != nil {
		return nil, err
	}
	memUserID, _ := doc.Data["user_id"].(string)
	memEmail, _ := doc.Data["email"].(string)
	if memUserID != userID && memEmail != email {
		return nil, status.Error(codes.PermissionDenied, "cannot update another user's membership")
	}
	return t.teams.UpdateMembershipStatus(ctx, projectID, teamID, membershipID, statusVal, roles)
}

func (t *Teams) DeleteMembership(ctx context.Context, teamID, membershipID string) error {
	projectID, _, _, roles, err := t.principal(ctx)
	if err != nil {
		return err
	}
	return t.teams.DeleteMembership(ctx, projectID, teamID, membershipID, roles)
}
