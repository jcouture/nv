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

package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFilesOverride(t *testing.T) {
	tmpDir := t.TempDir()
	base := filepath.Join(tmpDir, ".env")
	override := filepath.Join(tmpDir, ".env.local")

	if err := os.WriteFile(base, []byte("FOO=bar\nBAZ=qux\n"), 0o644); err != nil {
		t.Fatalf("write base env: %v", err)
	}
	if err := os.WriteFile(override, []byte("FOO=overridden\n"), 0o644); err != nil {
		t.Fatalf("write override env: %v", err)
	}

	t.Setenv("PATH", "/usr/bin")
	t.Setenv("EXTRA", "1")
	l := New()
	env, err := l.LoadFiles(base, override)
	if err != nil {
		t.Fatalf("LoadFiles error: %v", err)
	}

	if env["FOO"] != "overridden" {
		t.Fatalf("FOO=%q, want overridden", env["FOO"])
	}
	if env["BAZ"] != "qux" {
		t.Fatalf("BAZ=%q, want qux", env["BAZ"])
	}
	if env["PATH"] != "/usr/bin" {
		t.Fatalf("PATH=%q, want preserved", env["PATH"])
	}
}

func TestLoadCascadeSkipsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	base := filepath.Join(tmpDir, ".env")
	envFile := filepath.Join(tmpDir, ".env.development")

	if err := os.WriteFile(base, []byte("BASE=1\n"), 0o644); err != nil {
		t.Fatalf("write base env: %v", err)
	}
	if err := os.WriteFile(envFile, []byte("ENV=dev\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	l := New()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	env, err := l.LoadCascade("development")
	if err != nil {
		t.Fatalf("LoadCascade error: %v", err)
	}
	if env["BASE"] != "1" {
		t.Fatalf("BASE=%q, want 1", env["BASE"])
	}
	if env["ENV"] != "dev" {
		t.Fatalf("ENV=%q, want dev", env["ENV"])
	}
}

func TestPreservedEnvWithCustomVar(t *testing.T) {
	const key = "NV_TEST_PRESERVE"
	t.Setenv(key, "kept")

	l := New(WithPreserve([]string{key}))
	env := l.PreservedEnv()
	if env[key] != "kept" {
		t.Fatalf("expected %s=kept, got %q", key, env[key])
	}
}

func TestWithTracer(t *testing.T) {
	called := false
	l := New(WithTracer(func(event TraceEvent) {
		called = true
	}))

	_, err := l.LoadFilesWithEnv(map[string]string{"BASE": "1"}, t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing env file")
	}
	if called {
		t.Fatal("expected tracer not to be called on error")
	}

	file := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(file, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	called = false
	_, err = l.LoadFilesWithEnv(map[string]string{"BASE": "1"}, file)
	if err != nil {
		t.Fatalf("LoadFilesWithEnv error: %v", err)
	}
	if !called {
		t.Fatal("expected tracer to be called when file loads")
	}
}

func TestLoadFilesWithEnvNilPreserves(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(path, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	const key = "NV_TEST_KEEP"
	t.Setenv(key, "1")

	l := New(WithPreserve([]string{key}))
	env, err := l.LoadFilesWithEnv(nil, path)
	if err != nil {
		t.Fatalf("LoadFilesWithEnv error: %v", err)
	}
	if env[key] != "1" {
		t.Fatalf("expected %s=1, got %q", key, env[key])
	}
	if env["FOO"] != "bar" {
		t.Fatalf("expected FOO=bar, got %q", env["FOO"])
	}
}

func TestLoadOptionalFilesWithEnvSkipsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	base := filepath.Join(tmpDir, ".env")
	missing := filepath.Join(tmpDir, ".env.local")

	if err := os.WriteFile(base, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write base env: %v", err)
	}

	l := New()
	env, err := l.LoadOptionalFilesWithEnv(nil, base, missing)
	if err != nil {
		t.Fatalf("LoadOptionalFilesWithEnv error: %v", err)
	}
	if env["FOO"] != "bar" {
		t.Fatalf("FOO=%q, want bar", env["FOO"])
	}
}
