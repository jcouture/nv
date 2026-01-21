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

package exec

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type Runner struct {
	env     map[string]string
	command string
	args    []string
}

func NewRunner(env map[string]string, command string, args []string) *Runner {
	return &Runner{
		env:     env,
		command: command,
		args:    args,
	}
}

func (r *Runner) Run() (int, error) {
	// #nosec G204 -- command execution is the intended behavior for this runner.
	cmd := exec.Command(r.command, r.args...)
	setProcessGroup(cmd)

	cmd.Env = make([]string, 0, len(r.env))
	for k, v := range r.env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return 1, err
	}

	restoreTerminal := setForegroundProcessGroup(cmd)
	defer restoreTerminal()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signalsToForward()...)

	go func() {
		for sig := range sigChan {
			forwardSignal(cmd, sig)
		}
	}()

	err := cmd.Wait()
	signal.Stop(sigChan)
	close(sigChan)

	exitCode, waitErr := exitCodeFromWait(err)
	if waitErr != nil {
		return exitCode, waitErr
	}

	return exitCode, nil
}

func exitCodeFromWait(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() {
				return 128 + int(status.Signal()), nil
			}
			return status.ExitStatus(), nil
		}
	}

	return 1, err
}
