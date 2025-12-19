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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
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

func TestRunDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	opts := &runOptions{
		envFiles: []string{envFile},
		dryRun:   true,
	}

	stdout, stderr := captureOutput(t, func() {
		if err := runRun(opts, []string{"echo", "hello"}); err != nil {
			t.Fatalf("runRun error: %v", err)
		}
	})

	if !strings.Contains(stdout, "FOO=bar") {
		t.Fatalf("stdout missing env output: %s", stdout)
	}
	if !strings.Contains(stderr, "Dry run mode") {
		t.Fatalf("stderr missing dry run notice: %s", stderr)
	}
}

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	origStdout := os.Stdout
	origStderr := os.Stderr

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

	return stdoutBuf.String(), stderrBuf.String()
}
