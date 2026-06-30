package interceptor

import "testing"

func TestAPIKeyScopeAllowed(t *testing.T) {
	t.Parallel()
	method := "/orionid.server.v1.UsersService/ListUsers"

	if apiKeyScopeAllowed(method, nil) {
		t.Fatal("empty scopes should deny resource-scoped methods")
	}
	if !apiKeyScopeAllowed(method, []string{"*"}) {
		t.Fatal("wildcard scope should allow")
	}
	if !apiKeyScopeAllowed(method, []string{"users"}) {
		t.Fatal("matching resource scope should allow")
	}
	if !apiKeyScopeAllowed(method, []string{"users.read"}) {
		t.Fatal("prefixed resource scope should allow")
	}
	if apiKeyScopeAllowed(method, []string{"storage"}) {
		t.Fatal("unrelated scope should deny")
	}
}
