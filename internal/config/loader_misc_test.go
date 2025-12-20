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
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
)

func TestGetConfigDirAndPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	xdg.Reload()

	dir, err := GetConfigDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmpDir, "nv"), dir)

	path, err := GetConfigPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmpDir, "nv", "config.toml"), path)
}

func TestGetConfigDirErrorWhenEmpty(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	_, err := GetConfigDir()
	require.Error(t, err)

	_, err = GetConfigPath()
	require.Error(t, err)
}

func TestEnsureConfigDirErrorWhenEmpty(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	err := EnsureConfigDir()
	require.Error(t, err)
}

func TestSaveErrorWhenConfigHomeEmpty(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	cfg := Default()
	err := cfg.Save()
	require.Error(t, err)
}

func TestLegacyEnvExistsFalse(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	exists, err := LegacyEnvExists()
	require.NoError(t, err)
	require.False(t, exists)
}

func TestLegacyEnvExistsTrue(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	legacyPath := filepath.Join(tmpDir, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	exists, err := LegacyEnvExists()
	require.NoError(t, err)
	require.True(t, exists)
}

func TestConfigExistsTrue(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	xdg.Reload()

	path, err := GetConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte("[general]\nverbosity=1\n"), 0o600))

	exists, err := ConfigExists()
	require.NoError(t, err)
	require.True(t, exists)
}
