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
	"sort"
	"strings"

	"github.com/jcouture/nv/internal/exec"
	"github.com/jcouture/nv/internal/loader"
	"github.com/spf13/cobra"
)

type runOptions struct {
	envFiles  []string
	cascade   bool
	env       string
	overrides []string
	strict    bool
	preserve  []string
	dryRun    bool
}

func newRunCmd() *cobra.Command {
	opts := &runOptions{}

	cmd := &cobra.Command{
		Use:   "run [flags] -- <command> [args...]",
		Short: "Execute a command with loaded environment",
		Long:  "Load environment variables from .env files and execute the specified command.",
		Example: `  nvx run -e .env -- ./myapp
  nvx run -e .env -e .env.local -- npm start
  nvx run --cascade --env=production -- ./deploy.sh`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(opts, args)
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&opts.envFiles, "env-file", "e", []string{defaultEnvFile}, "Environment file(s) to load")
	flags.BoolVarP(&opts.cascade, "cascade", "c", false, "Enable cascading mode")
	flags.StringVar(&opts.env, "env", "", "Environment name for cascading")
	flags.StringSliceVarP(&opts.overrides, "override", "o", nil, "Inline overrides (KEY=value)")
	flags.BoolVar(&opts.strict, "strict", false, "Fail on unresolved interpolation")
	flags.StringSliceVarP(&opts.preserve, "preserve", "p", []string{"PATH", "HOME", "USER"}, "System variables to preserve")
	flags.BoolVar(&opts.dryRun, "dry-run", false, "Print environment without executing")

	return cmd
}

func runRun(opts *runOptions, args []string) error {
	loaderOpts := []loader.Option{
		loader.WithPreserve(opts.preserve),
		loader.WithStrict(opts.strict),
	}

	l := loader.New(loaderOpts...)

	var env map[string]string
	var err error

	if opts.cascade {
		env, err = l.LoadCascade(opts.env)
	} else {
		env, err = l.LoadFiles(opts.envFiles...)
	}
	if err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	for _, override := range opts.overrides {
		key, value, err := parseOverride(override)
		if err != nil {
			return fmt.Errorf("invalid override %q: %w", override, err)
		}
		env[key] = value
	}

	if opts.dryRun {
		fmt.Fprintln(os.Stderr, "# Dry run mode - command will not be executed")
		fmt.Fprintln(os.Stderr, "# Environment:")
		keys := make([]string, 0, len(env))
		for k := range env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(os.Stdout, "%s=%s\n", k, env[k])
		}
		fmt.Fprintln(os.Stderr, "# Command:")
		fmt.Fprintln(os.Stderr, "#", strings.Join(args, " "))
		return nil
	}

	runner := exec.NewRunner(env, args[0], args[1:])
	exitCode, err := runner.Run()
	if err != nil {
		return err
	}

	return exitError{code: exitCode}
}

func parseOverride(s string) (string, string, error) {
	for i, c := range s {
		if c == '=' {
			if i == 0 {
				return "", "", fmt.Errorf("missing key")
			}
			return s[:i], s[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("missing '=' in override")
}
