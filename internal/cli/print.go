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
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/jcouture/nv/pkg/env"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"
)

type printOptions struct {
	ignoreCase bool
	sort       bool
}

func newPrintCmd() *cobra.Command {
	opts := &printOptions{}

	cmd := &cobra.Command{
		Use:   "print [flags] [name]",
		Short: "Print environment variables",
		Long:  "Print the names and values of the variables in the current environment.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrint(opts, args)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&opts.ignoreCase, "ignore-case", false, "Ignore case distinctions in name")
	flags.BoolVar(&opts.sort, "sort", false, "Sort alphabetically by name")

	return cmd
}

func runPrint(opts *printOptions, args []string) error {
	vars := env.Getvars()
	names := env.Getnames(vars)

	if opts.sort {
		sort.Strings(names)
	}

	var query string
	if len(args) > 0 {
		query = args[0]
	}

	if query != "" {
		if opts.ignoreCase {
			names = fuzzy.FindNormalizedFold(query, names)
		} else {
			names = fuzzy.FindNormalized(query, names)
		}
	}

	for _, name := range names {
		printVar(name, vars[name])
	}

	return nil
}

func printVar(name, value string) {
	fmt.Fprintf(os.Stdout, "%v=%v\n", color.HiGreenString(name), color.HiBlueString(value))
}
