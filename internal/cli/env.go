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
	"maps"
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
	"github.com/jcouture/nv/internal/config"
	"github.com/jcouture/nv/internal/loader"
)

type envOptions struct {
	envFiles  []string
	cascade   bool
	env       string
	overrides []string
	strict    bool
	preserve  []string
	globals   map[string]string
	priority  string
	autoLocal bool
	verbose   bool
	trace     bool
}

func loadEnvironment(opts envOptions) (map[string]string, error) {
	loaderOpts := []loader.Option{
		loader.WithPreserve(opts.preserve),
		loader.WithStrict(opts.strict),
	}

	if opts.verbose || opts.trace {
		loaderOpts = append(loaderOpts, loader.WithTracer(traceLoaderEvent))
	}

	l := loader.New(loaderOpts...)

	var env map[string]string
	var err error

	baseEnv := l.PreservedEnv()
	if opts.trace && opts.priority == config.GlobalsPriorityFirst {
		traceGlobals(baseEnv, opts.globals)
	}
	if opts.priority == config.GlobalsPriorityFirst {
		mergeEnv(baseEnv, opts.globals)
	}

	if opts.cascade {
		env, err = l.LoadCascadeWithEnv(opts.env, baseEnv)
	} else {
		files := opts.envFiles
		if opts.autoLocal {
			autoLocal := autoLocalForFiles(files)
			env, err = l.LoadFilesWithEnv(baseEnv, files...)
			if err == nil && autoLocal != "" {
				env, err = l.LoadOptionalFilesWithEnv(env, autoLocal)
			}
		} else {
			env, err = l.LoadFilesWithEnv(baseEnv, files...)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	if opts.priority == config.GlobalsPriorityLast {
		if opts.trace {
			traceGlobals(env, opts.globals)
		}
		mergeEnv(env, opts.globals)
	}

	type overrideChange struct {
		key    string
		exists bool
	}
	var changes []overrideChange
	for _, override := range opts.overrides {
		key, value, err := parseOverride(override)
		if err != nil {
			return nil, fmt.Errorf("invalid override %q: %w", override, err)
		}
		_, exists := env[key]
		env[key] = value
		changes = append(changes, overrideChange{key: key, exists: exists})
	}

	if (opts.verbose || opts.trace) && len(changes) > 0 {
		_, _ = color.New(color.FgHiBlue).Fprintln(os.Stderr, "Applying overrides:")
		for _, change := range changes {
			if change.exists {
				_, _ = color.New(color.FgYellow).Fprintf(os.Stderr, "  %s overridden\n", change.key)
				continue
			}
			_, _ = color.New(color.FgGreen).Fprintf(os.Stderr, "  %s added\n", change.key)
		}
	}

	return env, nil
}

func traceLoaderEvent(event loader.TraceEvent) {
	switch event.Status {
	case "missing":
		_, _ = color.New(color.FgYellow).Fprintf(os.Stderr, "Skipping missing env file: %s\n", event.File)
	case "loaded":
		_, _ = color.New(color.FgHiCyan).Fprintf(os.Stderr, "Loaded env file: %s\n", event.File)
		for key := range event.Overwritten {
			_, _ = color.New(color.FgYellow).Fprintf(os.Stderr, "  %s overridden\n", key)
		}
		for key := range event.Added {
			_, _ = color.New(color.FgGreen).Fprintf(os.Stderr, "  %s added\n", key)
		}
	}
}

func traceGlobals(env map[string]string, globals map[string]string) {
	if len(globals) == 0 {
		return
	}

	keys := make([]string, 0, len(globals))
	for key := range globals {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Fprintln(os.Stderr, "Loading global env:")
	for _, key := range keys {
		val := globals[key]
		if prev, ok := env[key]; ok {
			if prev != val {
				_, _ = color.New(color.FgYellow).Fprintf(os.Stderr, "  %s overridden\n", key)
			}
			continue
		}
		_, _ = color.New(color.FgGreen).Fprintf(os.Stderr, "  %s added\n", key)
	}
}

func mergeEnv(dst map[string]string, src map[string]string) {
	maps.Copy(dst, src)
}

func autoLocalForFiles(files []string) string {
	hasLocal := false
	for _, file := range files {
		if filepath.Base(file) == ".env.local" {
			hasLocal = true
			break
		}
	}
	for _, file := range files {
		if filepath.Base(file) == ".env" {
			if hasLocal {
				return ""
			}
			return filepath.Join(filepath.Dir(file), ".env.local")
		}
	}
	return ""
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
