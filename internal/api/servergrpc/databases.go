package servergrpc

import (
	"context"

	serverv1 "github.com/deeploop-ai/fleet/genproto/server/v1"
	sharedv1 "github.com/deeploop-ai/fleet/genproto/shared/v1"
	appserver "github.com/deeploop-ai/fleet/internal/app/server"
	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/pkg/contexts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DatabasesService struct {
	serverv1.UnimplementedDatabasesServiceServer
	databases *appserver.Databases
}

func NewDatabasesService(databases *appserver.Databases) *DatabasesService {
	return &DatabasesService{databases: databases}
}

func (s *DatabasesService) projectID(ctx context.Context) string {
	p, ok := contexts.Principal(ctx)
	if !ok {
		return ""
	}
	return p.ProjectID
}

func (s *DatabasesService) CreateDatabase(ctx context.Context, req *serverv1.CreateDatabaseRequest) (*serverv1.Database, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	id := req.GetId()
	if id == "" {
		id = "default"
	}
	if err := s.databases.CreateDatabase(ctx, projectID, id, req.GetName()); err != nil {
		return nil, err
	}
	return &serverv1.Database{Id: id, Name: req.GetName()}, nil
}

func (s *DatabasesService) ListDatabases(ctx context.Context, _ *sharedv1.ListRequest) (*serverv1.ListDatabasesResponse, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	cols, err := s.databases.ListDatabases(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]*serverv1.Database, len(cols))
	for i := range cols {
		out[i] = mapDatabase(&cols[i])
	}
	return &serverv1.ListDatabasesResponse{Databases: out, Meta: &sharedv1.ListResponseMeta{}}, nil
}

func (s *DatabasesService) GetDatabase(ctx context.Context, req *serverv1.GetDatabaseRequest) (*serverv1.Database, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	col, err := s.databases.GetDatabase(ctx, projectID, req.GetId())
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, status.Error(codes.NotFound, "database not found")
	}
	return mapDatabase(col), nil
}

func (s *DatabasesService) DeleteDatabase(ctx context.Context, req *serverv1.GetDatabaseRequest) (*sharedv1.Empty, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	if err := s.databases.DeleteDatabase(ctx, projectID, req.GetId()); err != nil {
		return nil, err
	}
	return &sharedv1.Empty{}, nil
}

func (s *DatabasesService) CreateCollection(ctx context.Context, req *serverv1.CreateCollectionRequest) (*serverv1.Collection, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	if err := s.databases.CreateCollection(ctx, projectID, req.GetDatabaseId(), req.GetId(), req.GetName(), nil, nil); err != nil {
		return nil, err
	}
	return &serverv1.Collection{
		Id:         req.GetId(),
		DatabaseId: req.GetDatabaseId(),
		Name:       req.GetName(),
	}, nil
}

func (s *DatabasesService) ListCollections(ctx context.Context, req *serverv1.ListCollectionsRequest) (*serverv1.ListCollectionsResponse, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	cols, err := s.databases.ListCollections(ctx, projectID, req.GetDatabaseId())
	if err != nil {
		return nil, err
	}
	out := make([]*serverv1.Collection, len(cols))
	for i := range cols {
		out[i] = mapCollection(&cols[i])
	}
	return &serverv1.ListCollectionsResponse{Collections: out, Meta: &sharedv1.ListResponseMeta{}}, nil
}

func (s *DatabasesService) GetCollection(ctx context.Context, req *serverv1.GetCollectionRequest) (*serverv1.Collection, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	col, err := s.databases.GetCollection(ctx, projectID, req.GetDatabaseId(), req.GetCollectionId())
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, status.Error(codes.NotFound, "collection not found")
	}
	return mapCollection(col), nil
}

func (s *DatabasesService) DeleteCollection(ctx context.Context, req *serverv1.GetCollectionRequest) (*sharedv1.Empty, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	if err := s.databases.DeleteCollection(ctx, projectID, req.GetDatabaseId(), req.GetCollectionId()); err != nil {
		return nil, err
	}
	return &sharedv1.Empty{}, nil
}

func (s *DatabasesService) CreateAttribute(ctx context.Context, req *serverv1.CreateAttributeRequest) (*serverv1.Attribute, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	attr := databases.Attribute{
		ID:       req.GetKey(),
		Key:      req.GetKey(),
		Type:     s.databases.MapAttributeType(req.GetType()),
		Size:     int(req.GetSize()),
		Required: req.GetRequired(),
		Array:    req.GetArray(),
	}
	if req.GetDefaultValue() != "" {
		attr.Default = req.GetDefaultValue()
	}
	if err := s.databases.CreateAttribute(ctx, projectID, req.GetDatabaseId(), req.GetCollectionId(), attr); err != nil {
		return nil, err
	}
	return &serverv1.Attribute{
		Id:           attr.ID,
		Key:          attr.Key,
		Type:         attr.Type,
		Size:         int32(attr.Size),
		Required:     attr.Required,
		Array:        attr.Array,
		DefaultValue: req.GetDefaultValue(),
	}, nil
}

func (s *DatabasesService) CreateIndex(ctx context.Context, req *serverv1.CreateIndexRequest) (*serverv1.Index, error) {
	projectID := s.projectID(ctx)
	if projectID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing project context")
	}
	idx := databases.Index{
		ID:         req.GetId(),
		Type:       req.GetType(),
		Attributes: req.GetAttributes(),
		Orders:     req.GetOrders(),
	}
	if err := s.databases.CreateIndex(ctx, projectID, req.GetDatabaseId(), req.GetCollectionId(), idx); err != nil {
		return nil, err
	}
	return &serverv1.Index{
		Id:         idx.ID,
		Type:       idx.Type,
		Attributes: idx.Attributes,
		Orders:     idx.Orders,
	}, nil
}

func mapDatabase(c *databases.Collection) *serverv1.Database {
	if c == nil {
		return nil
	}
	return &serverv1.Database{
		Id:        c.ID,
		Name:      c.Name,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func mapCollection(c *databases.Collection) *serverv1.Collection {
	if c == nil {
		return nil
	}
	out := &serverv1.Collection{
		Id:         c.ID,
		DatabaseId: c.DatabaseID,
		Name:       c.Name,
		CreatedAt:  timestamppb.New(c.CreatedAt),
		UpdatedAt:  timestamppb.New(c.UpdatedAt),
	}
	for _, p := range c.Permissions {
		out.Permissions = append(out.Permissions, p.Type+":"+p.Role)
	}
	for _, a := range c.Attributes {
		out.Attributes = append(out.Attributes, &serverv1.Attribute{
			Id:       a.ID,
			Key:      a.Key,
			Type:     a.Type,
			Size:     int32(a.Size),
			Required: a.Required,
			Array:    a.Array,
		})
	}
	for _, i := range c.Indexes {
		out.Indexes = append(out.Indexes, &serverv1.Index{
			Id:         i.ID,
			Type:       i.Type,
			Attributes: i.Attributes,
			Orders:     i.Orders,
		})
	}
	return out
}
