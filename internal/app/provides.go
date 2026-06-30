package app

import (
	"github.com/deeploop-ai/orionid/internal/app/client"
	"github.com/deeploop-ai/orionid/internal/app/console"
	"github.com/deeploop-ai/orionid/internal/app/functions"
	"github.com/deeploop-ai/orionid/internal/app/server"
	"github.com/deeploop-ai/orionid/internal/app/storage"
	domainauth "github.com/deeploop-ai/orionid/internal/domain/auth"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	client.NewUserRoles,
	wire.Bind(new(domainauth.UserRoleResolver), new(*client.UserRoles)),
	client.NewAccount,
	client.NewDatabases,
	client.NewTeams,
	server.NewProjects,
	server.NewUsers,
	server.NewAPIKeys,
	server.NewTeams,
	server.NewDatabases,
	console.NewAuth,
	storage.NewStorage,
	functions.NewFunctions,
)
