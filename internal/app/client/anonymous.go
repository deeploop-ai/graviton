package client

import (
	"context"
	"fmt"
	"strings"

	domainauth "github.com/deeploop-ai/graviton/internal/domain/auth"
	"github.com/deeploop-ai/graviton/internal/domain/databases"
	"github.com/deeploop-ai/graviton/internal/domain/users"
	"github.com/deeploop-ai/graviton/pkg/idgen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateAnonymousSessionCommand struct {
	ProjectID string
}

func (a *Account) CreateAnonymousSession(ctx context.Context, cmd CreateAnonymousSessionCommand) (*User, *TokenBundle, string, error) {
	projectID := strings.TrimSpace(cmd.ProjectID)
	if projectID == "" {
		return nil, nil, "", status.Error(codes.InvalidArgument, "project_id is required")
	}
	if err := a.ensureProjectReady(ctx, projectID); err != nil {
		return nil, nil, "", err
	}

	userID := idgen.UUID().String()
	email := anonymousEmail(userID)
	userDoc := databases.Document{
		ID: userID,
		Data: map[string]any{
			"email":          email,
			"password_hash":  "",
			"name":           "Anonymous",
			"status":         users.StatusActive,
			"email_verified": false,
			"labels":         []any{"anonymous"},
			"prefs":          map[string]any{},
		},
	}
	if _, err := a.docDB.CreateDocument(ctx, projectID, "default", "users", userDoc, userDocumentPermissions(userID), databases.SystemPrincipal); err != nil {
		return nil, nil, "", err
	}
	user := mapUserDoc(&userDoc)
	return a.finishSignInWithProvider(ctx, projectID, user, domainauth.ProviderAnonymous)
}

func anonymousEmail(userID string) string {
	shortID := userID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	return fmt.Sprintf("anon_%s@graviton.local", shortID)
}
