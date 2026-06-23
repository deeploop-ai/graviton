package databases

import "errors"

// ErrPermissionDenied is returned when the caller lacks document-level permission.
var ErrPermissionDenied = errors.New("permission denied")

// SimpleDocumentUpdate builds a DocumentUpdate for data and optional permission changes.
func SimpleDocumentUpdate(doc Document, perms []Permission) DocumentUpdate {
	return DocumentUpdate{Document: doc, Permissions: perms}
}
