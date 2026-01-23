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
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func TestLoadConfigExistsInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("[general]\nverbosity=\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg := Load()
	if cfg.General.Verbosity != Default().General.Verbosity {
		t.Fatalf("verbosity=%d want %d", cfg.General.Verbosity, Default().General.Verbosity)
	}
}

func TestLoadConfigExistsValid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	cfg := Default()
	cfg.General.Verbosity = 2
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := Load()
	if loaded.General.Verbosity != 2 {
		t.Fatalf("verbosity=%d want 2", loaded.General.Verbosity)
	}
}

func TestLoadCreatesDefaultConfigWhenNoLegacy(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath: %v", err)
	}
	_, err = os.Stat(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected not exist error, got %v", err)
	}

	_ = Load()

	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
}

func TestLoadIgnoresInvalidLegacy(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	legacyPath := filepath.Join(tmpDir, ".nv")
	if err := os.WriteFile(legacyPath, []byte("INVALID"), 0o600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	_, err = io.Copy(w, bytes.NewBufferString("y\n"))
	if err != nil {
		t.Fatalf("write to pipe: %v", err)
	}
	w.Close()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = oldStdin })

	cfg := Load()
	if cfg.General.Verbosity != Default().General.Verbosity {
		t.Fatalf("verbosity=%d want %d", cfg.General.Verbosity, Default().General.Verbosity)
	}
}

func TestConfigExistsError(t *testing.T) {
	tmpDir := t.TempDir()
	badHome := filepath.Join(tmpDir, "bad")
	if err := os.WriteFile(badHome, []byte("file"), 0o600); err != nil {
		t.Fatalf("write bad home: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", badHome)
	xdg.Reload()

	_, err := ConfigExists()
	if err == nil {
		t.Fatal("expected error from ConfigExists")
	}
}

func TestLoadReturnsDefaultOnError(t *testing.T) {
	tmpDir := t.TempDir()
	badHome := filepath.Join(tmpDir, "bad")
	if err := os.WriteFile(badHome, []byte("file"), 0o600); err != nil {
		t.Fatalf("write bad home: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", badHome)
	xdg.Reload()

	cfg := Load()
	if cfg.General.Verbosity != Default().General.Verbosity {
		t.Fatalf("verbosity=%d want %d", cfg.General.Verbosity, Default().General.Verbosity)
	}
}

func TestLoadConfigHomeError(t *testing.T) {
	orig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = orig })

	xdg.ConfigHome = ""
	cfg := Load()
	if cfg.General.Verbosity != Default().General.Verbosity {
		t.Fatalf("verbosity=%d want %d", cfg.General.Verbosity, Default().General.Verbosity)
	}
}
