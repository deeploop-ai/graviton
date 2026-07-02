package idgen_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/deeploop-ai/graviton/internal/domain/idgen"
	"github.com/deeploop-ai/graviton/internal/domain/projects"
	infraidgen "github.com/deeploop-ai/graviton/internal/infra/idgen"
	"github.com/deeploop-ai/graviton/internal/pkg/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type stubProjectRepo struct {
	settings map[string]any
}

func (s stubProjectRepo) GetProject(context.Context, string) (*projects.Project, error) {
	return &projects.Project{ID: "proj-1", Settings: s.settings}, nil
}
func (stubProjectRepo) CreateProject(context.Context, *projects.Project) error { return nil }
func (stubProjectRepo) GetProjectByName(context.Context, string) (*projects.Project, error) {
	return nil, nil
}
func (stubProjectRepo) ListProjects(context.Context) ([]projects.Project, error) { return nil, nil }
func (stubProjectRepo) UpdateProject(context.Context, *projects.Project) error     { return nil }
func (stubProjectRepo) DeleteProject(context.Context, string) error                { return nil }

func TestService_NewID_Strategies(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg := &config.AppConfig{
		Idgen: &config.IdGen{
			DefaultStrategy: "uuid",
			Resources: &config.IdGen_Resources{
				Users: "sequence",
			},
		},
	}
	svc, err := infraidgen.NewService(cfg, rdb, stubProjectRepo{})
	require.NoError(t, err)

	id1, err := svc.NewID(ctx, "proj-1", idgen.ResourceUsers)
	require.NoError(t, err)
	require.Equal(t, "1", id1)
	id2, err := svc.NewID(ctx, "proj-1", idgen.ResourceUsers)
	require.NoError(t, err)
	require.Equal(t, "2", id2)

	cfgSnow := &config.AppConfig{
		Idgen: &config.IdGen{
			DefaultStrategy: "snowflake",
			Snowflake:       &config.IdGen_Snowflake{NodeId: 2},
		},
	}
	svcSnow, err := infraidgen.NewService(cfgSnow, nil, stubProjectRepo{settings: map[string]any{"idgen.users": "snowflake"}})
	require.NoError(t, err)
	sfID, err := svcSnow.NewID(ctx, "proj-1", idgen.ResourceUsers)
	require.NoError(t, err)
	require.NotEmpty(t, sfID)

	cfgRandom := &config.AppConfig{
		Idgen: &config.IdGen{
			Random: &config.IdGen_Random{Length: 8, Charset: "numeric"},
		},
	}
	svcRandom, err := infraidgen.NewService(cfgRandom, nil, stubProjectRepo{settings: map[string]any{"idgen.users": "random"}})
	require.NoError(t, err)
	randID, err := svcRandom.NewID(ctx, "proj-1", idgen.ResourceUsers)
	require.NoError(t, err)
	require.Len(t, randID, 8)
}
