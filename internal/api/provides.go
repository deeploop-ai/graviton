package api

import (
	"github.com/deeploop-ai/orionid/internal/api/clientgrpc"
	"github.com/deeploop-ai/orionid/internal/api/consolegrpc"
	"github.com/deeploop-ai/orionid/internal/api/servergrpc"
	"github.com/deeploop-ai/orionid/internal/api/serverhttp"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	clientgrpc.NewAccountService,
	clientgrpc.NewDatabasesService,
	clientgrpc.NewTeamsService,
	servergrpc.NewHealthService,
	servergrpc.NewProjectsService,
	servergrpc.NewStorageService,
	servergrpc.NewUsersService,
	servergrpc.NewAPIKeysService,
	servergrpc.NewOAuthProvidersService,
	servergrpc.NewTeamsService,
	servergrpc.NewDatabasesService,
	serverhttp.NewFileHandler,
	serverhttp.NewOAuthHandler,
	consolegrpc.NewAuthService,
)
