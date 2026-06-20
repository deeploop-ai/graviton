package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/deeploop-ai/fleet/internal/infra/bun/bunrepo"
	"github.com/deeploop-ai/fleet/internal/infra/bun/model"
	"github.com/deeploop-ai/fleet/internal/infra/clients"
	"github.com/deeploop-ai/fleet/internal/pkg/config"
	"github.com/deeploop-ai/fleet/pkg/idgen"
	"github.com/deeploop-ai/fleet/pkg/jwtparser"
	"github.com/deeploop-ai/fleet/pkg/password"
)

// CreateTestConsoleAdmin inserts a console admin and returns the model plus cleanup.
func CreateTestConsoleAdmin(ctx context.Context, db *clients.Database, role string) (*model.ConsoleAdmin, func()) {
	hash, err := password.Hash("Admin@123")
	if err != nil {
		panic(err)
	}
	admin := &model.ConsoleAdmin{
		ID:           idgen.UUID().String(),
		Email:        fmt.Sprintf("admin-%d@fleet.local", time.Now().UnixNano()),
		PasswordHash: hash,
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if _, err := db.NewInsert().Model(admin).Exec(ctx); err != nil {
		panic(err)
	}
	cleanup := func() {
		_, _ = db.NewDelete().Model((*model.ConsoleAdmin)(nil)).Where("id = ?", admin.ID).Exec(ctx)
		_, _ = db.NewDelete().Model((*model.ConsoleAdminProject)(nil)).Where("admin_id = ?", admin.ID).Exec(ctx)
	}
	return admin, cleanup
}

// SignConsoleAdminToken issues a console admin JWT compatible with auth.Validator.
func SignConsoleAdminToken(cfg *config.AppConfig, admin *model.ConsoleAdmin) (string, error) {
	now := time.Now()
	return jwtparser.Generate([]byte(cfg.GetSecurity().GetJwt().GetSecret()), jwtparser.Claims{
		TokenID:   idgen.UUID().String(),
		UserID:    admin.ID,
		Username:  admin.Email,
		ActorKind: "admin",
		Roles:     []string{admin.Role},
		ExpiresAt: now.Add(time.Hour).Unix(),
		IssuedAt:  now.Unix(),
	})
}

// GrantConsoleAdminProject links a non-platform admin to a project.
func GrantConsoleAdminProject(ctx context.Context, db *clients.Database, adminID, projectID string) error {
	return bunrepo.NewConsoleAdminProjectRepository(db).GrantProjectAccess(ctx, adminID, projectID)
}
