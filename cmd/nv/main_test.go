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

package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHelpAndVersion(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
		wantCode int
	}{
		{
			name:     "version flag",
			args:     []string{"nv", "version"},
			contains: "nv version",
			wantCode: 0,
		},
		{
			name:     "short version flag",
			args:     []string{"nv", "-v"},
			contains: "nv version",
			wantCode: 0,
		},
		{
			name:     "long version flag",
			args:     []string{"nv", "--version"},
			contains: "nv version",
			wantCode: 0,
		},
		{
			name:     "help for short args",
			args:     []string{"nv"},
			contains: "usage: nv",
			wantCode: 0,
		},
		{
			name:     "help for unknown flag",
			args:     []string{"nv", "help"},
			contains: "usage: nv",
			wantCode: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout := captureStdout(t, func() {
				code := run(tc.args)
				if code != tc.wantCode {
					t.Fatalf("expected code %d, got %d", tc.wantCode, code)
				}
			})
			if !strings.Contains(stdout, tc.contains) {
				t.Fatalf("unexpected output: %s", stdout)
			}
		})
	}
}

func TestMainExitCode(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	t.Cleanup(func() {
		os.Args = origArgs
		exitFunc = origExit
	})

	os.Args = []string{"nv", "version"}
	exitCode := -1
	exitFunc = func(code int) {
		exitCode = code
	}

	main()

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
}

func TestRunExecPath(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	origExec := execFunc
	origLookPath := lookPathFunc
	origEnv := snapshotEnv()
	t.Cleanup(func() {
		execFunc = origExec
		lookPathFunc = origLookPath
		restoreEnv(origEnv)
	})

	lookPathFunc = func(file string) (string, error) {
		return file, nil
	}
	execFunc = func(argv0 string, argv []string, envv []string) error {
		return nil
	}

	code := run([]string{"nv", envFile, "echo", "hello"})
	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
}

func TestRunMissingEnvFile(t *testing.T) {
	code := run([]string{"nv", "missing.env", "echo"})
	if code != -1 {
		t.Fatalf("expected code -1, got %d", code)
	}
}

func TestRunExecError(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	origExec := execFunc
	origLookPath := lookPathFunc
	origEnv := snapshotEnv()
	t.Cleanup(func() {
		execFunc = origExec
		lookPathFunc = origLookPath
		restoreEnv(origEnv)
	})

	lookPathFunc = func(file string) (string, error) {
		return "", errors.New("missing")
	}
	execFunc = func(argv0 string, argv []string, envv []string) error {
		return errors.New("boom")
	}

	code := run([]string{"nv", envFile, "echo", "hello"})
	if code != -1 {
		t.Fatalf("expected code -1, got %d", code)
	}
}

func TestRunSetvarsError(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	origSetvars := setvarsFunc
	origEnv := snapshotEnv()
	t.Cleanup(func() {
		setvarsFunc = origSetvars
		restoreEnv(origEnv)
	})

	setvarsFunc = func(vars map[string]string) error {
		return errors.New("setvars failed")
	}

	code := run([]string{"nv", envFile, "echo"})
	if code != -1 {
		t.Fatalf("expected code -1, got %d", code)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = writer

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = buf.ReadFrom(reader)
		close(done)
	}()

	fn()

	_ = writer.Close()
	<-done

	os.Stdout = origStdout

	return buf.String()
}

func snapshotEnv() map[string]string {
	env := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

func restoreEnv(vars map[string]string) {
	current := snapshotEnv()
	for key := range current {
		if _, ok := vars[key]; !ok {
			_ = os.Unsetenv(key)
		}
	}
	for key, val := range vars {
		_ = os.Setenv(key, val)
	}
}
