package testutil

import "testing"

func TestReplaceDatabaseName(t *testing.T) {
	got, err := replaceDatabaseName("postgres://orionid:orionid@127.0.0.1:5433/ORIONID_test?sslmode=disable", "ORIONID_test_1_1")
	if err != nil {
		t.Fatalf("replaceDatabaseName: %v", err)
	}
	want := "postgres://orionid:orionid@127.0.0.1:5433/ORIONID_test_1_1?sslmode=disable"
	if got != want {
		t.Fatalf("replaceDatabaseName() = %q, want %q", got, want)
	}
}

func TestTestDBPrefix(t *testing.T) {
	t.Setenv("ORIONID_TEST_DATABASE_SOURCE", "postgres://orionid:orionid@127.0.0.1:5433/custom_test?sslmode=disable")
	if got := testDBPrefix(); got != "custom_test" {
		t.Fatalf("testDBPrefix() = %q, want custom_test", got)
	}
}
