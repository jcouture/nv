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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
)

func TestGetConfigDirWithXDG(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", temp)
	xdg.Reload()

	dir, err := GetConfigDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(temp, "nv"), dir)
}

func TestGetConfigDirWithoutXDG(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", "")
	xdg.Reload()

	dir, err := GetConfigDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(temp, ".config", "nv"), dir)
}

func TestLoadFromPathValid(t *testing.T) {
	path := filepath.Join("testdata", "valid_config.toml")
	cfg, err := LoadFromPath(path)
	require.NoError(t, err)
	require.Equal(t, 2, cfg.General.Verbosity)
	require.Equal(t, GlobalsPriorityLast, cfg.Globals.Priority)
	require.Equal(t, "us-east-1", cfg.Globals.Env["AWS_REGION"])
}

func TestLoadFromPathInvalidSyntax(t *testing.T) {
	path := filepath.Join("testdata", "invalid_syntax.toml")
	_, err := LoadFromPath(path)
	require.Error(t, err)
}

func TestLoadFromPathInvalidValuesFixes(t *testing.T) {
	path := filepath.Join("testdata", "invalid_values.toml")
	cfg, err := LoadFromPath(path)
	require.NoError(t, err)

	defaults := Default()
	require.Equal(t, defaults.Globals.Priority, cfg.Globals.Priority)
	require.Equal(t, defaults.General.Verbosity, cfg.General.Verbosity)
	require.Empty(t, cfg.Globals.Env)
}

func TestLoadWithMigrationNoConfig(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))
	xdg.Reload()

	cfg, migrated, err := LoadWithMigration()
	require.NoError(t, err)
	require.False(t, migrated)
	require.Equal(t, Default().General.Verbosity, cfg.General.Verbosity)
}

func TestSaveToPathCreatesDir(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "nested", "config.toml")

	cfg := Default()
	require.NoError(t, cfg.SaveToPath(path))
	_, err := os.Stat(path)
	require.NoError(t, err)
}

func TestSaveToPathInitializesGlobalsEnv(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "config.toml")

	cfg := Default()
	cfg.Globals.Env = nil
	require.NoError(t, cfg.SaveToPath(path))
	require.NotNil(t, cfg.Globals.Env)
}

func TestSaveToPathPermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	temp := t.TempDir()
	readonly := filepath.Join(temp, "readonly")
	require.NoError(t, os.MkdirAll(readonly, 0o555))

	path := filepath.Join(readonly, "config.toml")
	cfg := Default()
	require.NoError(t, os.Chmod(readonly, 0o555))

	err := cfg.SaveToPath(path)
	require.Error(t, err)
}

func TestConcurrentConfigReads(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "config.toml")
	cfg := Default()
	require.NoError(t, cfg.SaveToPath(path))

	var wg sync.WaitGroup
	errCh := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loaded, err := LoadFromPath(path)
			if err != nil {
				errCh <- err
				return
			}
			if loaded.General.Verbosity != cfg.General.Verbosity {
				errCh <- fmt.Errorf("verbosity=%d, want %d", loaded.General.Verbosity, cfg.General.Verbosity)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}

func TestExpandHome(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "tilde only", in: "~", want: temp},
		{name: "tilde subdir", in: "~/config", want: filepath.Join(temp, "config")},
		{name: "no tilde", in: "/tmp/nv", want: "/tmp/nv"},
		{name: "other tilde", in: "~other/path", want: "~other/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandHome(tt.in)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestExpandHomeError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("home dir resolution differs on windows")
	}
	t.Setenv("HOME", "")
	_, err := ExpandHome("~/config")
	require.Error(t, err)
}

func TestGetLegacyEnvPath(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)

	path, err := GetLegacyEnvPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(temp, ".nv"), path)
}

func TestGetLegacyEnvPathError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("home dir resolution differs on windows")
	}
	t.Setenv("HOME", "")
	_, err := GetLegacyEnvPath()
	require.Error(t, err)
}

func TestConfigExists(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", temp)
	xdg.Reload()

	exists, err := ConfigExists()
	require.NoError(t, err)
	require.False(t, exists)

	cfg := Default()
	require.NoError(t, cfg.Save())

	exists, err = ConfigExists()
	require.NoError(t, err)
	require.True(t, exists)
}
