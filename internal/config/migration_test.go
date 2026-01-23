// Copyright (c) 2020-2025 Jonathan Couture
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
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
)

func TestDetectLegacyEnv(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	path := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(path, []byte("FOO=bar"), 0o600))

	exists, err := DetectLegacyEnv()
	require.NoError(t, err)
	require.True(t, exists)
}

func TestParseLegacyEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".nv")
	require.NoError(t, os.WriteFile(path, []byte("FOO=bar"), 0o600))

	env, err := ParseLegacyEnvFile(path)
	require.NoError(t, err)
	require.Equal(t, "bar", env["FOO"])
}

func TestParseLegacyEnvFileInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".nv")
	require.NoError(t, os.WriteFile(path, []byte("FOO"), 0o600))

	_, err := ParseLegacyEnvFile(path)
	require.Error(t, err)
}

func TestMigrateLegacyEnv(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	migrated, err := MigrateLegacyEnv()
	require.NoError(t, err)
	require.True(t, migrated)

	cfg := Load()
	require.Equal(t, "bar", cfg.Globals.Env["FOO"])
}

func TestMigrateLegacyEnvNoLegacy(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	migrated, err := MigrateLegacyEnv()
	require.NoError(t, err)
	require.False(t, migrated)
}

func TestMigrateNonInteractive(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	require.NoError(t, MigrateNonInteractive())

	cfg := Load()
	require.Equal(t, "bar", cfg.Globals.Env["FOO"])
}

func TestBackupAndDeleteLegacyEnv(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	require.NoError(t, BackupLegacyEnv())
	backupPath := filepath.Join(filepath.Join(temp, "xdg", "nv"), "nv.backup")
	_, err := os.Stat(backupPath)
	require.NoError(t, err)

	require.NoError(t, DeleteLegacyEnv())
	_, err = os.Stat(legacyPath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestPromptMigrationNonInteractive(t *testing.T) {
	oldInteractive := isInteractiveStdin
	isInteractiveStdin = func() (bool, error) { return false, nil }
	t.Cleanup(func() { isInteractiveStdin = oldInteractive })

	should, err := PromptMigration()
	require.NoError(t, err)
	require.False(t, should)
}
