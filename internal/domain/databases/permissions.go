package databases

import (
	"fmt"
	"strings"
)

// SystemRoles bypass document-level _perms checks. Use only for internal
// infrastructure paths (session validation, post-create reads, email lookup).
var SystemRoles = []string{"__system__"}

// DefaultDocumentPermissions grants open access for server-managed documents.
func DefaultDocumentPermissions() []Permission {
	return []Permission{
		{Type: "read", Role: "any"},
		{Type: "create", Role: "any"},
		{Type: "update", Role: "any"},
		{Type: "delete", Role: "any"},
		{Type: "read", Role: "keys"},
		{Type: "update", Role: "keys"},
		{Type: "delete", Role: "keys"},
	}
}

// ParsePermissionStrings converts "read:any" style strings into Permission values.
func ParsePermissionStrings(items []string) ([]Permission, error) {
	if len(items) == 0 {
		return DefaultDocumentPermissions(), nil
	}
	out := make([]Permission, 0, len(items))
	for _, item := range items {
		typ, role, ok := strings.Cut(strings.TrimSpace(item), ":")
		if !ok || typ == "" || role == "" {
			return nil, fmt.Errorf("invalid permission %q (expected type:role)", item)
		}
		out = append(out, Permission{Type: typ, Role: role})
	}
	return out, nil
}
