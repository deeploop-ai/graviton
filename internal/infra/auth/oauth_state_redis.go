package auth

import (
	"context"
	"encoding/json"
	"time"

	domainauth "github.com/deeploop-ai/orionid/internal/domain/auth"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const oauthStateTTL = 10 * time.Minute

type RedisOAuthStateStore struct {
	rdb *redis.Client
}

func NewRedisOAuthStateStore(rdb *redis.Client) *RedisOAuthStateStore {
	return &RedisOAuthStateStore{rdb: rdb}
}

func (s *RedisOAuthStateStore) Save(ctx context.Context, state domainauth.OAuthState, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = oauthStateTTL
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return status.Error(codes.Internal, "oauth state encode failed")
	}
	key := oauthStateKey(state.StateID)
	if err := s.rdb.Set(ctx, key, payload, ttl).Err(); err != nil {
		return status.Error(codes.Internal, "oauth state store failed")
	}
	return nil
}

func (s *RedisOAuthStateStore) Get(ctx context.Context, stateID string) (*domainauth.OAuthState, error) {
	raw, err := s.rdb.Get(ctx, oauthStateKey(stateID)).Bytes()
	if err == redis.Nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired oauth state")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "oauth state lookup failed")
	}
	var state domainauth.OAuthState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, status.Error(codes.Internal, "oauth state decode failed")
	}
	return &state, nil
}

func (s *RedisOAuthStateStore) Delete(ctx context.Context, stateID string) error {
	if err := s.rdb.Del(ctx, oauthStateKey(stateID)).Err(); err != nil {
		return status.Error(codes.Internal, "oauth state cleanup failed")
	}
	return nil
}

func oauthStateKey(stateID string) string {
	return "orionid:oauth:state:" + stateID
}
