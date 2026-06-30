package infra

import (
	domainauth "github.com/deeploop-ai/orionid/internal/domain/auth"
	"github.com/deeploop-ai/orionid/internal/infra/auth"
	"github.com/deeploop-ai/orionid/internal/infra/bun"
	"github.com/deeploop-ai/orionid/internal/infra/clients"
	"github.com/deeploop-ai/orionid/internal/infra/documentdb"
	infrafunctions "github.com/deeploop-ai/orionid/internal/infra/functions"
	inframessaging "github.com/deeploop-ai/orionid/internal/infra/messaging"
	infrastorage "github.com/deeploop-ai/orionid/internal/infra/storage"
	"github.com/deeploop-ai/orionid/internal/infra/server"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	clients.NewDataClients,
	clients.NewDatabase,
	clients.NewRedis,

	auth.NewValidator,
	auth.NewSessionService,
	auth.NewRedisOTPChallengeStore,
	auth.NewRedisOAuthStateStore,
	wire.Bind(new(domainauth.SessionService), new(*auth.SessionService)),
	wire.Bind(new(domainauth.OTPChallengeStore), new(*auth.RedisOTPChallengeStore)),
	wire.Bind(new(domainauth.OAuthStateStore), new(*auth.RedisOAuthStateStore)),

	inframessaging.ProviderSet,

	bun.ProviderSet,
	documentdb.ProviderSet,
	infrastorage.ProviderSet,
	infrafunctions.ProviderSet,

	server.NewGRPCServer,
	server.NewGRPCGatewayServer,
	server.NewMetricsServer,
	server.NewHealthCheckFunc,
)
