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

func TestLoadWithMigrationDecline(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(tmpDir, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	oldInteractive := isInteractiveStdin
	isInteractiveStdin = func() (bool, error) { return true, nil }
	t.Cleanup(func() { isInteractiveStdin = oldInteractive })

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	_, err = io.Copy(w, bytes.NewBufferString("n\n"))
	require.NoError(t, err)
	w.Close()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = oldStdin })

	cfg, migrated, err := LoadWithMigration()
	require.NoError(t, err)
	require.False(t, migrated)
	require.Equal(t, Default().General.Verbosity, cfg.General.Verbosity)
}

func TestLoadWithMigrationAccept(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(tmpDir, ".nv")
	require.NoError(t, os.WriteFile(legacyPath, []byte("FOO=bar"), 0o600))

	oldInteractive := isInteractiveStdin
	isInteractiveStdin = func() (bool, error) { return true, nil }
	t.Cleanup(func() { isInteractiveStdin = oldInteractive })

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	_, err = io.Copy(w, bytes.NewBufferString("y\n"))
	require.NoError(t, err)
	w.Close()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = oldStdin })

	cfg, migrated, err := LoadWithMigration()
	require.NoError(t, err)
	require.True(t, migrated)
	require.Equal(t, "bar", cfg.Globals.Env["FOO"])
}
