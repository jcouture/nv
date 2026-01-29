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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func TestDetectLegacyEnv(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	path := filepath.Join(temp, ".nv")
	if err := os.WriteFile(path, []byte("FOO=bar"), 0o600); err != nil {
		t.Fatalf("write legacy env: %v", err)
	}

	exists, err := DetectLegacyEnv()
	if err != nil {
		t.Fatalf("DetectLegacyEnv: %v", err)
	}
	if !exists {
		t.Fatal("expected legacy env to exist")
	}
}

func TestParseLegacyEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".nv")
	if err := os.WriteFile(path, []byte("FOO=bar"), 0o600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	env, err := ParseLegacyEnvFile(path)
	if err != nil {
		t.Fatalf("ParseLegacyEnvFile: %v", err)
	}
	if env["FOO"] != "bar" {
		t.Fatalf("FOO=%s want bar", env["FOO"])
	}
}

func TestParseLegacyEnvFileInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".nv")
	if err := os.WriteFile(path, []byte("FOO"), 0o600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	_, err := ParseLegacyEnvFile(path)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestMigrateLegacyEnv(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	if err := os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	migrated, err := MigrateLegacyEnv()
	if err != nil {
		t.Fatalf("MigrateLegacyEnv: %v", err)
	}
	if !migrated {
		t.Fatal("expected migration to occur")
	}

	cfg := Load()
	if cfg.Globals.Env["FOO"] != "bar" {
		t.Fatalf("FOO=%s want bar", cfg.Globals.Env["FOO"])
	}
}

func TestMigrateLegacyEnvNoLegacy(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	migrated, err := MigrateLegacyEnv()
	if err != nil {
		t.Fatalf("MigrateLegacyEnv: %v", err)
	}
	if migrated {
		t.Fatal("expected no migration")
	}
}

func TestMigrateNonInteractive(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	if err := os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	if err := MigrateNonInteractive(); err != nil {
		t.Fatalf("MigrateNonInteractive: %v", err)
	}

	cfg := Load()
	if cfg.Globals.Env["FOO"] != "bar" {
		t.Fatalf("FOO=%s want bar", cfg.Globals.Env["FOO"])
	}
}

func TestBackupAndDeleteLegacyEnv(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	if err := os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	if err := BackupLegacyEnv(); err != nil {
		t.Fatalf("BackupLegacyEnv: %v", err)
	}
	backupPath := filepath.Join(filepath.Join(temp, "xdg", "nv"), "nv.backup")
	_, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("stat backup: %v", err)
	}

	if err := DeleteLegacyEnv(); err != nil {
		t.Fatalf("DeleteLegacyEnv: %v", err)
	}
	_, err = os.Stat(legacyPath)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected legacy file removed, got %v", err)
	}
}

func TestPromptMigrationNonInteractive(t *testing.T) {
	oldInteractive := isInteractiveStdin
	isInteractiveStdin = func() (bool, error) { return false, nil }
	t.Cleanup(func() { isInteractiveStdin = oldInteractive })

	should, err := PromptMigration()
	if err != nil {
		t.Fatalf("PromptMigration: %v", err)
	}
	if should {
		t.Fatal("expected prompt to skip when non-interactive")
	}
}
