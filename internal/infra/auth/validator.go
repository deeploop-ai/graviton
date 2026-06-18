package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/domain/projects"
	"github.com/deeploop-ai/fleet/internal/domain/shared"
	"github.com/deeploop-ai/fleet/internal/pkg/config"
	"github.com/deeploop-ai/fleet/pkg/idgen"
	"github.com/deeploop-ai/fleet/pkg/jwtparser"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Validator struct {
	cfg          *config.AppConfig
	apiKeyRepo   projects.APIKeyRepository
	adminRepo    projects.ConsoleAdminRepository
	docDB        databases.DocumentDB
	sessionCodec *SessionCookieCodec
}

func NewValidator(
	cfg *config.AppConfig,
	apiKeyRepo projects.APIKeyRepository,
	adminRepo projects.ConsoleAdminRepository,
	docDB databases.DocumentDB,
) *Validator {
	return &Validator{
		cfg:          cfg,
		apiKeyRepo:   apiKeyRepo,
		adminRepo:    adminRepo,
		docDB:        docDB,
		sessionCodec: NewSessionCookieCodec(cfg.GetSecurity().GetJwt().GetSecret()),
	}
}

func (v *Validator) ValidateToken(ctx context.Context, token string) (*shared.Principal, error) {
	return v.ValidateCredential(ctx, token, shared.CredentialTypeToken)
}

func (v *Validator) ValidateCredential(ctx context.Context, raw string, credentialType shared.CredentialType) (*shared.Principal, error) {
	switch credentialType {
	case shared.CredentialTypeAPIKey:
		return v.validateAPIKey(ctx, raw)
	case shared.CredentialTypeToken:
		claims, ok := jwtparser.Parse([]byte(v.cfg.GetSecurity().GetJwt().GetSecret()), raw)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}
		return v.principalFromJWT(ctx, claims)
	case shared.CredentialTypeSession:
		// Try JWT first (console or token-style session).
		if claims, ok := jwtparser.Parse([]byte(v.cfg.GetSecurity().GetJwt().GetSecret()), raw); ok {
			return v.principalFromJWT(ctx, claims)
		}
		projectID, sessionID, err := v.sessionCodec.Verify(raw)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid session")
		}
		return v.principalFromSession(ctx, projectID, sessionID)
	}
	return nil, status.Error(codes.Unauthenticated, "unsupported credential type")
}

func (v *Validator) validateAPIKey(ctx context.Context, raw string) (*shared.Principal, error) {
	hash := sha256.Sum256([]byte(raw))
	hashStr := hex.EncodeToString(hash[:])
	key, err := v.apiKeyRepo.GetAPIKeyBySecretHash(ctx, hashStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "api key validation failed")
	}
	if key == nil || !key.Enabled {
		return nil, status.Error(codes.Unauthenticated, "invalid or disabled api key")
	}
	if key.ExpireAt != nil && key.ExpireAt.Before(time.Now()) {
		return nil, status.Error(codes.Unauthenticated, "api key expired")
	}
	return &shared.Principal{
		ActorID:        idgen.ID(key.ID),
		ActorKind:      shared.ActorKindService,
		CredentialType: shared.CredentialTypeAPIKey,
		ProjectID:      key.ProjectID,
		APIKeyID:       key.ID,
		Roles:          []string{"keys"},
		Permissions:    key.Scopes,
	}, nil
}

func (v *Validator) principalFromJWT(ctx context.Context, claims *jwtparser.Claims) (*shared.Principal, error) {
	switch claims.ActorKind {
	case "admin":
		admin, err := v.adminRepo.GetConsoleAdmin(ctx, claims.UserID)
		if err != nil {
			return nil, status.Error(codes.Internal, "admin lookup failed")
		}
		if admin == nil {
			return nil, status.Error(codes.Unauthenticated, "admin not found")
		}
		return &shared.Principal{
			ActorID:         idgen.ID(admin.ID),
			ActorKind:       shared.ActorKindAdmin,
			CredentialType:  shared.CredentialTypeToken,
			IsPlatformAdmin: admin.Role == "owner" || admin.Role == "admin",
			UserID:          admin.ID,
			Email:           admin.Email,
			Roles:           []string{admin.Role},
		}, nil
	default:
		if claims.SessionID != "" && claims.ProjectID != "" {
			if err := v.validateEndUserSession(ctx, claims.ProjectID, claims.SessionID); err != nil {
				return nil, err
			}
		}
		return &shared.Principal{
			ActorID:        idgen.ID(claims.UserID),
			ActorKind:      shared.ActorKindEndUser,
			CredentialType: shared.CredentialTypeToken,
			ProjectID:      claims.ProjectID,
			UserID:         claims.UserID,
			SessionID:      claims.SessionID,
			Email:          claims.Username,
			Roles:          append([]string{"users", fmt.Sprintf("user:%s", claims.UserID)}, claims.Roles...),
		}, nil
	}
}

func (v *Validator) principalFromSession(ctx context.Context, projectID, sessionID string) (*shared.Principal, error) {
	sessionDoc, err := v.docDB.GetDocument(ctx, projectID, "default", "sessions", sessionID, databases.SystemRoles)
	if err != nil {
		return nil, status.Error(codes.Internal, "session lookup failed")
	}
	if sessionDoc == nil {
		return nil, status.Error(codes.Unauthenticated, "session not found")
	}
	expireAtRaw, ok := sessionDoc.Data["expire_at"]
	if ok {
		expireAt, err := parseTime(expireAtRaw)
		if err == nil && expireAt.Before(time.Now()) {
			return nil, status.Error(codes.Unauthenticated, "session expired")
		}
	}
	userID, _ := sessionDoc.Data["user_id"].(string)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid session")
	}
	return &shared.Principal{
		ActorID:        idgen.ID(userID),
		ActorKind:      shared.ActorKindEndUser,
		CredentialType: shared.CredentialTypeSession,
		ProjectID:      projectID,
		UserID:         userID,
		SessionID:      sessionID,
		Roles:          []string{"users", fmt.Sprintf("user:%s", userID)},
	}, nil
}

func (v *Validator) validateEndUserSession(ctx context.Context, projectID, sessionID string) error {
	sessionDoc, err := v.docDB.GetDocument(ctx, projectID, "default", "sessions", sessionID, databases.SystemRoles)
	if err != nil {
		return status.Error(codes.Unauthenticated, "session lookup failed")
	}
	if sessionDoc == nil {
		return status.Error(codes.Unauthenticated, "session not found or revoked")
	}
	if expireAtRaw, ok := sessionDoc.Data["expire_at"]; ok {
		expireAt, err := parseTime(expireAtRaw)
		if err == nil && expireAt.Before(time.Now()) {
			return status.Error(codes.Unauthenticated, "session expired")
		}
	}
	return nil
}

func parseTime(v any) (time.Time, error) {
	switch t := v.(type) {
	case time.Time:
		return t, nil
	case string:
		return time.Parse(time.RFC3339Nano, t)
	}
	return time.Time{}, fmt.Errorf("unsupported time type")
}
