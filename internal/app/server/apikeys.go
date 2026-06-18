package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/projects"
	"github.com/deeploop-ai/fleet/pkg/idgen"
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

func (a *APIKeys) Get(ctx context.Context, id string) (*projects.APIKey, error) {
	return a.repo.GetAPIKey(ctx, id)
}

func (a *APIKeys) Delete(ctx context.Context, id string) error {
	return a.repo.DeleteAPIKey(ctx, id)
}
