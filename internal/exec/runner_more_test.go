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

func TestRunnerRun(t *testing.T) {
	cases := []struct {
		name     string
		env      map[string]string
		command  string
		args     []string
		wantCode int
		wantErr  bool
	}{
		{
			name:     "success",
			env:      map[string]string{},
			command:  "sh",
			args:     []string{"-c", "exit 0"},
			wantCode: 0,
		},
		{
			name:     "non-zero exit status",
			env:      map[string]string{},
			command:  "sh",
			args:     []string{"-c", "exit 7"},
			wantCode: 7,
		},
		{
			name:     "signal exit code",
			env:      map[string]string{},
			command:  "sh",
			args:     []string{"-c", "kill -TERM $$"},
			wantCode: 128 + int(syscall.SIGTERM),
		},
		{
			name:     "missing command returns error",
			env:      map[string]string{},
			command:  "does-not-exist",
			args:     nil,
			wantCode: 1,
			wantErr:  true,
		},
		{
			name:     "env is applied to subprocess",
			env:      map[string]string{"NV_RUNNER_TEST_VAR": "expected"},
			command:  "sh",
			args:     []string{"-c", `test "$NV_RUNNER_TEST_VAR" = "expected"`},
			wantCode: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := NewRunner(tc.env, tc.command, tc.args)
			code, err := runner.Run()
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d", code, tc.wantCode)
			}
		})
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
