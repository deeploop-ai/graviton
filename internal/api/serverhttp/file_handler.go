package serverhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/deeploop-ai/graviton/internal/app/storage"
	"github.com/deeploop-ai/graviton/internal/domain/databases"
	"github.com/deeploop-ai/graviton/internal/domain/shared"
	"github.com/deeploop-ai/graviton/internal/infra/auth"
	"github.com/deeploop-ai/graviton/internal/pkg/config"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FileHandler provides HTTP multipart upload/download for storage.
type FileHandler struct {
	cfg       *config.AppConfig
	validator *auth.Validator
	storage   *storage.Storage
}

// NewFileHandler creates a new file HTTP handler.
func NewFileHandler(
	cfg *config.AppConfig,
	validator *auth.Validator,
	storage *storage.Storage,
) *FileHandler {
	return &FileHandler{cfg: cfg, validator: validator, storage: storage}
}

// Register attaches the upload/download routes to the gateway mux.
func (h *FileHandler) Register(mux *runtime.ServeMux) {
	_ = mux.HandlePath("POST", "/v1/storage/buckets/{bucketId}/files", h.upload)
	_ = mux.HandlePath("GET", "/v1/storage/buckets/{bucketId}/files/{fileId}/download", h.download)
	_ = mux.HandlePath("GET", "/v1/storage/buckets/{bucketId}/files/{fileId}/view", h.download)
}

// maxUploadBytes caps the total size of a multipart upload request body.
const maxUploadBytes = 100 << 20 // 100 MiB

func (h *FileHandler) upload(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	ctx := r.Context()
	principal, err := h.authenticate(r)
	if err != nil {
		httpError(w, err)
		return
	}
	projectID := h.projectID(r, principal)
	if projectID == "" {
		httpError(w, status.Error(codes.Unauthenticated, "missing project context"))
		return
	}
	bucketID := pathParams["bucketId"]
	if bucketID == "" {
		httpError(w, status.Error(codes.InvalidArgument, "missing bucket id"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpError(w, status.Error(codes.InvalidArgument, "invalid multipart form or file too large"))
		return
	}
	defer r.MultipartForm.RemoveAll()

	fileHeader := r.MultipartForm.File["file"]
	if len(fileHeader) == 0 {
		file, fh, err := r.FormFile("file")
		if err != nil {
			httpError(w, status.Error(codes.InvalidArgument, "missing file"))
			return
		}
		defer file.Close()
		h.createFile(ctx, w, projectID, bucketID, file, fh.Size, fh.Filename, fh.Header.Get("Content-Type"), principal)
		return
	}

	fh := fileHeader[0]
	f, err := fh.Open()
	if err != nil {
		httpError(w, status.Error(codes.Internal, "cannot open uploaded file"))
		return
	}
	defer f.Close()

	h.createFile(ctx, w, projectID, bucketID, f, fh.Size, fh.Filename, fh.Header.Get("Content-Type"), principal)
}

func (h *FileHandler) createFile(ctx context.Context, w http.ResponseWriter, projectID, bucketID string, r io.Reader, size int64, name, contentType string, principal *shared.Principal) {
	file, err := h.storage.CreateFile(ctx, storage.CreateFileCommand{
		ProjectID:   projectID,
		OwnerUserID: principal.UserID,
		BucketID:    bucketID,
		Name:        name,
		MimeType:    contentType,
	}, r, size, databases.Principal{Roles: principal.Roles, PlatformAdmin: principal.IsPlatformAdmin})
	if err != nil {
		httpError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":         file.ID,
		"bucket_id":  file.BucketID,
		"name":       file.Name,
		"mime_type":  file.MimeType,
		"size":       file.Size,
		"created_at": file.CreatedAt,
		"updated_at": file.UpdatedAt,
	})
}

func (h *FileHandler) download(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	ctx := r.Context()
	principal, err := h.authenticate(r)
	if err != nil {
		httpError(w, err)
		return
	}
	projectID := h.projectID(r, principal)
	if projectID == "" {
		httpError(w, status.Error(codes.Unauthenticated, "missing project context"))
		return
	}
	bucketID := pathParams["bucketId"]
	fileID := pathParams["fileId"]
	if bucketID == "" || fileID == "" {
		httpError(w, status.Error(codes.InvalidArgument, "missing bucket or file id"))
		return
	}

	file, reader, err := h.storage.GetFile(ctx, projectID, bucketID, fileID, databases.Principal{Roles: principal.Roles, PlatformAdmin: principal.IsPlatformAdmin})
	if err != nil {
		httpError(w, err)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", file.MimeType)
	disposition := "attachment"
	if !strings.HasSuffix(r.URL.Path, "/download") {
		disposition = "inline"
	}
	w.Header().Set("Content-Disposition", contentDispositionHeader(disposition, file.Name))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, reader)
}

func (h *FileHandler) authenticate(r *http.Request) (*shared.Principal, error) {
	ctx := r.Context()
	if key := r.Header.Get("X-Api-Key"); key != "" {
		return h.validator.ValidateCredential(ctx, key, shared.CredentialTypeAPIKey)
	}
	if authz := r.Header.Get("Authorization"); authz != "" {
		token := strings.TrimPrefix(authz, "Bearer ")
		return h.validator.ValidateToken(ctx, token)
	}
	for _, c := range r.Cookies() {
		if strings.HasPrefix(c.Name, "GRAVITON_session_") {
			return h.validator.ValidateCredential(ctx, c.Value, shared.CredentialTypeSession)
		}
	}
	return nil, status.Error(codes.Unauthenticated, "authentication credential is not provided")
}

func (h *FileHandler) projectID(r *http.Request, p *shared.Principal) string {
	if p == nil {
		return ""
	}
	switch p.CredentialType {
	case shared.CredentialTypeAPIKey:
		return p.ProjectID
	case shared.CredentialTypeToken, shared.CredentialTypeSession:
		if p.ActorKind == shared.ActorKindAdmin {
			if pid := strings.TrimSpace(r.Header.Get("X-Graviton-Project")); pid != "" {
				return pid
			}
		}
		return p.ProjectID
	default:
		return p.ProjectID
	}
}

func safeFilename(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r == '"', r == '\\':
			b.WriteByte('_')
		case r < 32, r == 127:
			// drop control characters to prevent header injection
			continue
		case r == '\n', r == '\r':
			continue
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "download"
	}
	return out
}

func contentDispositionHeader(disposition, name string) string {
	safe := safeFilename(name)
	ascii := asciiFilenameFallback(safe)
	encoded := url.PathEscape(safe)
	return fmt.Sprintf(`%s; filename="%s"; filename*=UTF-8''%s`, disposition, ascii, encoded)
}

func asciiFilenameFallback(name string) string {
	var b strings.Builder
	for _, r := range name {
		if r >= 32 && r <= 126 && r != '"' && r != '\\' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "._ ")
	if out == "" {
		return "download"
	}
	return out
}

func httpError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		st = status.New(codes.Internal, err.Error())
	}
	httpStatus := runtime.HTTPStatusFromCode(st.Code())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	payload, _ := json.Marshal(map[string]any{
		"error": map[string]string{
			"type":    st.Code().String(),
			"message": st.Message(),
		},
	})
	_, _ = w.Write(payload)
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
