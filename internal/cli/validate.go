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
	"path/filepath"

	"github.com/fatih/color"
	"github.com/jcouture/nv/internal/validator"
	"github.com/spf13/cobra"
)

type validateOptions struct {
	envFiles     []string
	cascade      bool
	env          string
	overrides    []string
	strict       bool
	preserve     []string
	schemaFile   string
	schemaStrict bool
	verbose      bool
}

type validationOptions struct {
	schemaFile string
	strict     bool
	verbose    bool
	envFiles   []string
}

func newValidateCmd() *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "validate [flags]",
		Short: "Validate environment variables against a schema",
		Long:  "Load environment variables from .env files and validate them against a schema file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(opts)
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&opts.envFiles, "env-file", "e", []string{defaultEnvFile}, "Environment file(s) to load")
	flags.BoolVarP(&opts.cascade, "cascade", "c", false, "Enable cascading mode")
	flags.StringVar(&opts.env, "env", "", "Environment name for cascading")
	flags.StringSliceVarP(&opts.overrides, "override", "o", nil, "Inline overrides (KEY=value)")
	flags.BoolVar(&opts.strict, "strict", false, "Fail on unresolved interpolation")
	flags.StringSliceVarP(&opts.preserve, "preserve", "p", []string{"PATH", "HOME", "USER"}, "System variables to preserve")
	flags.StringVar(&opts.schemaFile, "schema", defaultSchemaFile, "Schema file to validate against")
	flags.StringVar(&opts.schemaFile, "schema-file", defaultSchemaFile, "Schema file to validate against")
	flags.BoolVar(&opts.schemaStrict, "schema-strict", false, "Warn on environment variables not present in schema")

	return cmd
}

func runValidate(opts *validateOptions) error {
	if level := verbosityLevel(); level >= 2 {
		opts.verbose = true
	}
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

	return validateEnvironment(env, validationOptions{
		schemaFile: opts.schemaFile,
		strict:     opts.schemaStrict,
		verbose:    opts.verbose,
		envFiles:   opts.envFiles,
	})
}

func validateEnvironment(env map[string]string, opts validationOptions) error {
	result, err := validator.Validate(opts.schemaFile, env, validator.Options{Strict: opts.strict})
	if err != nil {
		_, _ = color.New(color.FgRed).Fprintf(os.Stderr, "Validation failed: %s\n", err)
		return exitError{code: 1}
	}

	if result.EmptySchema {
		_, _ = color.New(color.FgYellow).Fprintf(os.Stderr, "Validation skipped: schema file is empty (%s)\n", result.SchemaPath)
	}

	if hasCircularSchema(opts.envFiles, result.SchemaPath) {
		_, _ = color.New(color.FgYellow).Fprintf(os.Stderr, "Validation warning: schema file matches an env file (%s)\n", result.SchemaPath)
	}

	if len(result.Missing) > 0 {
		_, _ = color.New(color.FgRed).Fprintln(os.Stderr, "Validation failed: Missing required environment variables:")
		for _, key := range result.Missing {
			fmt.Fprintf(os.Stderr, "  - %s (defined in %s)\n", key, result.SchemaPath)
		}
		return exitError{code: 1}
	}

	if opts.strict && len(result.Extra) > 0 {
		_, _ = color.New(color.FgYellow).Fprintln(os.Stderr, "Validation warning: Environment variables not present in schema:")
		for _, key := range result.Extra {
			fmt.Fprintf(os.Stderr, "  - %s\n", key)
		}
	}

	if opts.verbose {
		color.Green("Validation passed")
	}

	return nil
}

func hasCircularSchema(envFiles []string, schemaPath string) bool {
	if schemaPath == "" || len(envFiles) == 0 {
		return false
	}

	schemaAbs, err := filepath.Abs(schemaPath)
	if err != nil {
		schemaAbs = filepath.Clean(schemaPath)
	}

	for _, file := range envFiles {
		fileAbs, err := filepath.Abs(file)
		if err != nil {
			fileAbs = filepath.Clean(file)
		}
		if fileAbs == schemaAbs {
			return true
		}
	}

	return false
}
