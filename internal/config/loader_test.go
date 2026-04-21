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
	"runtime"
	"sync"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	want := setTestConfigHome(t)
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir: %v", err)
	}
	if dir != want {
		t.Fatalf("dir=%s want %s", dir, want)
	}
}

func TestLoadFromPathValid(t *testing.T) {
	path := filepath.Join("testdata", "valid_config.toml")
	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}
	if cfg.General.Verbosity != 2 {
		t.Fatalf("verbosity=%d want 2", cfg.General.Verbosity)
	}
	if cfg.Globals.Priority != GlobalsPriorityLast {
		t.Fatalf("priority=%s want %s", cfg.Globals.Priority, GlobalsPriorityLast)
	}
	if cfg.Globals.Env["AWS_REGION"] != "us-east-1" {
		t.Fatalf("AWS_REGION=%s want us-east-1", cfg.Globals.Env["AWS_REGION"])
	}
}

func TestLoadFromPathInvalidSyntax(t *testing.T) {
	path := filepath.Join("testdata", "invalid_syntax.toml")
	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatal("expected error for invalid syntax")
	}
}

func TestLoadFromPathInvalidValuesFixes(t *testing.T) {
	path := filepath.Join("testdata", "invalid_values.toml")
	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	defaults := Default()
	if cfg.Globals.Priority != defaults.Globals.Priority {
		t.Fatalf("priority=%s want %s", cfg.Globals.Priority, defaults.Globals.Priority)
	}
	if cfg.General.Verbosity != defaults.General.Verbosity {
		t.Fatalf("verbosity=%d want %d", cfg.General.Verbosity, defaults.General.Verbosity)
	}
	if len(cfg.Globals.Env) != 0 {
		t.Fatalf("expected empty env, got %v", cfg.Globals.Env)
	}
}

func TestSaveToPathCreatesDir(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "nested", "config.toml")

	cfg := Default()
	if err := cfg.SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath: %v", err)
	}
	_, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat saved config: %v", err)
	}
}

func TestSaveToPathInitializesGlobalsEnv(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "config.toml")

	cfg := Default()
	cfg.Globals.Env = nil
	if err := cfg.SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath: %v", err)
	}
	if cfg.Globals.Env == nil {
		t.Fatal("Globals.Env should be initialized")
	}
}

func TestSaveToPathPermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	temp := t.TempDir()
	readonly := filepath.Join(temp, "readonly")
	if err := os.MkdirAll(readonly, 0o555); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	path := filepath.Join(readonly, "config.toml")
	cfg := Default()
	if err := os.Chmod(readonly, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	err := cfg.SaveToPath(path)
	if err == nil {
		t.Fatal("expected permission error")
	}
}

func TestConcurrentConfigReads(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "config.toml")
	cfg := Default()
	if err := cfg.SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath: %v", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 5)
	for range 5 {
		wg.Go(func() {
			loaded, err := LoadFromPath(path)
			if err != nil {
				errCh <- err
				return
			}
			if loaded.General.Verbosity != cfg.General.Verbosity {
				errCh <- fmt.Errorf("verbosity=%d, want %d", loaded.General.Verbosity, cfg.General.Verbosity)
			}
		})
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("LoadFromPath: %v", err)
		}
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
			if err != nil {
				t.Fatalf("ExpandHome: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %s want %s", got, tt.want)
			}
		})
	}
}

func TestExpandHomeError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("home dir resolution differs on windows")
	}
	t.Setenv("HOME", "")
	_, err := ExpandHome("~/config")
	if err == nil {
		t.Fatal("expected error when HOME unset")
	}
}

func TestConfigExists(t *testing.T) {
	_ = setTestConfigHome(t)

	exists, err := ConfigExists()
	if err != nil {
		t.Fatalf("ConfigExists: %v", err)
	}
	if exists {
		t.Fatalf("expected config to be absent")
	}

	cfg := Default()
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	exists, err = ConfigExists()
	if err != nil {
		t.Fatalf("ConfigExists: %v", err)
	}
	if !exists {
		t.Fatalf("expected config to exist")
	}
}
