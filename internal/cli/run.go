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

	"github.com/jcouture/nv/internal/config"
	"github.com/jcouture/nv/internal/exec"
	"github.com/jcouture/nv/internal/exporter"
	"github.com/spf13/cobra"
)

type runOptions struct {
	envFiles     []string
	cascade      bool
	env          string
	overrides    []string
	strict       bool
	preserve     []string
	dryRun       bool
	validate     bool
	schemaFile   string
	schemaStrict bool
	format       string
	unredacted   bool
	maskPatterns []string
}

func newRunCmd() *cobra.Command {
	opts := &runOptions{}

	cmd := &cobra.Command{
		Use:   "run [flags] -- <command> [args...]",
		Short: "Execute a command with loaded environment",
		Long:  "Load environment variables from .env files and execute the specified command. When --env-file is provided, cascading is automatically disabled (with a warning) so explicit files take precedence.",
		Example: `  nv run -e .env -- ./myapp
	  nv run -e .env -e .env.local -- npm start
	  nv run --cascade --env=production -- ./deploy.sh`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(cmd, opts, args)
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
	flags.BoolVar(&opts.validate, "validate", false, "Validate environment variables before execution")
	flags.StringVar(&opts.schemaFile, "schema", defaultSchemaFile, "Schema file to validate against")
	flags.StringVar(&opts.schemaFile, "schema-file", defaultSchemaFile, "Schema file to validate against")
	flags.BoolVar(&opts.schemaStrict, "schema-strict", false, "Warn on environment variables not present in schema")
	flags.StringVar(&opts.format, "format", exporter.FormatShell, "Export format for dry run (shell or json)")
	flags.BoolVar(&opts.unredacted, "unredacted", false, "Show unredacted values in dry run output")
	flags.StringSliceVar(&opts.maskPatterns, "mask-pattern", nil, "Additional regex patterns to mask by value")

	return cmd
}

func runRun(cmd *cobra.Command, opts *runOptions, args []string) error {
	cfg, migrated, err := config.LoadWithMigration()

	flags := cmd.Flags()
	explicitEnvFiles := flags.Changed("env-file")
	if !flags.Changed("env-file") {
		opts.envFiles = []string{cfg.Defaults.EnvFile}
	}
	if !flags.Changed("cascade") {
		opts.cascade = cfg.Defaults.Cascade
	}
	if !flags.Changed("dry-run") {
		opts.dryRun = cfg.Defaults.DryRun
	}
	if !flags.Changed("validate") {
		opts.validate = cfg.Validation.Enabled
	}
	if !flags.Changed("schema") && !flags.Changed("schema-file") {
		opts.schemaFile = cfg.Validation.SchemaFile
	}
	if !flags.Changed("schema-strict") {
		opts.schemaStrict = cfg.Validation.Strict
	}

	if level := verbosityLevel(); level > 0 {
		cfg.General.Verbosity = level
	}

	warnCascade := explicitEnvFiles && opts.cascade
	if warnCascade {
		opts.cascade = false
	}

	verboseOutput := cfg.General.Verbosity >= 2
	if verboseOutput {
		fmt.Fprintf(os.Stderr, "Loading config from: %s\n", configPath())
	}
	if cfg.General.Verbosity >= 1 && migrated {
		fmt.Fprintf(os.Stderr, "Successfully migrated ~/.nv to config\n")
	}
	env, err := loadEnvironment(envOptions{
		envFiles:  opts.envFiles,
		cascade:   opts.cascade,
		env:       opts.env,
		overrides: opts.overrides,
		strict:    opts.strict,
		preserve:  opts.preserve,
		globals:   cfg.GetGlobalEnv(),
		priority:  cfg.Globals.Priority,
		autoLocal: !flags.Changed("env-file") && cfg.Defaults.AutoLocal,
		verbose:   cfg.General.Verbosity >= 2,
		trace:     cfg.General.Verbosity >= 2,
	})
	if warnCascade {
		if verboseOutput {
			fmt.Fprintln(os.Stderr)
		}
		fmt.Fprintf(os.Stderr, "warning: --env-file provided; disabling --cascade and using only explicit env files\n")
		fmt.Fprintln(os.Stderr)
	}
	if err != nil {
		return err
	}

	if opts.validate {
		if err := validateEnvironment(env, validationOptions{
			schemaFile: opts.schemaFile,
			strict:     opts.schemaStrict,
			envFiles:   opts.envFiles,
		}); err != nil {
			return err
		}
	}

	if opts.dryRun {
		return exporter.Write(os.Stdout, env, exporter.Options{
			Format:       opts.format,
			Unredacted:   opts.unredacted,
			MaskPatterns: opts.maskPatterns,
		})
	}

	runner := exec.NewRunner(env, args[0], args[1:])
	exitCode, err := runner.Run()
	if err != nil {
		return err
	}

	return exitError{code: exitCode}
}

func configPath() string {
	exists, err := config.ConfigExists()
	if err != nil {
		return "unknown"
	}
	if !exists {
		return "-"
	}

	path, err := config.GetConfigPath()
	if err != nil {
		return "unknown"
	}
	return path
}
