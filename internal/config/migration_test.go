// Copyright 2015-2025 Jean-Philippe Couture
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
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
)

func TestDetectLegacyEnv(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	exists, err := DetectLegacyEnv()
	require.NoError(t, err)
	require.True(t, exists)
}

func TestParseLegacyEnvFile(t *testing.T) {
	path := filepath.Join("testdata", "legacy_nv_simple")
	env, err := ParseLegacyEnvFile(path)
	require.NoError(t, err)
	require.Equal(t, "bar", env["FOO"])
}

func TestParseLegacyEnvFileComplex(t *testing.T) {
	t.Setenv("NV_TEST", "from-env")
	path := filepath.Join("testdata", "legacy_nv_complex")
	env, err := ParseLegacyEnvFile(path)
	require.NoError(t, err)
	require.Equal(t, "bar baz", env["FOO"])
	require.Equal(t, "from-env", env["FROM_ENV"])
}

func TestParseLegacyEnvFileInvalid(t *testing.T) {
	path := filepath.Join("testdata", "legacy_nv_invalid")
	_, err := ParseLegacyEnvFile(path)
	require.Error(t, err)
}

func TestCreateConfigWithGlobals(t *testing.T) {
	globals := map[string]string{"FOO": "bar"}
	cfg := CreateConfigWithGlobals(globals)
	require.Equal(t, "bar", cfg.Globals.Env["FOO"])
}

func TestMigrationSuccessFlow(t *testing.T) {
	temp := t.TempDir()
	configHome := filepath.Join(temp, "xdg")
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	migrated, err := MigrateLegacyEnv()
	require.NoError(t, err)
	require.True(t, migrated)

	configPath := filepath.Join(configHome, "nv", "config.toml")
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	backupPath := filepath.Join(configHome, "nv", "nv.backup")
	_, err = os.Stat(backupPath)
	require.NoError(t, err)

	_, err = os.Stat(legacyPath)
	require.Error(t, err)

	cfg, err := LoadFromPath(configPath)
	require.NoError(t, err)
	require.Equal(t, "bar", cfg.Globals.Env["FOO"])
}

func TestMigrationFailureKeepsLegacy(t *testing.T) {
	temp := t.TempDir()
	configHome := filepath.Join(temp, "xdg")
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("INVALID"), 0o600))

	migrated, err := MigrateLegacyEnv()
	require.Error(t, err)
	require.False(t, migrated)

	_, err = os.Stat(legacyPath)
	require.NoError(t, err)
}

func TestMigrateNonInteractive(t *testing.T) {
	temp := t.TempDir()
	configHome := filepath.Join(temp, "xdg")
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	require.NoError(t, MigrateNonInteractive())
	configPath := filepath.Join(configHome, "nv", "config.toml")
	_, err := os.Stat(configPath)
	require.NoError(t, err)
}

func TestMigrationSkipWhenConfigExists(t *testing.T) {
	temp := t.TempDir()
	configHome := filepath.Join(temp, "xdg")
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	cfg := Default()
	require.NoError(t, cfg.Save())

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	migrated, err := MigrateLegacyEnv()
	require.NoError(t, err)
	require.False(t, migrated)

	_, err = os.Stat(legacyPath)
	require.NoError(t, err)
}

func TestMigrationSkipWhenLegacyMissing(t *testing.T) {
	temp := t.TempDir()
	configHome := filepath.Join(temp, "xdg")
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	migrated, err := MigrateLegacyEnv()
	require.NoError(t, err)
	require.False(t, migrated)
}

func TestPromptMigrationAcceptDecline(t *testing.T) {
	temp := t.TempDir()
	configHome := filepath.Join(temp, "xdg")
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	cases := []struct {
		name   string
		input  string
		accept bool
	}{
		{name: "accept default", input: "\n", accept: true},
		{name: "accept y", input: "y\n", accept: true},
		{name: "decline", input: "n\n", accept: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			oldInteractive := isInteractiveStdin
			isInteractiveStdin = func() (bool, error) { return true, nil }
			t.Cleanup(func() { isInteractiveStdin = oldInteractive })

			old := os.Stdin
			defer func() { os.Stdin = old }()

			r, w, err := os.Pipe()
			require.NoError(t, err)
			_, err = io.Copy(w, bytes.NewBufferString(tt.input))
			require.NoError(t, err)
			w.Close()
			os.Stdin = r

			accept, err := PromptMigration()
			require.NoError(t, err)
			require.Equal(t, tt.accept, accept)
		})
	}
}

func TestPromptMigrationNonInteractive(t *testing.T) {
	oldInteractive := isInteractiveStdin
	isInteractiveStdin = func() (bool, error) { return false, nil }
	t.Cleanup(func() { isInteractiveStdin = oldInteractive })

	accept, err := PromptMigration()
	require.NoError(t, err)
	require.False(t, accept)
}

func TestPromptMigrationEOF(t *testing.T) {
	oldInteractive := isInteractiveStdin
	isInteractiveStdin = func() (bool, error) { return true, nil }
	t.Cleanup(func() { isInteractiveStdin = oldInteractive })

	old := os.Stdin
	defer func() { os.Stdin = old }()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	w.Close()
	os.Stdin = r

	accept, err := PromptMigration()
	require.NoError(t, err)
	require.False(t, accept)
}

func TestPromptMigrationStdinError(t *testing.T) {
	oldInteractive := isInteractiveStdin
	isInteractiveStdin = func() (bool, error) { return false, os.ErrInvalid }
	t.Cleanup(func() { isInteractiveStdin = oldInteractive })

	_, err := PromptMigration()
	require.Error(t, err)
}

func TestMigrateNonInteractiveSkipWhenConfigExists(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	cfg := Default()
	require.NoError(t, cfg.Save())

	require.NoError(t, MigrateNonInteractive())
}

func TestBackupLegacyEnv(t *testing.T) {
	temp := t.TempDir()
	configHome := filepath.Join(temp, "xdg")
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	require.NoError(t, BackupLegacyEnv())
	backupPath := filepath.Join(configHome, "nv", "nv.backup")
	content, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	require.Equal(t, "FOO=bar", string(content))
}

func TestDeleteLegacyEnv(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	require.NoError(t, DeleteLegacyEnv())
	_, err := os.Stat(legacyPath)
	require.Error(t, err)
}

func TestCurrentEnvIncludesSetVar(t *testing.T) {
	const key = "NV_TEST_CURRENT_ENV"
	t.Setenv(key, "value")

	env := currentEnv()
	require.Equal(t, "value", env[key])
}

func TestBackupLegacyEnvConfigDirError(t *testing.T) {
	temp := t.TempDir()
	configHome := filepath.Join(temp, "xdg-file")
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	require.NoError(t, os.WriteFile(configHome, []byte("file"), 0o600))

	legacyPath := filepath.Join(temp, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	err := BackupLegacyEnv()
	require.Error(t, err)
}
