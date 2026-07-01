package bunrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/deeploop-ai/graviton/internal/domain/projects"
	"github.com/deeploop-ai/graviton/internal/infra/bun/model"
	"github.com/deeploop-ai/graviton/internal/infra/clients"
)

type oauthProviderRepo struct {
	db *clients.Database
}

func NewOAuthProviderRepository(db *clients.Database) projects.OAuthProviderRepository {
	return &oauthProviderRepo{db: db}
}

func (r *oauthProviderRepo) GetOAuthProvider(ctx context.Context, projectID, provider string) (*projects.OAuthProvider, error) {
	m := new(model.ProjectOAuthProvider)
	err := r.db.NewSelect().Model(m).
		Where("project_id = ? AND provider = ?", projectID, provider).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapOAuthProvider(m), nil
}

func (r *oauthProviderRepo) ListOAuthProviders(ctx context.Context, projectID string) ([]projects.OAuthProvider, error) {
	var rows []model.ProjectOAuthProvider
	err := r.db.NewSelect().Model(&rows).
		Where("project_id = ?", projectID).
		Order("provider ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]projects.OAuthProvider, len(rows))
	for i := range rows {
		out[i] = *mapOAuthProvider(&rows[i])
	}
	return out, nil
}

func (r *oauthProviderRepo) UpsertOAuthProvider(ctx context.Context, cfg *projects.OAuthProvider) error {
	if cfg == nil {
		return errors.New("oauth provider is nil")
	}
	now := time.Now()
	m := &model.ProjectOAuthProvider{
		ProjectID:    cfg.ProjectID,
		Provider:     cfg.Provider,
		Enabled:      cfg.Enabled,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       append([]string(nil), cfg.Scopes...),
		UpdatedAt:    now,
	}
	_, err := r.db.NewInsert().Model(m).
		On("CONFLICT (project_id, provider) DO UPDATE").
		Set("enabled = EXCLUDED.enabled").
		Set("client_id = EXCLUDED.client_id").
		Set("client_secret = EXCLUDED.client_secret").
		Set("scopes = EXCLUDED.scopes").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func (r *oauthProviderRepo) DeleteOAuthProvider(ctx context.Context, projectID, provider string) error {
	_, err := r.db.NewDelete().Model((*model.ProjectOAuthProvider)(nil)).
		Where("project_id = ? AND provider = ?", projectID, provider).
		Exec(ctx)
	return err
}

func mapOAuthProvider(m *model.ProjectOAuthProvider) *projects.OAuthProvider {
	if m == nil {
		return nil
	}
	return &projects.OAuthProvider{
		ProjectID:    m.ProjectID,
		Provider:     m.Provider,
		Enabled:      m.Enabled,
		ClientID:     m.ClientID,
		ClientSecret: m.ClientSecret,
		Scopes:       append([]string(nil), m.Scopes...),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
