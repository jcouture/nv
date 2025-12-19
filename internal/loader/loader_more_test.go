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

func TestLoadFilesMissingFile(t *testing.T) {
	l := New()
	_, err := l.LoadFiles("does-not-exist.env")
	if err == nil {
		t.Fatal("expected error for missing env file")
	}
}

func TestLoadFilesStrictInterpolation(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=${BAR}\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	l := New(WithStrict(true))
	_, err := l.LoadFiles(envFile)
	if err == nil {
		t.Fatal("expected strict interpolation error")
	}
}

func TestLoadFilesCustomPreserve(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	t.Setenv("HOME", "/tmp/home")
	t.Setenv("PATH", "/bin")

	l := New(WithPreserve([]string{"HOME"}))
	env, err := l.LoadFiles(envFile)
	if err != nil {
		t.Fatalf("LoadFiles error: %v", err)
	}
	if env["HOME"] != "/tmp/home" {
		t.Fatalf("HOME=%q, want /tmp/home", env["HOME"])
	}
	if _, ok := env["PATH"]; ok {
		t.Fatalf("expected PATH to be absent")
	}
}

func TestLoadCascadeUsesNVEnv(t *testing.T) {
	tmpDir := t.TempDir()
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

	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("BASE=1\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.staging"), []byte("ENV=staging\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	t.Setenv("NV_ENV", "staging")

	l := New()
	env, err := l.LoadCascade("")
	if err != nil {
		t.Fatalf("LoadCascade error: %v", err)
	}
	if env["ENV"] != "staging" {
		t.Fatalf("ENV=%q, want staging", env["ENV"])
	}
}

func TestLoadCascadeDefaultEnv(t *testing.T) {
	tmpDir := t.TempDir()
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

	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("BASE=1\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.development"), []byte("ENV=dev\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	l := New()
	env, err := l.LoadCascade("")
	if err != nil {
		t.Fatalf("LoadCascade error: %v", err)
	}
	if env["ENV"] != "dev" {
		t.Fatalf("ENV=%q, want dev", env["ENV"])
	}
}
