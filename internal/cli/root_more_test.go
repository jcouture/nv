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
	"os"
	"testing"
)

func TestExecuteCommandUnknown(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	os.Args = []string{"nv", "does-not-exist"}
	code := executeCommand("")
	if code == 0 {
		t.Fatalf("expected non-zero exit code")
	}
}

func TestExecuteCommandVersion(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	os.Args = []string{"nv", "version"}
	code := executeCommand("")
	if code != 0 {
		t.Fatalf("expected zero exit code")
	}
}

func TestNewRootCmdName(t *testing.T) {
	cmd := NewRootCmd("custom")
	if cmd.Use != "custom" {
		t.Fatalf("Use=%q, want custom", cmd.Use)
	}
}

func TestExecuteUsesExitFunc(t *testing.T) {
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

	Execute()
	if exitCode != -1 {
		t.Fatalf("expected exitFunc not to be called, got %d", exitCode)
	}
}

func TestExecuteUsesExitFuncOnError(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	t.Cleanup(func() {
		os.Args = origArgs
		exitFunc = origExit
	})

	os.Args = []string{"nv", "does-not-exist"}
	exitCode := -1
	exitFunc = func(code int) {
		exitCode = code
	}

	Execute()
	if exitCode == -1 {
		t.Fatalf("expected exitFunc to be called")
	}
}
