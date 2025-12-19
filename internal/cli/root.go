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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
	exitFunc  = os.Exit
)

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func NewRootCmd(name string) *cobra.Command {
	if name == "" {
		name = "nvx"
	}
	var noColor bool
	rootCmd := &cobra.Command{
		Use:           name,
		Short:         "Environment variable manager",
		Long:          "Loads context-specific environment variables from .env files before executing a command.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureColors(noColor)
	}

	rootCmd.AddCommand(
		newRunCmd(),
		newExportCmd(),
		newPrintCmd(),
		newValidateCmd(),
		newVersionCmd(),
	)

	return rootCmd
}

func Execute() {
	name := os.Getenv("NVX_COMMAND_NAME")
	exitCode := executeCommand(name)
	if exitCode != 0 {
		exitFunc(exitCode)
	}
}

func executeCommand(name string) int {
	if err := NewRootCmd(name).Execute(); err != nil {
		if exitErr, ok := err.(exitError); ok {
			return exitErr.code
		}
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
