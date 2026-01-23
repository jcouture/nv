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

func TestLoadConfigExistsInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	path, err := GetConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte("[general]\nverbosity=\n"), 0o600))

	cfg := Load()
	require.Equal(t, Default().General.Verbosity, cfg.General.Verbosity)
}

func TestLoadConfigExistsValid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	cfg := Default()
	cfg.General.Verbosity = 2
	require.NoError(t, cfg.Save())

	loaded := Load()
	require.Equal(t, 2, loaded.General.Verbosity)
}

func TestLoadIgnoresInvalidLegacy(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(tmpDir, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("INVALID"), 0o600))

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	_, err = io.Copy(w, bytes.NewBufferString("y\n"))
	require.NoError(t, err)
	w.Close()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = oldStdin })

	cfg := Load()
	require.Equal(t, Default().General.Verbosity, cfg.General.Verbosity)
}

func TestConfigExistsError(t *testing.T) {
	tmpDir := t.TempDir()
	badHome := filepath.Join(tmpDir, "bad")
	require.NoError(t, os.WriteFile(badHome, []byte("file"), 0o600))
	t.Setenv("XDG_CONFIG_HOME", badHome)
	xdg.Reload()

	_, err := ConfigExists()
	require.Error(t, err)
}

func TestLoadReturnsDefaultOnError(t *testing.T) {
	tmpDir := t.TempDir()
	badHome := filepath.Join(tmpDir, "bad")
	require.NoError(t, os.WriteFile(badHome, []byte("file"), 0o600))
	t.Setenv("XDG_CONFIG_HOME", badHome)
	xdg.Reload()

	cfg := Load()
	require.Equal(t, Default().General.Verbosity, cfg.General.Verbosity)
}

func TestLoadConfigHomeError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	cfg := Load()
	require.Equal(t, Default().General.Verbosity, cfg.General.Verbosity)
}
