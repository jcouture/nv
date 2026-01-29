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
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/mattn/go-isatty"
	"golang.org/x/sys/unix"
)

func setForegroundProcessGroup(cmd *exec.Cmd) func() {
	if cmd == nil || cmd.Process == nil {
		return func() {}
	}
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return func() {}
	}

	fd := int(os.Stdin.Fd())
	originalGroup, err := unix.IoctlGetInt(fd, unix.TIOCGPGRP)
	if err != nil {
		return func() {}
	}

	pgid, err := unix.Getpgid(cmd.Process.Pid)
	if err != nil {
		return func() {}
	}

	runWithoutSigttou(func() {
		_ = unix.IoctlSetPointerInt(fd, unix.TIOCSPGRP, pgid)
	})

	return func() {
		runWithoutSigttou(func() {
			_ = unix.IoctlSetPointerInt(fd, unix.TIOCSPGRP, originalGroup)
		})
	}
}

func runWithoutSigttou(fn func()) {
	signal.Ignore(syscall.SIGTTOU)
	defer signal.Reset(syscall.SIGTTOU)
	fn()
}
