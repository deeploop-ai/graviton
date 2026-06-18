package console

import (
	"context"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/projects"
	"github.com/deeploop-ai/fleet/internal/pkg/config"
	"github.com/deeploop-ai/fleet/pkg/idgen"
	"github.com/deeploop-ai/fleet/pkg/jwtparser"
	"github.com/deeploop-ai/fleet/pkg/password"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth struct {
	cfg       *config.AppConfig
	adminRepo projects.ConsoleAdminRepository
}

func NewAuth(cfg *config.AppConfig, adminRepo projects.ConsoleAdminRepository) *Auth {
	return &Auth{cfg: cfg, adminRepo: adminRepo}
}

type SignInCommand struct {
	Email    string
	Password string
}

func (a *Auth) SignIn(ctx context.Context, cmd SignInCommand) (string, int64, error) {
	admin, err := a.adminRepo.GetConsoleAdminByEmail(ctx, cmd.Email)
	if err != nil {
		return "", 0, status.Error(codes.Internal, "admin lookup failed")
	}
	if admin == nil {
		return "", 0, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if ok, _ := password.Verify(cmd.Password, admin.PasswordHash); !ok {
		return "", 0, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	ttl := 24 * time.Hour
	if d, err := time.ParseDuration(a.cfg.GetSecurity().GetJwt().GetAccessTtl()); err == nil {
		ttl = d
	}
	now := time.Now()
	claims := jwtparser.Claims{
		TokenID:   idgen.UUID().String(),
		UserID:    admin.ID,
		Username:  admin.Email,
		ActorKind: "admin",
		Roles:     []string{admin.Role},
		ExpiresAt: now.Add(ttl).Unix(),
		IssuedAt:  now.Unix(),
	}
	token, err := jwtparser.Generate([]byte(a.cfg.GetSecurity().GetJwt().GetSecret()), claims)
	if err != nil {
		return "", 0, err
	}
	return token, claims.ExpiresAt, nil
}
