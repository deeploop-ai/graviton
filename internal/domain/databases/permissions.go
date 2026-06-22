package databases

import (
	"fmt"
	"strings"
)

// SystemRoles bypass document-level _perms checks. Use only for internal
// infrastructure paths (session validation, post-create reads, email lookup).
var SystemRoles = []string{"__system__"}

// DefaultCollectionPermissions returns a reasonable default permission set for
// user-created collections that do not specify explicit permissions.
func DefaultCollectionPermissions() []Permission {
	return []Permission{
		{Type: "create", Role: "users"},
		{Type: "read", Role: "any"},
		{Type: "update", Role: "users"},
		{Type: "delete", Role: "users"},
		{Type: "create", Role: "keys"},
		{Type: "read", Role: "keys"},
		{Type: "update", Role: "keys"},
		{Type: "delete", Role: "keys"},
		{Type: "create", Role: "admin"},
		{Type: "read", Role: "admin"},
		{Type: "update", Role: "admin"},
		{Type: "delete", Role: "admin"},
	}
}

// CollectionAllows reports whether the collection-level permission list grants
// the given operation type to any of the provided roles.
func CollectionAllows(perms []Permission, permType string, roles []string) bool {
	for _, p := range perms {
		if p.Type != permType {
			continue
		}
		for _, r := range roles {
			if p.Role == r {
				return true
			}
		}
	}
	return false
}

// ParsePermissionStrings converts "read:any" style strings into Permission values.
func ParsePermissionStrings(items []string) ([]Permission, error) {
	if len(items) == 0 {
		return DefaultCollectionPermissions(), nil
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
