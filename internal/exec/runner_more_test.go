//go:build !windows

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

package exec

import (
	"errors"
	"syscall"
	"testing"
)

func TestRunnerStartError(t *testing.T) {
	runner := NewRunner(map[string]string{}, "does-not-exist", nil)
	code, err := runner.Run()
	if err == nil {
		t.Fatal("expected error for missing command")
	}
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func TestRunnerExitStatus(t *testing.T) {
	runner := NewRunner(map[string]string{}, "sh", []string{"-c", "exit 7"})
	code, err := runner.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 7 {
		t.Fatalf("expected exit code 7, got %d", code)
	}
}

func TestRunnerSignalExit(t *testing.T) {
	runner := NewRunner(map[string]string{}, "sh", []string{"-c", "kill -TERM $$"})
	code, err := runner.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 128+int(syscall.SIGTERM) {
		t.Fatalf("expected signal exit code, got %d", code)
	}
}

func TestRunnerSuccess(t *testing.T) {
	runner := NewRunner(map[string]string{}, "sh", []string{"-c", "exit 0"})
	code, err := runner.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestExitCodeFromWaitNonExitError(t *testing.T) {
	err := errors.New("wait failed")
	code, waitErr := exitCodeFromWait(err)
	if waitErr == nil {
		t.Fatal("expected error")
	}
	if code != 1 {
		t.Fatalf("expected code 1, got %d", code)
	}
}

func TestRunnerEnvApplied(t *testing.T) {
	runner := NewRunner(map[string]string{"FOO": "bar"}, "sh", []string{"-c", "exit 0"})
	code, err := runner.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}
