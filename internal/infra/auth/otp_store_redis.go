package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/deeploop-ai/orionid/pkg/idgen"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	otpChallengeTTL   = 5 * time.Minute
	otpSendCooldown   = 60 * time.Second
	otpIPWindow       = time.Hour
	otpIPMaxPerWindow = 10
	otpMaxAttempts    = 5
)

type otpChallengeRecord struct {
	ProjectID string `json:"project_id"`
	Email     string `json:"email"`
	CodeHash  string `json:"code_hash"`
	Attempts  int    `json:"attempts"`
}

// RedisOTPChallengeStore stores OTP challenges in Redis.
type RedisOTPChallengeStore struct {
	rdb *redis.Client
}

func NewRedisOTPChallengeStore(rdb *redis.Client) *RedisOTPChallengeStore {
	return &RedisOTPChallengeStore{rdb: rdb}
}

func (s *RedisOTPChallengeStore) CheckSendRateLimit(ctx context.Context, projectID, email, ip string) error {
	sendKey := fmt.Sprintf("orionid:otp:send:%s:%s", projectID, email)
	ok, err := s.rdb.SetNX(ctx, sendKey, "1", otpSendCooldown).Result()
	if err != nil {
		return status.Error(codes.Internal, "otp rate limit check failed")
	}
	if !ok {
		return status.Error(codes.ResourceExhausted, "otp send cooldown active")
	}

	if ip == "" {
		return nil
	}
	ipKey := fmt.Sprintf("orionid:otp:ip:%s:%s", projectID, ip)
	count, err := s.rdb.Incr(ctx, ipKey).Result()
	if err != nil {
		return status.Error(codes.Internal, "otp ip rate limit check failed")
	}
	if count == 1 {
		if err := s.rdb.Expire(ctx, ipKey, otpIPWindow).Err(); err != nil {
			return status.Error(codes.Internal, "otp ip rate limit check failed")
		}
	}
	if count > otpIPMaxPerWindow {
		return status.Error(codes.ResourceExhausted, "otp ip rate limit exceeded")
	}
	return nil
}

func (s *RedisOTPChallengeStore) CreateEmailChallenge(ctx context.Context, projectID, email, codeHash string) (string, time.Time, error) {
	challengeID := newChallengeID()
	expireAt := time.Now().Add(otpChallengeTTL)
	record := otpChallengeRecord{
		ProjectID: projectID,
		Email:     email,
		CodeHash:  codeHash,
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return "", time.Time{}, status.Error(codes.Internal, "otp challenge encode failed")
	}
	key := challengeKey(challengeID)
	if err := s.rdb.Set(ctx, key, payload, otpChallengeTTL).Err(); err != nil {
		return "", time.Time{}, status.Error(codes.Internal, "otp challenge store failed")
	}
	return challengeID, expireAt, nil
}

func (s *RedisOTPChallengeStore) VerifyEmailChallenge(ctx context.Context, projectID, challengeID, email, codeHash string) error {
	key := challengeKey(challengeID)
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return status.Error(codes.Unauthenticated, "invalid or expired otp challenge")
	}
	if err != nil {
		return status.Error(codes.Internal, "otp challenge lookup failed")
	}

	var record otpChallengeRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return status.Error(codes.Internal, "otp challenge decode failed")
	}
	if record.ProjectID != projectID || record.Email != email {
		return status.Error(codes.Unauthenticated, "invalid or expired otp challenge")
	}
	if record.Attempts >= otpMaxAttempts {
		_ = s.rdb.Del(ctx, key).Err()
		return status.Error(codes.ResourceExhausted, "otp attempts exceeded")
	}

	record.Attempts++
	if record.CodeHash != codeHash {
		payload, encErr := json.Marshal(record)
		if encErr != nil {
			return status.Error(codes.Internal, "otp challenge encode failed")
		}
		ttl, ttlErr := s.rdb.TTL(ctx, key).Result()
		if ttlErr != nil || ttl <= 0 {
			ttl = otpChallengeTTL
		}
		if err := s.rdb.Set(ctx, key, payload, ttl).Err(); err != nil {
			return status.Error(codes.Internal, "otp challenge update failed")
		}
		return status.Error(codes.Unauthenticated, "invalid otp code")
	}

	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return status.Error(codes.Internal, "otp challenge cleanup failed")
	}
	return nil
}

func challengeKey(challengeID string) string {
	return "orionid:otp:ch:" + challengeID
}

func newChallengeID() string {
	return idgen.UUID().String()
}
