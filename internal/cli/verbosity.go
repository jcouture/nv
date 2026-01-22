package cli

import "github.com/jcouture/nv/internal/config"

func verbosityLevel() int {
	if level := config.GetVerbosityOverride(); level > 0 {
		return level
	}
	return 0
}

func setVerbosityOverride(enabled bool) {
	if enabled {
		config.SetVerbosityOverride(2)
		return
	}
	config.ClearVerbosityOverride()
}
