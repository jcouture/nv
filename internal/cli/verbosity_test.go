package cli

import (
	"testing"

	"github.com/jcouture/nv/internal/config"
)

func TestVerbosityLevelUsesOverride(t *testing.T) {
	t.Cleanup(config.ClearVerbosityOverride)

	config.SetVerbosityOverride(5)
	if got := verbosityLevel(); got != 5 {
		t.Fatalf("verbosityLevel=%d want 5", got)
	}
}

func TestVerbosityLevelDefault(t *testing.T) {
	config.ClearVerbosityOverride()
	if got := verbosityLevel(); got != 0 {
		t.Fatalf("verbosityLevel=%d want 0", got)
	}
}

func TestSetVerbosityOverride(t *testing.T) {
	t.Cleanup(config.ClearVerbosityOverride)

	setVerbosityOverride(true)
	if got := config.GetVerbosityOverride(); got != 2 {
		t.Fatalf("override=%d want 2 when enabled", got)
	}

	setVerbosityOverride(false)
	if got := config.GetVerbosityOverride(); got != 0 {
		t.Fatalf("override=%d want cleared when disabled", got)
	}
}
