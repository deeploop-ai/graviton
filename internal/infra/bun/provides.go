package bun

import (
	"github.com/deeploop-ai/fleet/internal/infra/bun/bunrepo"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	bunrepo.NewProjectRepository,
	bunrepo.NewAPIKeyRepository,
	bunrepo.NewConsoleAdminRepository,
	bunrepo.NewConsoleAdminProjectRepository,
	bunrepo.NewAuditRepository,
)
