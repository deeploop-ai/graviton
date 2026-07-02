package projects

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIDGenStrategyForResource(t *testing.T) {
	t.Parallel()
	settings := map[string]any{
		"idgen.users": "random",
		"idgen": map[string]any{
			"sessions": "sequence",
		},
	}
	require.Equal(t, "random", IDGenStrategyForResource(settings, "users", "uuid"))
	require.Equal(t, "sequence", IDGenStrategyForResource(settings, "sessions", "uuid"))
	require.Equal(t, "uuid", IDGenStrategyForResource(settings, "documents", "uuid"))
	require.Equal(t, "ulid", IDGenStrategyForResource(map[string]any{"idgen.users": "ulid"}, "users", "uuid"))
	require.Equal(t, "snowflake", IDGenStrategyForResource(map[string]any{"idgen.default": "snowflake"}, "documents", "uuid"))
}
