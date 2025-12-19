//go:build !windows

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

package exec_test

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/jcouture/nv/internal/exec"
)

func TestSignalForwarding(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping signal test in short mode")
	}

	runner := exec.NewRunner(
		map[string]string{},
		"sleep",
		[]string{"10"},
	)

	done := make(chan int, 1)
	go func() {
		code, _ := runner.Run()
		done <- code
	}()

	time.Sleep(100 * time.Millisecond)

	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGINT)

	select {
	case code := <-done:
		if code != 128+int(syscall.SIGINT) && code != 130 {
			t.Fatalf("expected signal exit code, got %d", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("process did not terminate after SIGINT")
	}
}
