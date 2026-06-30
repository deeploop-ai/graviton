package auth

import (
	"context"
	"time"
)

// OTPChallengeStore persists one-time password challenges and enforces send rate limits.
type OTPChallengeStore interface {
	CheckSendRateLimit(ctx context.Context, projectID, email, ip string) error
	CreateEmailChallenge(ctx context.Context, projectID, email, codeHash string) (challengeID string, expireAt time.Time, err error)
	VerifyEmailChallenge(ctx context.Context, projectID, challengeID, email, codeHash string) error
}
