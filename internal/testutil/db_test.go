package testutil

import "testing"

func TestReplaceDatabaseName(t *testing.T) {
	got, err := replaceDatabaseName("postgres://fleet:fleet@127.0.0.1:5433/fleet_test?sslmode=disable", "fleet_test_1_1")
	if err != nil {
		t.Fatalf("replaceDatabaseName: %v", err)
	}
	want := "postgres://fleet:fleet@127.0.0.1:5433/fleet_test_1_1?sslmode=disable"
	if got != want {
		t.Fatalf("replaceDatabaseName() = %q, want %q", got, want)
	}
}

func TestTestDBPrefix(t *testing.T) {
	t.Setenv("FLEET_TEST_DATABASE_SOURCE", "postgres://fleet:fleet@127.0.0.1:5433/custom_test?sslmode=disable")
	if got := testDBPrefix(); got != "custom_test" {
		t.Fatalf("testDBPrefix() = %q, want custom_test", got)
	}
}
