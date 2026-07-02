package console

import (
	"context"
	"time"

	domainauth "github.com/deeploop-ai/graviton/internal/domain/auth"
	"github.com/deeploop-ai/graviton/internal/domain/projects"
	"github.com/deeploop-ai/graviton/internal/domain/shared"
	"github.com/deeploop-ai/graviton/internal/pkg/config"
	"github.com/deeploop-ai/graviton/internal/pkg/contexts"
	"github.com/deeploop-ai/graviton/pkg/idgen"
	"github.com/deeploop-ai/graviton/pkg/jwtparser"
	"github.com/deeploop-ai/graviton/pkg/password"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth struct {
	cfg              *config.AppConfig
	adminRepo        projects.ConsoleAdminRepository
	adminRevokeStore domainauth.AdminTokenRevokeStore
}

func NewAuth(cfg *config.AppConfig, adminRepo projects.ConsoleAdminRepository, adminRevokeStore domainauth.AdminTokenRevokeStore) *Auth {
	return &Auth{cfg: cfg, adminRepo: adminRepo, adminRevokeStore: adminRevokeStore}
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

func (a *Auth) RefreshToken(ctx context.Context, cmd RefreshTokenCommand) (*TokenPair, error) {
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
	if err := a.checkAdminTokenRevoked(ctx, claims); err != nil {
		return nil, err
	}
	return a.issueAdminTokens(claims.UserID, claims.Username, firstRole(claims.Roles))
}

func (a *Auth) SignOut(ctx context.Context) error {
	p, ok := contexts.Principal(ctx)
	if !ok || p.ActorKind != shared.ActorKindAdmin || p.UserID == "" || a.adminRevokeStore == nil {
		return nil
	}
	refreshTTL := 7 * 24 * time.Hour
	if d, err := time.ParseDuration(a.cfg.GetSecurity().GetJwt().GetRefreshTtl()); err == nil {
		refreshTTL = d
	}
	return a.adminRevokeStore.RevokeBefore(ctx, p.UserID, time.Now(), refreshTTL)
}

func (a *Auth) checkAdminTokenRevoked(ctx context.Context, claims *jwtparser.Claims) error {
	if a.adminRevokeStore == nil || claims == nil || claims.UserID == "" {
		return nil
	}
	revokedBefore, err := a.adminRevokeStore.RevokedBefore(ctx, claims.UserID)
	if err != nil {
		return err
	}
	if !revokedBefore.IsZero() && claims.IssuedAt < revokedBefore.Unix() {
		return status.Error(codes.Unauthenticated, "token revoked")
	}
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
