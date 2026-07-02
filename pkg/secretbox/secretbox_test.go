package secretbox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	t.Parallel()
	secret := "test-secret-key"
	plain := "super-secret-oauth-client-secret"
	enc, err := Encrypt(plain, secret)
	require.NoError(t, err)
	require.True(t, stringsHasPrefix(enc, prefix))

	out, err := Decrypt(enc, secret)
	require.NoError(t, err)
	require.Equal(t, plain, out)
}

func TestDecryptLegacyPlaintext(t *testing.T) {
	t.Parallel()
	out, err := Decrypt("plain-text-secret", "test-secret-key")
	require.NoError(t, err)
	require.Equal(t, "plain-text-secret", out)
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
