package databases

import "testing"

func TestExpandPermissionRoles_GuestsDoNotGetUsers(t *testing.T) {
	expanded := ExpandPermissionRoles([]string{"guests"})
	if contains(expanded, "users") {
		t.Fatal("guests should not implicitly receive users role")
	}
	if !contains(expanded, "any") {
		t.Fatal("any should always be present")
	}
}

func TestExpandPermissionRoles_AuthenticatedGetsUsers(t *testing.T) {
	expanded := ExpandPermissionRoles([]string{"users", "user:alice"})
	if !contains(expanded, "users") {
		t.Fatal("authenticated caller should have users role")
	}
}

func TestCollectionAllows_WriteGrantsUpdate(t *testing.T) {
	perms := []Permission{{Type: "write", Role: "users"}}
	if !CollectionAllows(perms, "update", []string{"users"}) {
		t.Fatal("write should grant update")
	}
	if CollectionAllows(perms, "read", []string{"users"}) {
		t.Fatal("write should not grant read")
	}
}

func TestAllowsDocumentAccess_DocumentSecurityOR(t *testing.T) {
	coll := &Collection{
		DocumentSecurity: true,
		Permissions:    []Permission{{Type: "read", Role: "any"}},
	}
	docPerms := []Permission{{Type: "read", Role: "user:bob"}}
	if !AllowsDocumentAccess(coll, docPerms, true, "read", []string{"user:alice"}) {
		t.Fatal("collection read:any should allow alice")
	}
	if !AllowsDocumentAccess(coll, docPerms, true, "read", []string{"user:carol"}) {
		t.Fatal("collection read:any should allow carol via OR")
	}
	if !AllowsDocumentAccess(coll, docPerms, true, "read", []string{"user:bob"}) {
		t.Fatal("document read:user:bob should allow bob")
	}

	locked := &Collection{
		DocumentSecurity: true,
		Permissions:    []Permission{{Type: "create", Role: "users"}},
	}
	if AllowsDocumentAccess(locked, docPerms, true, "read", []string{"user:carol"}) {
		t.Fatal("carol should not have access without collection or document read")
	}
}

func TestAllowsDocumentAccess_DocumentSecurityOffIgnoresDocPerms(t *testing.T) {
	coll := &Collection{
		DocumentSecurity: false,
		Permissions:    []Permission{{Type: "read", Role: "any"}},
	}
	docPerms := []Permission{{Type: "read", Role: "user:bob"}}
	if !AllowsDocumentAccess(coll, docPerms, true, "read", []string{"user:carol"}) {
		t.Fatal("document perms ignored when documentSecurity=false")
	}
}

func TestParsePermissionStrings_WriteExpands(t *testing.T) {
	perms, err := ParsePermissionStrings([]string{"write:users"})
	if err != nil {
		t.Fatal(err)
	}
	if len(perms) != 3 {
		t.Fatalf("expected 3 perms, got %d", len(perms))
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
