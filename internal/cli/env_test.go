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

package cli

import (
	"path/filepath"
	"testing"
)

func TestAutoLocalForFilesAbsolute(t *testing.T) {
	dir := t.TempDir()
	files := []string{filepath.Join(dir, ".env")}
	out := autoLocalForFiles(files)
	if out != filepath.Join(dir, ".env.local") {
		t.Fatalf("out=%s want %s", out, filepath.Join(dir, ".env.local"))
	}
}

func TestMergeEnv(t *testing.T) {
	base := map[string]string{"A": "1"}
	mergeEnv(base, map[string]string{"B": "2", "A": "3"})
	if base["A"] != "3" || base["B"] != "2" {
		t.Fatalf("merged base=%v want A=3 B=2", base)
	}
}

func TestAutoLocalForFilesNoEnv(t *testing.T) {
	files := []string{"/tmp/custom.env"}
	out := autoLocalForFiles(files)
	if out != "" {
		t.Fatalf("expected empty result, got %s", out)
	}
}

func TestAutoLocalForFilesAlreadyPresent(t *testing.T) {
	dir := t.TempDir()
	files := []string{filepath.Join(dir, ".env"), filepath.Join(dir, ".env.local")}
	out := autoLocalForFiles(files)
	if out != "" {
		t.Fatalf("expected empty result, got %s", out)
	}
}
