// Copyright 2015-2026 Jean-Philippe Couture
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
)

func GetConfigDir() (string, error) {
	return platformConfigDir()
}

func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func EnsureConfigDir() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o750)
}

func ConfigExists() (bool, error) {
	path, err := GetConfigPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetLegacyEnvPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".nv"), nil
}

func LegacyEnvExists() (bool, error) {
	path, err := GetLegacyEnvPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ExpandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if path == "~" {
		return home, nil
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}

func Load() *Config {
	cfg, _, err := LoadWithMigration()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return Default()
	}
	return cfg
}

func LoadWithMigration() (*Config, bool, error) {
	configExists, err := ConfigExists()
	if err != nil {
		return Default(), false, err
	}
	if configExists {
		path, err := GetConfigPath()
		if err != nil {
			return Default(), false, err
		}
		cfg, err := LoadFromPath(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse config file, using defaults: %v\n", err)
			fmt.Fprintln(os.Stderr, "To fix: nv config validate")
			return Default(), false, nil
		}
		return cfg, false, nil
	}

	legacyExists, err := LegacyEnvExists()
	if err != nil {
		return Default(), false, err
	}
	if legacyExists {
		shouldMigrate, err := PromptMigration()
		if err != nil {
			return Default(), false, err
		}
		if shouldMigrate {
			migrated, err := MigrateLegacyEnv()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
				fmt.Fprintln(os.Stderr, "Your ~/.nv file is unchanged. You can:")
				fmt.Fprintln(os.Stderr, "  - Try again: nv config migrate")
				fmt.Fprintln(os.Stderr, "  - Manually edit config: nv config edit")
				fmt.Fprintln(os.Stderr, "  - Continue using ~/.nv (not recommended)")
				return Default(), false, nil
			}
			path, err := GetConfigPath()
			if err != nil {
				return Default(), migrated, err
			}
			cfg, err := LoadFromPath(path)
			if err != nil {
				return Default(), migrated, err
			}
			return cfg, migrated, nil
		}

		cfg := Default()
		if err := cfg.Save(); err != nil {
			return cfg, false, err
		}
		fmt.Fprintln(os.Stderr, "Migration skipped. Your ~/.nv file remains unchanged.")
		fmt.Fprintln(os.Stderr, "To migrate later, run: nv config migrate")
		fmt.Fprintln(os.Stderr, "To use nv without ~/.nv, delete it manually.")
		return cfg, false, nil
	}

	cfg := Default()
	if err := cfg.Save(); err != nil {
		return cfg, false, err
	}
	return cfg, false, nil
}

func LoadFromPath(path string) (*Config, error) {
	cfg := Default()
	tree, err := toml.LoadFile(path)
	if err != nil {
		return nil, err
	}
	if err := tree.Unmarshal(cfg); err != nil {
		return nil, err
	}
	cfg.defined = definedFromTree(tree)
	cfg.MergeWithDefaults()

	if errs := cfg.Validate(); len(errs) > 0 {
		fixed, fixedFields := cfg.Fix()
		for _, field := range fixedFields {
			fmt.Fprintf(os.Stderr, "Config value reset to default: %s\n", field)
		}
		cfg = fixed
	}

	return cfg, nil
}

func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	return c.SaveToPath(path)
}

func (c *Config) SaveToPath(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304 -- path is controlled by config location or tests.
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied accessing config: %w\nFix with: chmod 600 %s", err, path)
		}
		return err
	}
	defer file.Close()
	if err := os.Chmod(path, 0o600); err != nil {
		return err
	}

	if c.Globals.Env == nil {
		c.Globals.Env = map[string]string{}
	}

	encoder := toml.NewEncoder(file)
	return encoder.Encode(c)
}

func (c *Config) MergeWithDefaults() {
	defaults := Default()
	if c.defined == nil {
		return
	}

	if !c.defined["general.verbosity"] {
		c.General.Verbosity = defaults.General.Verbosity
	}

	if !c.defined["defaults.env_file"] {
		c.Defaults.EnvFile = defaults.Defaults.EnvFile
	}
	if !c.defined["defaults.auto_local"] {
		c.Defaults.AutoLocal = defaults.Defaults.AutoLocal
	}
	if !c.defined["defaults.cascade"] {
		c.Defaults.Cascade = defaults.Defaults.Cascade
	}
	if !c.defined["defaults.dry_run"] {
		c.Defaults.DryRun = defaults.Defaults.DryRun
	}

	if !c.defined["validation.enabled"] {
		c.Validation.Enabled = defaults.Validation.Enabled
	}
	if !c.defined["validation.schema_file"] {
		c.Validation.SchemaFile = defaults.Validation.SchemaFile
	}
	if !c.defined["validation.strict"] {
		c.Validation.Strict = defaults.Validation.Strict
	}

	if !c.defined["globals.priority"] {
		c.Globals.Priority = defaults.Globals.Priority
	}
	if !c.defined["globals.env"] {
		c.Globals.Env = defaults.Globals.Env
	}
}

func definedFromTree(tree *toml.Tree) map[string]bool {
	defined := make(map[string]bool)
	keys := []struct {
		path []string
		key  string
	}{
		{[]string{"general", "verbosity"}, "general.verbosity"},
		{[]string{"defaults", "env_file"}, "defaults.env_file"},
		{[]string{"defaults", "auto_local"}, "defaults.auto_local"},
		{[]string{"defaults", "cascade"}, "defaults.cascade"},
		{[]string{"defaults", "dry_run"}, "defaults.dry_run"},
		{[]string{"validation", "enabled"}, "validation.enabled"},
		{[]string{"validation", "schema_file"}, "validation.schema_file"},
		{[]string{"validation", "strict"}, "validation.strict"},
		{[]string{"globals", "priority"}, "globals.priority"},
		{[]string{"globals", "env"}, "globals.env"},
	}

	for _, item := range keys {
		if tree.HasPath(item.path) {
			defined[item.key] = true
		}
	}
	return defined
}
