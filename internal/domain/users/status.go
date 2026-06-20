package users

import "fmt"

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusBlocked  = "blocked"
)

var validStatuses = map[string]struct{}{
	StatusActive:   {},
	StatusInactive: {},
	StatusBlocked:  {},
}

// ValidateStatus reports whether s is an allowed user status value.
func ValidateStatus(s string) error {
	if _, ok := validStatuses[s]; !ok {
		return fmt.Errorf("invalid user status %q (allowed: active, inactive, blocked)", s)
	}
	return nil
}

// CanAuthenticate reports whether a user with the given status may sign in or use tokens.
// Empty status is treated as active (collection default).
func CanAuthenticate(s string) bool {
	if s == "" {
		return true
	}
	return s == StatusActive
}
