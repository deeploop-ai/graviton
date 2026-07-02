package interceptor

import "testing"

func TestAPIKeyScopeAllowed(t *testing.T) {
	t.Parallel()
	method := "/graviton.server.v1.UsersService/ListUsers"

	if APIKeyScopeAllowed(method, nil) {
		t.Fatal("empty scopes should deny resource-scoped methods")
	}
	if !APIKeyScopeAllowed(method, []string{"*"}) {
		t.Fatal("wildcard scope should allow")
	}
	if !APIKeyScopeAllowed(method, []string{"users"}) {
		t.Fatal("matching resource scope should allow")
	}
	if !APIKeyScopeAllowed(method, []string{"users.read"}) {
		t.Fatal("prefixed resource scope should allow")
	}
	if APIKeyScopeAllowed(method, []string{"storage"}) {
		t.Fatal("unrelated scope should deny")
	}
}
