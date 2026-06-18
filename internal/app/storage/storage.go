package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/domain/projects"
	"github.com/deeploop-ai/fleet/internal/domain/storage"
	"github.com/deeploop-ai/fleet/internal/pkg/config"
	"github.com/deeploop-ai/fleet/pkg/idgen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Storage struct {
	cfg         *config.AppConfig
	projectRepo projects.Repository
	docDB       databases.DocumentDB
	store       storage.ObjectStore
}

func NewStorage(
	cfg *config.AppConfig,
	projectRepo projects.Repository,
	docDB databases.DocumentDB,
	store storage.ObjectStore,
) *Storage {
	return &Storage{cfg: cfg, projectRepo: projectRepo, docDB: docDB, store: store}
}

type CreateBucketCommand struct {
	ProjectID   string
	Name        string
	Permissions []string
}

type CreateFileCommand struct {
	ProjectID   string
	BucketID    string
	Name        string
	MimeType    string
	Metadata    map[string]string
	Permissions []string
}

func (s *Storage) CreateBucket(ctx context.Context, cmd CreateBucketCommand) (*storage.Bucket, error) {
	project, err := s.resolveProject(ctx, cmd.ProjectID)
	if err != nil {
		return nil, err
	}
	if err := s.docDB.EnsureSystemCollections(ctx, project.ID, project.InternalID); err != nil {
		return nil, err
	}

	bucketID := idgen.UUID().String()
	now := time.Now()
	bucketDoc := databases.Document{
		ID: bucketID,
		Data: map[string]any{
			"name":        cmd.Name,
			"permissions": cmd.Permissions,
		},
	}
	perms := bucketPermissions(bucketID, cmd.Permissions)
	if _, err := s.docDB.CreateDocument(ctx, project.ID, "default", "buckets", bucketDoc, perms); err != nil {
		return nil, fmt.Errorf("create bucket document: %w", err)
	}

	return &storage.Bucket{
		ID:          bucketID,
		ProjectID:   project.ID,
		Name:        cmd.Name,
		Permissions: cmd.Permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Storage) ListBuckets(ctx context.Context, projectID string, q databases.Query, roles []string) ([]storage.Bucket, int64, error) {
	project, err := s.resolveProject(ctx, projectID)
	if err != nil {
		return nil, 0, err
	}
	if err := s.docDB.EnsureSystemCollections(ctx, project.ID, project.InternalID); err != nil {
		return nil, 0, err
	}

	list, err := s.docDB.ListDocuments(ctx, project.ID, "default", "buckets", q, roles)
	if err != nil {
		return nil, 0, err
	}
	buckets := make([]storage.Bucket, 0, len(list.Documents))
	for _, d := range list.Documents {
		buckets = append(buckets, *mapBucketDoc(&d))
	}
	return buckets, list.TotalCount, nil
}

func (s *Storage) DeleteBucket(ctx context.Context, projectID, bucketID string) error {
	project, err := s.resolveProject(ctx, projectID)
	if err != nil {
		return err
	}
	// Delete all file objects in this bucket.
	files, _, _, err := s.ListFiles(ctx, projectID, bucketID, databases.Query{PageSize: 1000}, []string{"admin"})
	if err != nil {
		return err
	}
	for _, f := range files {
		_ = s.store.Delete(ctx, defaultBucketName(s.cfg), objectKey(projectID, bucketID, f.ID))
	}
	return s.docDB.DeleteDocument(ctx, project.ID, "default", "buckets", bucketID)
}

func (s *Storage) CreateFile(ctx context.Context, cmd CreateFileCommand, content io.Reader, size int64) (*storage.File, error) {
	project, err := s.resolveProject(ctx, cmd.ProjectID)
	if err != nil {
		return nil, err
	}
	if err := s.docDB.EnsureSystemCollections(ctx, project.ID, project.InternalID); err != nil {
		return nil, err
	}

	// Verify bucket exists.
	if _, err := s.docDB.GetDocument(ctx, project.ID, "default", "buckets", cmd.BucketID); err != nil {
		return nil, status.Error(codes.NotFound, "bucket not found")
	}

	fileID := idgen.UUID().String()
	now := time.Now()
	fileDoc := databases.Document{
		ID: fileID,
		Data: map[string]any{
			"bucket_id": cmd.BucketID,
			"name":      cmd.Name,
			"mime_type": cmd.MimeType,
			"size":      size,
			"metadata":  cmd.Metadata,
		},
	}
	perms := filePermissions(fileID, cmd.Permissions)
	if _, err := s.docDB.CreateDocument(ctx, project.ID, "default", "files", fileDoc, perms); err != nil {
		return nil, fmt.Errorf("create file document: %w", err)
	}

	if err := s.store.EnsureBucket(ctx, defaultBucketName(s.cfg)); err != nil {
		return nil, fmt.Errorf("ensure storage bucket: %w", err)
	}
	if err := s.store.Put(ctx, defaultBucketName(s.cfg), objectKey(project.ID, cmd.BucketID, fileID), content, size, cmd.MimeType); err != nil {
		// Attempt rollback metadata.
		_ = s.docDB.DeleteDocument(ctx, project.ID, "default", "files", fileID)
		return nil, fmt.Errorf("upload file: %w", err)
	}

	return &storage.File{
		ID:        fileID,
		ProjectID: project.ID,
		BucketID:  cmd.BucketID,
		Name:      cmd.Name,
		MimeType:  cmd.MimeType,
		Size:      size,
		Metadata:  cmd.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Storage) GetFile(ctx context.Context, projectID, bucketID, fileID string) (*storage.File, io.ReadCloser, error) {
	project, err := s.resolveProject(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}
	doc, err := s.docDB.GetDocument(ctx, project.ID, "default", "files", fileID)
	if err != nil {
		return nil, nil, err
	}
	if doc == nil {
		return nil, nil, status.Error(codes.NotFound, "file not found")
	}
	file := mapFileDoc(doc)
	if file.BucketID != bucketID {
		return nil, nil, status.Error(codes.NotFound, "file not found in bucket")
	}
	reader, err := s.store.Get(ctx, defaultBucketName(s.cfg), objectKey(project.ID, bucketID, fileID))
	if err != nil {
		return file, nil, err
	}
	return file, reader, nil
}

func (s *Storage) DeleteFile(ctx context.Context, projectID, bucketID, fileID string) error {
	project, err := s.resolveProject(ctx, projectID)
	if err != nil {
		return err
	}
	if err := s.store.Delete(ctx, defaultBucketName(s.cfg), objectKey(project.ID, bucketID, fileID)); err != nil {
		// Continue to delete metadata even if object missing.
	}
	return s.docDB.DeleteDocument(ctx, project.ID, "default", "files", fileID)
}

func (s *Storage) ListFiles(ctx context.Context, projectID, bucketID string, q databases.Query, roles []string) ([]storage.File, int64, string, error) {
	project, err := s.resolveProject(ctx, projectID)
	if err != nil {
		return nil, 0, "", err
	}
	if err := s.docDB.EnsureSystemCollections(ctx, project.ID, project.InternalID); err != nil {
		return nil, 0, "", err
	}

	list, err := s.docDB.ListDocuments(ctx, project.ID, "default", "files", q, roles)
	if err != nil {
		return nil, 0, "", err
	}
	files := make([]storage.File, 0, len(list.Documents))
	for _, d := range list.Documents {
		files = append(files, *mapFileDoc(&d))
	}
	return files, list.TotalCount, list.NextPageToken, nil
}

func (s *Storage) resolveProject(ctx context.Context, projectID string) (*projects.Project, error) {
	p, err := s.projectRepo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, status.Error(codes.NotFound, "project not found")
	}
	return p, nil
}

func defaultBucketName(cfg *config.AppConfig) string {
	b := cfg.GetStorage().GetS3().GetBucket()
	if b == "" {
		return "fleet-files"
	}
	return b
}

func objectKey(projectID, bucketID, fileID string) string {
	return fmt.Sprintf("%s/%s/%s", projectID, bucketID, fileID)
}

func bucketPermissions(bucketID string, explicit []string) []databases.Permission {
	if len(explicit) > 0 {
		return parseRawPermissions(explicit)
	}
	return []databases.Permission{
		{Type: "read", Role: "any"},
		{Type: "create", Role: "users"},
		{Type: "update", Role: "users"},
		{Type: "delete", Role: "users"},
	}
}

func filePermissions(fileID string, explicit []string) []databases.Permission {
	if len(explicit) > 0 {
		return parseRawPermissions(explicit)
	}
	return []databases.Permission{
		{Type: "read", Role: "any"},
		{Type: "update", Role: fmt.Sprintf("user:%s", fileID)},
		{Type: "delete", Role: fmt.Sprintf("user:%s", fileID)},
	}
}

func parseRawPermissions(raw []string) []databases.Permission {
	var perms []databases.Permission
	for _, r := range raw {
		parts := strings.SplitN(r, ":", 2)
		if len(parts) == 2 {
			perms = append(perms, databases.Permission{Type: parts[0], Role: parts[1]})
		}
	}
	return perms
}

func mapBucketDoc(doc *databases.Document) *storage.Bucket {
	b := &storage.Bucket{
		ID:        doc.ID,
		ProjectID: "",
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
	if v, ok := doc.Data["name"].(string); ok {
		b.Name = v
	}
	if arr, ok := doc.Data["permissions"].([]any); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				b.Permissions = append(b.Permissions, s)
			}
		}
	}
	return b
}

func mapFileDoc(doc *databases.Document) *storage.File {
	f := &storage.File{
		ID:        doc.ID,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
		Metadata:  map[string]string{},
	}
	if v, ok := doc.Data["bucket_id"].(string); ok {
		f.BucketID = v
	}
	if v, ok := doc.Data["name"].(string); ok {
		f.Name = v
	}
	if v, ok := doc.Data["mime_type"].(string); ok {
		f.MimeType = v
	}
	if v, ok := doc.Data["size"].(float64); ok {
		f.Size = int64(v)
	}
	if v, ok := doc.Data["size"].(int64); ok {
		f.Size = v
	}
	if m, ok := doc.Data["metadata"].(map[string]any); ok {
		for k, v := range m {
			if s, ok := v.(string); ok {
				f.Metadata[k] = s
			}
		}
	}
	return f
}
