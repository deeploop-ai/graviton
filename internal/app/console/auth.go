package console

import (
	"context"
	"time"

	"github.com/deeploop-ai/graviton/internal/domain/projects"
	"github.com/deeploop-ai/graviton/internal/pkg/config"
	"github.com/deeploop-ai/graviton/pkg/idgen"
	"github.com/deeploop-ai/graviton/pkg/jwtparser"
	"github.com/deeploop-ai/graviton/pkg/password"
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

type RefreshTokenCommand struct {
	RefreshToken string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

func (a *Auth) SignIn(ctx context.Context, cmd SignInCommand) (*TokenPair, error) {
	admin, err := a.adminRepo.GetConsoleAdminByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "admin lookup failed")
	}
	if admin == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if ok, _ := password.Verify(cmd.Password, admin.PasswordHash); !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return a.issueAdminTokens(admin.ID, admin.Email, admin.Role)
}

func (a *Auth) RefreshToken(_ context.Context, cmd RefreshTokenCommand) (*TokenPair, error) {
	if cmd.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}
	claims, ok := jwtparser.Parse([]byte(a.cfg.GetSecurity().GetJwt().GetSecret()), cmd.RefreshToken)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	if claims.TokenType != jwtparser.TokenTypeRefresh || claims.ActorKind != "admin" {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	return a.issueAdminTokens(claims.UserID, claims.Username, firstRole(claims.Roles))
}

func (a *Auth) SignOut(context.Context) error {
	return nil
}

func (a *Auth) issueAdminTokens(adminID, email, role string) (*TokenPair, error) {
	accessTTL := 24 * time.Hour
	if d, err := time.ParseDuration(a.cfg.GetSecurity().GetJwt().GetAccessTtl()); err == nil {
		accessTTL = d
	}
	refreshTTL := 7 * 24 * time.Hour
	if d, err := time.ParseDuration(a.cfg.GetSecurity().GetJwt().GetRefreshTtl()); err == nil {
		refreshTTL = d
	}
	now := time.Now()
	accessClaims := jwtparser.Claims{
		TokenID:   idgen.UUID().String(),
		UserID:    adminID,
		Username:  email,
		ActorKind: "admin",
		Roles:     []string{role},
		TokenType: jwtparser.TokenTypeAccess,
		ExpiresAt: now.Add(accessTTL).Unix(),
		IssuedAt:  now.Unix(),
	}
	accessToken, err := jwtparser.Generate([]byte(a.cfg.GetSecurity().GetJwt().GetSecret()), accessClaims)
	if err != nil {
		return nil, err
	}
	refreshClaims := accessClaims
	refreshClaims.TokenID = idgen.UUID().String()
	refreshClaims.TokenType = jwtparser.TokenTypeRefresh
	refreshClaims.ExpiresAt = now.Add(refreshTTL).Unix()
	refreshToken, err := jwtparser.Generate([]byte(a.cfg.GetSecurity().GetJwt().GetSecret()), refreshClaims)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessClaims.ExpiresAt,
	}, nil
}

func firstRole(roles []string) string {
	if len(roles) == 0 {
		return "admin"
	}
	return roles[0]
}
