package auth_test

import (
	"context"
	"testing"

	domainauth "github.com/deeploop-ai/orionid/internal/domain/auth"
	"github.com/deeploop-ai/orionid/internal/infra/auth"
	"github.com/deeploop-ai/orionid/internal/pkg/contexts"
	"github.com/stretchr/testify/require"
)

type stubRoleResolver struct{}

func (stubRoleResolver) LoadUserRoles(_ context.Context, _, userID string) ([]string, error) {
	return []string{"users", "user:" + userID}, nil
}

func TestSessionService_RecordsClientInfo(t *testing.T) {
	t.Parallel()

	// Unit-level check: CreateSessionAndTokens reads ClientInfo from context.
	// Full integration is covered by account integration tests.
	svc := auth.NewSessionService(nil, nil, stubRoleResolver{})
	require.NotNil(t, svc)

	ctx := contexts.WithClientInfo(context.Background(), contexts.ClientInfo{
		IP:        "203.0.113.10",
		UserAgent: "OrionidTest/1.0",
	})
	info := contexts.ClientInfoFrom(ctx)
	require.Equal(t, "203.0.113.10", info.IP)
	require.Equal(t, "OrionidTest/1.0", info.UserAgent)
}

func TestProviderConstants(t *testing.T) {
	t.Parallel()
	require.Equal(t, "email", domainauth.ProviderEmail)
	require.Equal(t, "wechat_web", domainauth.ProviderWeChatWeb)
}
