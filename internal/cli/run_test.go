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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/adrg/xdg"
	"github.com/fatih/color"
)

func TestParseOverride(t *testing.T) {
	key, val, err := parseOverride("FOO=bar")
	if err != nil {
		t.Fatalf("parseOverride returned error: %v", err)
	}
	if key != "FOO" || val != "bar" {
		t.Fatalf("got %q=%q, want FOO=bar", key, val)
	}

	if _, _, err := parseOverride("NOVAL"); err == nil {
		t.Fatal("expected error for missing '='")
	}
}

func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	configDir := filepath.Join(tmpDir, "xdg", "nv")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	configPathOnDisk := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPathOnDisk, []byte("[general]\nverbosity=1\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	path := configPath()
	if path != configPathOnDisk {
		t.Fatalf("expected config path %s, got %s", configPathOnDisk, path)
	}
}

func TestRunConfigPathError(t *testing.T) {
	tmpDir := t.TempDir()
	badHome := filepath.Join(tmpDir, "bad")
	if err := os.WriteFile(badHome, []byte("file"), 0o600); err != nil {
		t.Fatalf("write bad home: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", badHome)
	xdg.Reload()

	path := configPath()
	if path != "unknown" {
		t.Fatalf("expected unknown config path")
	}
}

func TestConfigPathMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	path := configPath()
	if path != "-" {
		t.Fatalf("expected '-' for missing config, got %s", path)
	}
}

func TestRunDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()
	envFile := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	opts := &runOptions{
		envFiles: []string{envFile},
		dryRun:   true,
	}

	stdout, stderr := captureOutput(t, func() {
		cmd := newRunCmdForTest(t, map[string]string{
			"env-file": envFile,
			"dry-run":  "true",
			"cascade":  "false",
		})
		if err := runRun(cmd, opts, []string{"echo", "hello"}); err != nil {
			t.Fatalf("runRun error: %v", err)
		}
	})

	if !strings.Contains(stdout, "export FOO=\"bar\"") {
		t.Fatalf("stdout missing export output: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr output, got: %s", stderr)
	}
}

func TestRunUsesConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg"))
	xdg.Reload()

	cfgPath := filepath.Join(tmpDir, "xdg", "nv", "config.toml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	configData := []byte("[defaults]\nenv_file=\".env.custom\"\ndry_run=true\ncascade=false\n\n[validation]\nenabled=false\nschema_file=\".env.example\"\n\n[general]\nverbosity=1\n")
	if err := os.WriteFile(cfgPath, configData, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, ".env.custom"), []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	cmdFlags := newRunCmdForTest(t, nil)
	opts := &runOptions{format: "shell"}

	stdout, _ := captureOutput(t, func() {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd: %v", err)
		}
		if err := os.Chdir(workDir); err != nil {
			t.Fatalf("Chdir: %v", err)
		}
		t.Cleanup(func() { _ = os.Chdir(cwd) })

		if err := runRun(cmdFlags, opts, []string{"echo", "ok"}); err != nil {
			t.Fatalf("runRun: %v", err)
		}
	})

	if !strings.Contains(stdout, "export FOO=\"bar\"") {
		t.Fatalf("unexpected output: %s", stdout)
	}
}

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	origStdout := os.Stdout
	origStderr := os.Stderr
	origColorOut := color.Output
	origColorErr := color.Error

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	color.Output = stdoutWriter
	color.Error = stderrWriter

	var stdoutBuf, stderrBuf bytes.Buffer
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		_, _ = stdoutBuf.ReadFrom(stdoutReader)
		wg.Done()
	}()
	go func() {
		_, _ = stderrBuf.ReadFrom(stderrReader)
		wg.Done()
	}()

	fn()

	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()
	wg.Wait()

	os.Stdout = origStdout
	os.Stderr = origStderr
	color.Output = origColorOut
	color.Error = origColorErr

	return stdoutBuf.String(), stderrBuf.String()
}
