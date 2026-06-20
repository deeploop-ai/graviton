package users

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateStatus(t *testing.T) {
	for _, s := range []string{StatusActive, StatusInactive, StatusBlocked} {
		require.NoError(t, ValidateStatus(s))
	}
	require.Error(t, ValidateStatus("pending"))
	require.Error(t, ValidateStatus(""))
}

func TestCanAuthenticate(t *testing.T) {
	require.True(t, CanAuthenticate(""))
	require.True(t, CanAuthenticate(StatusActive))
	require.False(t, CanAuthenticate(StatusInactive))
	require.False(t, CanAuthenticate(StatusBlocked))
}
