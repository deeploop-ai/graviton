package infra

import (
	"github.com/deeploop-ai/orionid/internal/infra/auth"
	"github.com/deeploop-ai/orionid/internal/infra/bun"
	"github.com/deeploop-ai/orionid/internal/infra/clients"
	"github.com/deeploop-ai/orionid/internal/infra/documentdb"
	infrafunctions "github.com/deeploop-ai/orionid/internal/infra/functions"
	infrastorage "github.com/deeploop-ai/orionid/internal/infra/storage"
	"github.com/deeploop-ai/orionid/internal/infra/server"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	clients.NewDataClients,
	clients.NewDatabase,

	auth.NewValidator,

	bun.ProviderSet,
	documentdb.ProviderSet,
	infrastorage.ProviderSet,
	infrafunctions.ProviderSet,

	server.NewGRPCServer,
	server.NewGRPCGatewayServer,
	server.NewMetricsServer,
	server.NewHealthCheckFunc,
)
