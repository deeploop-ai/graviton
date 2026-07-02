package server

import (
	"testing"

	clientv1 "github.com/deeploop-ai/graviton/genproto/client/v1"
	"github.com/stretchr/testify/require"
)

func TestCollectMethodsByAccess_AuthenticatedRequiresUsersRole(t *testing.T) {
	t.Parallel()

	_, _, permissionMethods, err := collectMethodsByAccess(clientv1.File_client_v1_teams_proto)
	require.NoError(t, err)

	fullMethod := "/graviton.client.v1.TeamsService/CreateTeam"
	perms, ok := permissionMethods[fullMethod]
	require.True(t, ok, "TeamsService methods should require users permission")
	require.Equal(t, []string{"users"}, perms)
}

func TestCollectMethodsByAccess_AccountPublicMethods(t *testing.T) {
	t.Parallel()

	publicMethods, _, permissionMethods, err := collectMethodsByAccess(clientv1.File_client_v1_account_proto)
	require.NoError(t, err)

	require.Contains(t, publicMethods, "/graviton.client.v1.AccountService/SignIn")
	require.Equal(t, []string{"users"}, permissionMethods["/graviton.client.v1.AccountService/Me"])
}
