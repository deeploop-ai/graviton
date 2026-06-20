package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/projects"
	"github.com/deeploop-ai/fleet/pkg/idgen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type APIKeys struct {
	repo projects.APIKeyRepository
}

func NewAPIKeys(repo projects.APIKeyRepository) *APIKeys {
	return &APIKeys{repo: repo}
}

type CreateAPIKeyCommand struct {
	ProjectID string
	Name      string
	Scopes    []string
	ExpireAt  *time.Time
}

func (a *APIKeys) Create(ctx context.Context, cmd CreateAPIKeyCommand) (*projects.APIKey, string, error) {
	if cmd.Name == "" {
		return nil, "", status.Error(codes.InvalidArgument, "name is required")
	}
	id := idgen.UUID().String()
	secret := idgen.UUID().String() + idgen.UUID().String()
	hash := sha256.Sum256([]byte(secret))
	key := &projects.APIKey{
		ID:         id,
		ProjectID:  cmd.ProjectID,
		Name:       cmd.Name,
		SecretHash: hex.EncodeToString(hash[:]),
		Scopes:     cmd.Scopes,
		ExpireAt:   cmd.ExpireAt,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := a.repo.CreateAPIKey(ctx, key); err != nil {
		return nil, "", fmt.Errorf("create api key: %w", err)
	}
	return key, secret, nil
}

func (a *APIKeys) List(ctx context.Context, projectID string) ([]projects.APIKey, error) {
	return a.repo.ListAPIKeys(ctx, projectID)
}

func (a *APIKeys) Get(ctx context.Context, projectID, id string) (*projects.APIKey, error) {
	key, err := a.repo.GetAPIKey(ctx, id)
	if err != nil {
		return nil, err
	}
	if key != nil && key.ProjectID != projectID {
		return nil, nil
	}
	return key, nil
}

func (a *APIKeys) Delete(ctx context.Context, projectID, id string) error {
	key, err := a.repo.GetAPIKey(ctx, id)
	if err != nil {
		return err
	}
	if key == nil {
		return fmt.Errorf("api key not found")
	}
	if key.ProjectID != projectID {
		return fmt.Errorf("api key not found")
	}
	return a.repo.DeleteAPIKey(ctx, id)
}
