package cli

import "testing"

// Deprecated legacy binary tests removed; ensure main CLI still responds.
func TestRootCommandName(t *testing.T) {
	cmd := NewRootCmd("")
	if cmd.Use != "nv" {
		t.Fatalf("expected nv, got %s", cmd.Use)
	}
}
