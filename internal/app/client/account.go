package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/domain/projects"
	"github.com/deeploop-ai/fleet/internal/infra/auth"
	"github.com/deeploop-ai/fleet/internal/pkg/config"
	"github.com/deeploop-ai/fleet/internal/pkg/contexts"
	"github.com/deeploop-ai/fleet/pkg/idgen"
	"github.com/deeploop-ai/fleet/pkg/jwtparser"
	"github.com/deeploop-ai/fleet/pkg/password"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Account struct {
	cfg          *config.AppConfig
	projectRepo  projects.Repository
	docDB        databases.DocumentDatabase
	sessionCodec *auth.SessionCookieCodec
}

func NewAccount(
	cfg *config.AppConfig,
	projectRepo projects.Repository,
	docDB databases.DocumentDatabase,
) *Account {
	return &Account{
		cfg:          cfg,
		projectRepo:  projectRepo,
		docDB:        docDB,
		sessionCodec: auth.NewSessionCookieCodec(cfg.GetSecurity().GetJwt().GetSecret()),
	}
}

type SignUpCommand struct {
	ProjectID string
	Email     string
	Password  string
	Name      string
}

type SignInCommand struct {
	ProjectID string
	Email     string
	Password  string
}

type User struct {
	ID            string
	Email         string
	Name          string
	Status        string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TokenBundle struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

func (a *Account) SignUp(ctx context.Context, cmd SignUpCommand) (*User, *TokenBundle, string, error) {
	project, err := a.projectRepo.GetProject(ctx, cmd.ProjectID)
	if err != nil {
		return nil, nil, "", err
	}
	if project == nil {
		return nil, nil, "", status.Error(codes.NotFound, "project not found")
	}
	if err := a.docDB.EnsureSystemCollections(ctx, project.ID, project.InternalID); err != nil {
		return nil, nil, "", fmt.Errorf("ensure system collections: %w", err)
	}

	// Check email unique.
	list, err := a.docDB.ListDocuments(ctx, project.ID, "default", "users", databases.Query{
		Queries:  []string{fmt.Sprintf(`equal("email","%s")`, strings.ReplaceAll(cmd.Email, `"`, `""`))},
		PageSize: 1,
	}, []string{"admin"})
	if err != nil {
		return nil, nil, "", err
	}
	if len(list.Documents) > 0 {
		return nil, nil, "", status.Error(codes.AlreadyExists, "email already registered")
	}

	hash, err := password.Hash(cmd.Password)
	if err != nil {
		return nil, nil, "", err
	}

	userID := idgen.UUID().String()
	userDoc := databases.Document{
		ID: userID,
		Data: map[string]any{
			"email":          cmd.Email,
			"password_hash":  hash,
			"name":           cmd.Name,
			"status":         "active",
			"email_verified": false,
			"labels":         []any{},
			"prefs":          map[string]any{},
		},
	}
	userPerms := []databases.Permission{
		{Type: "read", Role: fmt.Sprintf("user:%s", userID)},
		{Type: "read", Role: "keys"},
		{Type: "read", Role: "admin"},
		{Type: "update", Role: fmt.Sprintf("user:%s", userID)},
		{Type: "update", Role: "keys"},
		{Type: "update", Role: "admin"},
		{Type: "delete", Role: fmt.Sprintf("user:%s", userID)},
		{Type: "delete", Role: "keys"},
		{Type: "delete", Role: "admin"},
	}
	if _, err := a.docDB.CreateDocument(ctx, project.ID, "default", "users", userDoc, userPerms); err != nil {
		return nil, nil, "", fmt.Errorf("create user document: %w", err)
	}

	user := mapUserDoc(&userDoc)
	return a.createSessionAndTokens(ctx, project.ID, user.ID, user.Email)
}

func (a *Account) SignIn(ctx context.Context, cmd SignInCommand) (*User, *TokenBundle, string, error) {
	project, err := a.projectRepo.GetProject(ctx, cmd.ProjectID)
	if err != nil {
		return nil, nil, "", err
	}
	if project == nil {
		return nil, nil, "", status.Error(codes.NotFound, "project not found")
	}
	if err := a.docDB.EnsureSystemCollections(ctx, project.ID, project.InternalID); err != nil {
		return nil, nil, "", err
	}

	list, err := a.docDB.ListDocuments(ctx, project.ID, "default", "users", databases.Query{
		Queries:  []string{fmt.Sprintf(`equal("email","%s")`, strings.ReplaceAll(cmd.Email, `"`, `""`))},
		PageSize: 1,
	}, []string{"admin"})
	if err != nil {
		return nil, nil, "", err
	}
	if len(list.Documents) == 0 {
		return nil, nil, "", status.Error(codes.Unauthenticated, "invalid credentials")
	}
	userDoc := list.Documents[0]
	hash, _ := userDoc.Data["password_hash"].(string)
	if ok, _ := password.Verify(cmd.Password, hash); !ok {
		return nil, nil, "", status.Error(codes.Unauthenticated, "invalid credentials")
	}

	user := mapUserDoc(&userDoc)
	return a.createSessionAndTokens(ctx, project.ID, user.ID, user.Email)
}

func (a *Account) Me(ctx context.Context) (*User, error) {
	p, ok := contexts.Principal(ctx)
	if !ok || p.UserID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	doc, err := a.docDB.GetDocument(ctx, p.ProjectID, "default", "users", p.UserID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return mapUserDoc(doc), nil
}

func (a *Account) SignOut(ctx context.Context) error {
	p, ok := contexts.Principal(ctx)
	if !ok || p.SessionID == "" {
		return nil
	}
	return a.docDB.DeleteDocument(ctx, p.ProjectID, "default", "sessions", p.SessionID)
}

func (a *Account) createSessionAndTokens(ctx context.Context, projectID, userID, email string) (*User, *TokenBundle, string, error) {
	expireAt := time.Now().Add(7 * 24 * time.Hour)
	sessionID := idgen.UUID().String()
	sessionSecret := idgen.UUID().String()
	sessionDoc := databases.Document{
		ID: sessionID,
		Data: map[string]any{
			"user_id":      userID,
			"secret_hash":  sessionSecret,
			"provider":     "email",
			"expire_at":    expireAt.Format(time.RFC3339Nano),
			"user_agent":   "",
			"ip":           "",
		},
	}
	sessionPerms := []databases.Permission{
		{Type: "read", Role: fmt.Sprintf("user:%s", userID)},
		{Type: "read", Role: "keys"},
		{Type: "read", Role: "admin"},
		{Type: "update", Role: fmt.Sprintf("user:%s", userID)},
		{Type: "update", Role: "keys"},
		{Type: "update", Role: "admin"},
		{Type: "delete", Role: fmt.Sprintf("user:%s", userID)},
		{Type: "delete", Role: "keys"},
		{Type: "delete", Role: "admin"},
	}
	if _, err := a.docDB.CreateDocument(ctx, projectID, "default", "sessions", sessionDoc, sessionPerms); err != nil {
		return nil, nil, "", err
	}

	accessTTL := 15 * time.Minute
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
		UserID:    userID,
		Username:  email,
		ActorKind: "end_user",
		ProjectID: projectID,
		SessionID: sessionID,
		Roles:     []string{"users", fmt.Sprintf("user:%s", userID)},
		ExpiresAt: now.Add(accessTTL).Unix(),
		IssuedAt:  now.Unix(),
	}
	accessToken, err := jwtparser.Generate([]byte(a.cfg.GetSecurity().GetJwt().GetSecret()), accessClaims)
	if err != nil {
		return nil, nil, "", err
	}
	refreshClaims := accessClaims
	refreshClaims.TokenID = idgen.UUID().String()
	refreshClaims.ExpiresAt = now.Add(refreshTTL).Unix()
	refreshToken, err := jwtparser.Generate([]byte(a.cfg.GetSecurity().GetJwt().GetSecret()), refreshClaims)
	if err != nil {
		return nil, nil, "", err
	}

	cookie := a.sessionCodec.Sign(projectID, sessionID)
	user, _ := a.docDB.GetDocument(ctx, projectID, "default", "users", userID)
	return mapUserDoc(user), &TokenBundle{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessClaims.ExpiresAt,
	}, cookie, nil
}

func mapUserDoc(doc *databases.Document) *User {
	if doc == nil {
		return nil
	}
	return &User{
		ID:            doc.ID,
		Email:         stringValue(doc.Data["email"]),
		Name:          stringValue(doc.Data["name"]),
		Status:        stringValue(doc.Data["status"]),
		EmailVerified: boolValue(doc.Data["email_verified"]),
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
	}
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func boolValue(v any) bool {
	b, _ := v.(bool)
	return b
}
