package servergrpc

import (
	"context"
	"fmt"

	serverv1 "github.com/deeploop-ai/fleet/genproto/server/v1"
	sharedv1 "github.com/deeploop-ai/fleet/genproto/shared/v1"
	appserver "github.com/deeploop-ai/fleet/internal/app/server"
	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/pkg/contexts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TeamsService struct {
	serverv1.UnimplementedTeamsServiceServer
	teams *appserver.Teams
}

func NewTeamsService(teams *appserver.Teams) *TeamsService {
	return &TeamsService{teams: teams}
}

func (s *TeamsService) projectID(ctx context.Context) string {
	p, ok := contexts.Principal(ctx)
	if !ok {
		return ""
	}
	return p.ProjectID
}

func (s *TeamsService) CreateTeam(ctx context.Context, req *serverv1.CreateTeamRequest) (*serverv1.Team, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	doc, err := s.teams.CreateTeam(ctx, projectID, req.GetName(), req.GetPermissions())
	if err != nil {
		return nil, err
	}
	return mapTeamDoc(doc), nil
}

func (s *TeamsService) ListTeams(ctx context.Context, req *sharedv1.ListRequest) (*serverv1.ListTeamsResponse, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	docs, total, _, err := s.teams.ListTeams(ctx, projectID, databases.Query{
		Queries:   req.GetQueries(),
		PageSize:  req.GetPageSize(),
		PageToken: req.GetPageToken(),
	}, principalRoles(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]*serverv1.Team, len(docs))
	for i := range docs {
		out[i] = mapTeamDoc(&docs[i])
	}
	return &serverv1.ListTeamsResponse{
		Teams: out,
		Meta:  &sharedv1.ListResponseMeta{PageSize: req.GetPageSize(), TotalCount: int32(total)},
	}, nil
}

func (s *TeamsService) GetTeam(ctx context.Context, req *serverv1.GetTeamRequest) (*serverv1.Team, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	doc, err := s.teams.GetTeam(ctx, projectID, req.GetId())
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, status.Error(codes.NotFound, "team not found")
	}
	return mapTeamDoc(doc), nil
}

func (s *TeamsService) DeleteTeam(ctx context.Context, req *serverv1.GetTeamRequest) (*sharedv1.Empty, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	if err := s.teams.DeleteTeam(ctx, projectID, req.GetId()); err != nil {
		return nil, err
	}
	return &sharedv1.Empty{}, nil
}

func mapTeamDoc(doc *databases.Document) *serverv1.Team {
	if doc == nil {
		return nil
	}
	t := &serverv1.Team{
		Id:        doc.ID,
		CreatedAt: timestamppb.New(doc.CreatedAt),
		UpdatedAt: timestamppb.New(doc.UpdatedAt),
	}
	if v, ok := doc.Data["name"].(string); ok {
		t.Name = v
	}
	if v, ok := doc.Data["total"].(float64); ok {
		t.Total = int32(v)
	}
	if v, ok := doc.Data["total"].(int64); ok {
		t.Total = int32(v)
	}
	if arr, ok := doc.Data["permissions"].([]any); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				t.Permissions = append(t.Permissions, s)
			}
		}
	}
	return t
}

// splitPermission is duplicated here to avoid extra package; consider moving to shared util.
func splitPermission(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}

func unusedTeamFmt() string { return fmt.Sprint(1) }
