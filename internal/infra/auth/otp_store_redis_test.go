package auth_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/deeploop-ai/orionid/internal/infra/auth"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRedisOTPChallengeStore_VerifyFlow(t *testing.T) {
	t.Parallel()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := auth.NewRedisOTPChallengeStore(rdb)
	ctx := context.Background()

	require.NoError(t, store.CheckSendRateLimit(ctx, "proj1", "user@example.com", "1.2.3.4"))

	code := "123456"
	challengeID, _, err := store.CreateEmailChallenge(ctx, "proj1", "user@example.com", auth.HashOTP(code))
	require.NoError(t, err)
	require.NotEmpty(t, challengeID)

	err = store.VerifyEmailChallenge(ctx, "proj1", challengeID, "user@example.com", auth.HashOTP("000000"))
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.Unauthenticated, st.Code())

	require.NoError(t, store.VerifyEmailChallenge(ctx, "proj1", challengeID, "user@example.com", auth.HashOTP(code)))

	err = store.VerifyEmailChallenge(ctx, "proj1", challengeID, "user@example.com", auth.HashOTP(code))
	require.Error(t, err)
	st, _ = status.FromError(err)
	require.Equal(t, codes.Unauthenticated, st.Code())
}

func TestRedisOTPChallengeStore_SendCooldown(t *testing.T) {
	t.Parallel()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := auth.NewRedisOTPChallengeStore(rdb)
	ctx := context.Background()

	require.NoError(t, store.CheckSendRateLimit(ctx, "proj1", "user@example.com", ""))
	err = store.CheckSendRateLimit(ctx, "proj1", "user@example.com", "")
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.ResourceExhausted, st.Code())
}

func TestGenerateOTP(t *testing.T) {
	t.Parallel()
	code, err := auth.GenerateOTP(6)
	require.NoError(t, err)
	require.Len(t, code, 6)
}
