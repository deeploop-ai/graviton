package testutil

import "testing"

func TestReplaceDatabaseName(t *testing.T) {
	got, err := replaceDatabaseName("postgres://graviton:graviton@127.0.0.1:5433/GRAVITON_test?sslmode=disable", "GRAVITON_test_1_1")
	if err != nil {
		t.Fatalf("replaceDatabaseName: %v", err)
	}
	want := "postgres://graviton:graviton@127.0.0.1:5433/GRAVITON_test_1_1?sslmode=disable"
	if got != want {
		t.Fatalf("replaceDatabaseName() = %q, want %q", got, want)
	}
}

func TestTestDBPrefix(t *testing.T) {
	t.Setenv("GRAVITON_TEST_DATABASE_SOURCE", "postgres://graviton:graviton@127.0.0.1:5433/custom_test?sslmode=disable")
	if got := testDBPrefix(); got != "custom_test" {
		t.Fatalf("testDBPrefix() = %q, want custom_test", got)
	}
}
