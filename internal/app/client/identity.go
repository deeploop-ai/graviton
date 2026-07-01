package client

import (
	"context"
	"fmt"
	"strings"

	domainauth "github.com/deeploop-ai/graviton/internal/domain/auth"
	"github.com/deeploop-ai/graviton/internal/domain/databases"
	"github.com/deeploop-ai/graviton/pkg/idgen"
	"github.com/deeploop-ai/graviton/pkg/query"
)

func (a *Account) findIdentity(ctx context.Context, projectID, provider, providerUID string) (*domainauth.Identity, error) {
	list, err := a.docDB.ListDocuments(ctx, projectID, "default", "identities", databases.Query{
		Queries: []string{
			query.BuildEqual("provider", provider),
			query.BuildEqual("provider_uid", providerUID),
		},
		PageSize: 1,
	}, databases.SystemPrincipal)
	if err != nil {
		return nil, err
	}
	if len(list.Documents) == 0 {
		return nil, nil
	}
	return mapIdentityDoc(&list.Documents[0]), nil
}

func (a *Account) createIdentity(ctx context.Context, projectID, userID string, info *domainauth.OAuthUserInfo, provider string) error {
	identityID := idgen.UUID().String()
	providerData := map[string]any{
		"name":       info.Name,
		"avatar_url": info.AvatarURL,
		"raw":        info.Raw,
	}
	if info.OpenID != "" {
		providerData["openid"] = info.OpenID
	}
	if info.UnionID != "" {
		providerData["unionid"] = info.UnionID
	}
	doc := databases.Document{
		ID: identityID,
		Data: map[string]any{
			"user_id":        userID,
			"provider":       provider,
			"provider_uid":   info.ProviderUID,
			"provider_email": info.Email,
			"provider_data":  providerData,
		},
	}
	perms := []databases.Permission{
		{Type: "read", Role: fmt.Sprintf("user:%s", userID)},
		{Type: "read", Role: "keys"},
		{Type: "read", Role: "admin"},
		{Type: "delete", Role: fmt.Sprintf("user:%s", userID)},
		{Type: "delete", Role: "keys"},
		{Type: "delete", Role: "admin"},
	}
	_, err := a.docDB.CreateDocument(ctx, projectID, "default", "identities", doc, perms, databases.SystemPrincipal)
	return err
}

func mapIdentityDoc(doc *databases.Document) *domainauth.Identity {
	if doc == nil {
		return nil
	}
	identity := &domainauth.Identity{
		ID:          doc.ID,
		UserID:      stringValue(doc.Data["user_id"]),
		Provider:    stringValue(doc.Data["provider"]),
		ProviderUID: stringValue(doc.Data["provider_uid"]),
		ProviderEmail: stringValue(doc.Data["provider_email"]),
	}
	if raw, ok := doc.Data["provider_data"].(map[string]any); ok {
		identity.ProviderData = raw
	}
	return identity
}

func (a *Account) resolveOAuthUser(ctx context.Context, projectID, provider string, info *domainauth.OAuthUserInfo) (*User, error) {
	if domainauth.IsWeChatProvider(provider) {
		return a.resolveWeChatUser(ctx, projectID, provider, info)
	}
	if info == nil || info.ProviderUID == "" {
		return nil, fmt.Errorf("oauth profile missing provider uid")
	}
	if strings.TrimSpace(info.Email) == "" {
		return nil, fmt.Errorf("oauth provider did not return an email address")
	}
	identity, err := a.findIdentity(ctx, projectID, provider, info.ProviderUID)
	if err != nil {
		return nil, err
	}
	if identity != nil {
		doc, err := a.docDB.GetDocument(ctx, projectID, "default", "users", identity.UserID, databases.SystemPrincipal)
		if err != nil {
			return nil, err
		}
		if doc == nil {
			return nil, fmt.Errorf("identity references missing user")
		}
		return mapUserDoc(doc), nil
	}

	if info.Email != "" {
		list, err := a.docDB.ListDocuments(ctx, projectID, "default", "users", databases.Query{
			Queries:  []string{query.BuildEqual("email", info.Email)},
			PageSize: 1,
		}, databases.SystemPrincipal)
		if err != nil {
			return nil, err
		}
		if len(list.Documents) > 0 {
			user := mapUserDoc(&list.Documents[0])
			if err := a.createIdentity(ctx, projectID, user.ID, info, provider); err != nil {
				return nil, err
			}
			return user, nil
		}
	}

	name := info.Name
	if name == "" && info.Email != "" {
		name = emailLocalPart(info.Email)
	}
	user, err := a.findOrCreateUserByEmail(ctx, projectID, info.Email, info.Email != "")
	if err != nil {
		return nil, err
	}
	if name != "" && user.Name == emailLocalPart(info.Email) {
		updated, err := a.docDB.UpdateDocument(ctx, projectID, "default", "users", databases.SimpleDocumentUpdate(databases.Document{
			ID:   user.ID,
			Data: map[string]any{"name": name},
		}, nil), databases.SystemPrincipal)
		if err == nil {
			user = mapUserDoc(&updated)
		}
	}
	if err := a.createIdentity(ctx, projectID, user.ID, info, provider); err != nil {
		return nil, err
	}
	return user, nil
}
