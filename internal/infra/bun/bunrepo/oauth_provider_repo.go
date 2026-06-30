package bunrepo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/deeploop-ai/orionid/internal/domain/projects"
	"github.com/deeploop-ai/orionid/internal/infra/bun/model"
	"github.com/deeploop-ai/orionid/internal/infra/clients"
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
	return &projects.OAuthProvider{
		ProjectID:    m.ProjectID,
		Provider:     m.Provider,
		Enabled:      m.Enabled,
		ClientID:     m.ClientID,
		ClientSecret: m.ClientSecret,
		Scopes:       append([]string(nil), m.Scopes...),
	}, nil
}
