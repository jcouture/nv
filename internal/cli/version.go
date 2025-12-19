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
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

type versionOptions struct {
	short  bool
	format string
}

func newVersionCmd() *cobra.Command {
	opts := &versionOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(opts)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.short, "short", "s", false, "Print only version number")
	flags.StringVarP(&opts.format, "format", "f", "text", "Output format: text, json")

	return cmd
}

func runVersion(opts *versionOptions) error {
	if opts.short {
		fmt.Fprintln(os.Stdout, Version)
		return nil
	}

	switch opts.format {
	case "json":
		data := map[string]string{
			"version": Version,
			"commit":  Commit,
			"built":   BuildDate,
			"go":      runtime.Version(),
		}
		enc := json.NewEncoder(os.Stdout)
		return enc.Encode(data)
	case "text":
		fmt.Fprintf(os.Stdout, "nvx version %s\n", Version)
		fmt.Fprintf(os.Stdout, "commit: %s\n", Commit)
		fmt.Fprintf(os.Stdout, "built: %s\n", BuildDate)
		fmt.Fprintf(os.Stdout, "go: %s\n", runtime.Version())
		return nil
	default:
		return fmt.Errorf("unknown format: %s", opts.format)
	}
}
