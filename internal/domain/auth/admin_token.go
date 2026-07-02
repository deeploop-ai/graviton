package auth

import (
	"context"
	"time"
)

// AdminTokenRevokeStore tracks admin token revocations. Sign-out invalidates all
// tokens issued before the recorded timestamp (covers access + refresh JWTs).
type AdminTokenRevokeStore interface {
	RevokeBefore(ctx context.Context, adminID string, revokedAt time.Time, ttl time.Duration) error
	RevokedBefore(ctx context.Context, adminID string) (time.Time, error)
}
