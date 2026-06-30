package client

import (
	"context"

	domainauth "github.com/deeploop-ai/orionid/internal/domain/auth"
	"github.com/deeploop-ai/orionid/internal/domain/databases"
	"github.com/deeploop-ai/orionid/internal/domain/messaging"
	"github.com/deeploop-ai/orionid/internal/domain/projects"
	infraauth "github.com/deeploop-ai/orionid/internal/infra/auth"
	inframessaging "github.com/deeploop-ai/orionid/internal/infra/messaging"
	"github.com/deeploop-ai/orionid/internal/pkg/config"
	"github.com/redis/go-redis/v9"
)

func NewTestAccount(cfg *config.AppConfig, projectRepo projects.Repository, docDB databases.DocumentDB) *Account {
	return NewTestAccountWithRedis(cfg, projectRepo, docDB, nil)
}

func NewTestAccountWithRedis(cfg *config.AppConfig, projectRepo projects.Repository, docDB databases.DocumentDB, rdb *redis.Client) *Account {
	roles := NewUserRoles(docDB)
	sessions := infraauth.NewSessionService(cfg, docDB, roles)
	var otp domainauth.OTPChallengeStore
	if rdb != nil {
		otp = infraauth.NewRedisOTPChallengeStore(rdb)
	}
	mailer := inframessaging.NewMailer(cfg)
	return NewAccount(cfg, projectRepo, docDB, sessions, otp, mailer)
}

// CaptureMailer records sent messages for tests.
type CaptureMailer struct {
	Subjects []string
	Bodies   []string
}

func (m *CaptureMailer) Send(_ context.Context, _, subject, body string) error {
	m.Subjects = append(m.Subjects, subject)
	m.Bodies = append(m.Bodies, body)
	return nil
}

func NewTestAccountWithMailer(cfg *config.AppConfig, projectRepo projects.Repository, docDB databases.DocumentDB, rdb *redis.Client, mailer messaging.Mailer) *Account {
	roles := NewUserRoles(docDB)
	sessions := infraauth.NewSessionService(cfg, docDB, roles)
	var otp domainauth.OTPChallengeStore
	if rdb != nil {
		otp = infraauth.NewRedisOTPChallengeStore(rdb)
	}
	return NewAccount(cfg, projectRepo, docDB, sessions, otp, mailer)
}
