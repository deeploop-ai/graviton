package databases

// SystemRoles bypass document-level _perms checks. Use only for internal
// infrastructure paths (session validation, post-create reads, email lookup).
var SystemRoles = []string{"__system__"}
