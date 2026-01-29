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

	"github.com/jcouture/nv/internal/exporter"
	"github.com/spf13/cobra"
)

type exportOptions struct {
	envFiles     []string
	cascade      bool
	env          string
	overrides    []string
	strict       bool
	preserve     []string
	format       string
	unredacted   bool
	maskPatterns []string
	verbose      bool
}

func newExportCmd() *cobra.Command {
	opts := &exportOptions{}

	cmd := &cobra.Command{
		Use:   "export [flags]",
		Short: "Export the compiled environment",
		Long:  "Load environment variables from .env files and print the compiled result.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(opts)
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&opts.envFiles, "env-file", "e", []string{defaultEnvFile}, "Environment file(s) to load")
	flags.BoolVarP(&opts.cascade, "cascade", "c", false, "Enable cascading mode")
	flags.StringVar(&opts.env, "env", "", "Environment name for cascading")
	flags.StringSliceVarP(&opts.overrides, "override", "o", nil, "Inline overrides (KEY=value)")
	flags.BoolVar(&opts.strict, "strict", false, "Fail on unresolved interpolation")
	flags.StringSliceVarP(&opts.preserve, "preserve", "p", []string{"PATH", "HOME", "USER"}, "System variables to preserve")
	flags.StringVar(&opts.format, "format", exporter.FormatShell, "Export format (shell or json)")
	flags.BoolVar(&opts.unredacted, "unredacted", false, "Show unredacted values")
	flags.StringSliceVar(&opts.maskPatterns, "mask-pattern", nil, "Additional regex patterns to mask by value")
	flags.BoolVarP(&opts.verbose, "verbose", "v", false, "Enable verbose env loading output")

	return cmd
}

func runExport(opts *exportOptions) error {
	env, err := loadEnvironment(envOptions{
		envFiles:  opts.envFiles,
		cascade:   opts.cascade,
		env:       opts.env,
		overrides: opts.overrides,
		strict:    opts.strict,
		preserve:  opts.preserve,
		verbose:   opts.verbose,
		trace:     opts.verbose,
	})
	if err != nil {
		return err
	}

	return exporter.Write(os.Stdout, env, exporter.Options{
		Format:       opts.format,
		Unredacted:   opts.unredacted,
		MaskPatterns: opts.maskPatterns,
	})
}
