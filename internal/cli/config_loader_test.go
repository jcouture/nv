package cli

import (
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/jcouture/nv/internal/config"
)

func TestLoadConfigForWriteCreatesDefaultWhenMissing(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", temp)
	xdg.Reload()

	cfg, err := loadConfigForWrite()
	if err != nil {
		t.Fatalf("loadConfigForWrite: %v", err)
	}
	if cfg.Defaults.EnvFile != config.Default().Defaults.EnvFile {
		t.Fatalf("env_file=%s want %s", cfg.Defaults.EnvFile, config.Default().Defaults.EnvFile)
	}
}

func TestLoadConfigForWriteLoadsExisting(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", temp)
	xdg.Reload()

	configPath := filepath.Join(temp, "nv", "config.toml")
	cfg := config.Default()
	cfg.Defaults.EnvFile = ".env.custom"
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("SaveToPath: %v", err)
	}

	loaded, err := loadConfigForWrite()
	if err != nil {
		t.Fatalf("loadConfigForWrite: %v", err)
	}
	if loaded.Defaults.EnvFile != ".env.custom" {
		t.Fatalf("env_file=%s want .env.custom", loaded.Defaults.EnvFile)
	}
}
