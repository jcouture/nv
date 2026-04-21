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

package cli

import (
	"github.com/jcouture/nv/internal/config"
	"path/filepath"
	"testing"
)

func TestLoadConfigForWriteCreatesDefaultWhenMissing(t *testing.T) {
	_ = setTestConfigHome(t)

	cfg, err := loadConfigForWrite()
	if err != nil {
		t.Fatalf("loadConfigForWrite: %v", err)
	}
	if cfg.Defaults.EnvFile != config.Default().Defaults.EnvFile {
		t.Fatalf("env_file=%s want %s", cfg.Defaults.EnvFile, config.Default().Defaults.EnvFile)
	}
}

func TestLoadConfigForWriteLoadsExisting(t *testing.T) {
	configDir := setTestConfigHome(t)

	configPath := filepath.Join(configDir, "config.toml")
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
