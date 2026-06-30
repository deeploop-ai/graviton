package server

import (
	"fmt"
	"time"

	clientv1 "github.com/deeploop-ai/orionid/genproto/client/v1"
	consolev1 "github.com/deeploop-ai/orionid/genproto/console/v1"
	serverv1 "github.com/deeploop-ai/orionid/genproto/server/v1"
	sharedv1 "github.com/deeploop-ai/orionid/genproto/shared/v1"
	"github.com/deeploop-ai/orionid/internal/api/clientgrpc"
	"github.com/deeploop-ai/orionid/internal/api/consolegrpc"
	"github.com/deeploop-ai/orionid/internal/api/servergrpc"
	"github.com/deeploop-ai/orionid/internal/domain/audit"
	"github.com/deeploop-ai/orionid/internal/infra/auth"
	"github.com/deeploop-ai/orionid/internal/pkg/config"
	"github.com/deeploop-ai/orionid/pkg/grpc/interceptor"
	"github.com/lynx-go/lynx"
	lynxgrpc "github.com/lynx-go/lynx/server/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func NewGRPCServer(
	app lynx.Lynx,
	cfg *config.AppConfig,
	validator *auth.Validator,
	auditRepo audit.Repository,
	account *clientgrpc.AccountService,
	clientDatabases *clientgrpc.DatabasesService,
	clientTeams *clientgrpc.TeamsService,
	health *servergrpc.HealthService,
	projects *servergrpc.ProjectsService,
	storage *servergrpc.StorageService,
	users *servergrpc.UsersService,
	apiKeys *servergrpc.APIKeysService,
	teams *servergrpc.TeamsService,
	databases *servergrpc.DatabasesService,
	consoleAuth *consolegrpc.AuthService,
) (*lynxgrpc.Server, error) {
	grpcCfg := cfg.GetServer().GetGrpc()
	timeout := parseDuration(grpcCfg.GetTimeout(), 30*time.Second)

	publicMethods, apiKeyMethods, permissionMethods, err := collectMethodsByAccess(
		clientv1.File_client_v1_account_proto,
		clientv1.File_client_v1_databases_proto,
		clientv1.File_client_v1_teams_proto,
		serverv1.File_server_v1_projects_proto,
		serverv1.File_server_v1_health_proto,
		serverv1.File_server_v1_storage_proto,
		serverv1.File_server_v1_users_proto,
		serverv1.File_server_v1_apikeys_proto,
		serverv1.File_server_v1_teams_proto,
		serverv1.File_server_v1_databases_proto,
		consolev1.File_console_v1_auth_proto,
	)
	if err != nil {
		return nil, err
	}

	authInterceptor, err := interceptor.NewAuthInterceptor(validator, publicMethods, apiKeyMethods, permissionMethods)
	if err != nil {
		return nil, err
	}
	auditInterceptor := interceptor.NewAuditInterceptor(auditRepo)

	srv := lynxgrpc.NewServer(
		lynxgrpc.WithAddr(grpcCfg.GetAddr()),
		lynxgrpc.WithTimeout(timeout),
		lynxgrpc.WithLogger(app.Logger()),
		lynxgrpc.WithInterceptors(
			authInterceptor.UnaryAuthMiddleware,
			auditInterceptor.UnaryAuditMiddleware,
		),
	)
	grpcSrv := srv.GetServer()

	clientv1.RegisterAccountServiceServer(grpcSrv, account)
	clientv1.RegisterDatabasesServiceServer(grpcSrv, clientDatabases)
	clientv1.RegisterTeamsServiceServer(grpcSrv, clientTeams)
	serverv1.RegisterHealthServiceServer(grpcSrv, health)
	serverv1.RegisterProjectsServiceServer(grpcSrv, projects)
	serverv1.RegisterStorageServiceServer(grpcSrv, storage)
	serverv1.RegisterUsersServiceServer(grpcSrv, users)
	serverv1.RegisterAPIKeysServiceServer(grpcSrv, apiKeys)
	serverv1.RegisterTeamsServiceServer(grpcSrv, teams)
	serverv1.RegisterDatabasesServiceServer(grpcSrv, databases)
	consolev1.RegisterConsoleAuthServiceServer(grpcSrv, consoleAuth)

	return srv, nil
}

func collectMethodsByAccess(fileDescs ...protoreflect.FileDescriptor) (publicMethods []string, apiKeyMethods []string, permissionMethods map[string][]string, err error) {
	permissionMethods = make(map[string][]string)
	for _, fileDesc := range fileDescs {
		services := fileDesc.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			serviceDefault := resolveServiceDefaultAccess(service)
			methods := service.Methods()
			for j := 0; j < methods.Len(); j++ {
				method := methods.Get(j)
				access, perms, ok := resolveMethodAccess(method, serviceDefault)
				if !ok || access == sharedv1.AccessLevel_ACCESS_LEVEL_UNSPECIFIED {
					return nil, nil, nil, fmt.Errorf("missing auth policy for method %s/%s", service.FullName(), method.Name())
				}
				fullMethod := fmt.Sprintf("/%s/%s", service.FullName(), method.Name())
				switch access {
				case sharedv1.AccessLevel_ACCESS_PUBLIC:
					publicMethods = append(publicMethods, fullMethod)
				case sharedv1.AccessLevel_ACCESS_API_KEY:
					apiKeyMethods = append(apiKeyMethods, fullMethod)
				case sharedv1.AccessLevel_ACCESS_PERMISSION:
					permissionMethods[fullMethod] = perms
				}
			}
		}
	}
	return publicMethods, apiKeyMethods, permissionMethods, nil
}

func resolveServiceDefaultAccess(service protoreflect.ServiceDescriptor) sharedv1.AccessLevel {
	options, ok := service.Options().(*descriptorpb.ServiceOptions)
	if !ok || options == nil || !proto.HasExtension(options, sharedv1.E_ServiceAuth) {
		return sharedv1.AccessLevel_ACCESS_LEVEL_UNSPECIFIED
	}
	ext := proto.GetExtension(options, sharedv1.E_ServiceAuth)
	policy, ok := ext.(*sharedv1.ServiceAuth)
	if !ok {
		return sharedv1.AccessLevel_ACCESS_LEVEL_UNSPECIFIED
	}
	return policy.GetDefaultAccess()
}

func resolveMethodAccess(method protoreflect.MethodDescriptor, serviceDefault sharedv1.AccessLevel) (sharedv1.AccessLevel, []string, bool) {
	options, ok := method.Options().(*descriptorpb.MethodOptions)
	if ok && options != nil && proto.HasExtension(options, sharedv1.E_MethodAuth) {
		ext := proto.GetExtension(options, sharedv1.E_MethodAuth)
		policy, ok := ext.(*sharedv1.MethodAuth)
		if ok && policy.GetAccess() != sharedv1.AccessLevel_ACCESS_LEVEL_UNSPECIFIED {
			return policy.GetAccess(), policy.GetPermissions(), true
		}
	}
	if serviceDefault != sharedv1.AccessLevel_ACCESS_LEVEL_UNSPECIFIED {
		return serviceDefault, nil, true
	}
	return sharedv1.AccessLevel_ACCESS_LEVEL_UNSPECIFIED, nil, false
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
