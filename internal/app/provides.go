package app

import (
	"github.com/deeploop-ai/fleet/internal/app/client"
	"github.com/deeploop-ai/fleet/internal/app/console"
	"github.com/deeploop-ai/fleet/internal/app/functions"
	"github.com/deeploop-ai/fleet/internal/app/server"
	"github.com/deeploop-ai/fleet/internal/app/storage"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	client.NewAccount,
	client.NewDatabases,
	server.NewProjects,
	server.NewUsers,
	server.NewAPIKeys,
	server.NewTeams,
	server.NewDatabases,
	console.NewAuth,
	storage.NewStorage,
	functions.NewFunctions,
)
