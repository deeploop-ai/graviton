package api

import (
	"github.com/deeploop-ai/fleet/internal/api/clientgrpc"
	"github.com/deeploop-ai/fleet/internal/api/consolegrpc"
	"github.com/deeploop-ai/fleet/internal/api/servergrpc"
	"github.com/deeploop-ai/fleet/internal/api/serverhttp"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	clientgrpc.NewAccountService,
	clientgrpc.NewDatabasesService,
	servergrpc.NewHealthService,
	servergrpc.NewProjectsService,
	servergrpc.NewStorageService,
	servergrpc.NewUsersService,
	servergrpc.NewAPIKeysService,
	servergrpc.NewTeamsService,
	servergrpc.NewDatabasesService,
	serverhttp.NewFileHandler,
	consolegrpc.NewAuthService,
)
