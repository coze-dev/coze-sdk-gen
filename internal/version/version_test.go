package version

import "testing"

func TestString(t *testing.T) {
	if got := String(); got == "" {
		t.Fatal("expected non-empty version string")
	}
}
