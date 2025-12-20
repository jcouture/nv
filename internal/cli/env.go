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
	"path/filepath"

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
}

func loadEnvironment(opts envOptions) (map[string]string, error) {
	loaderOpts := []loader.Option{
		loader.WithPreserve(opts.preserve),
		loader.WithStrict(opts.strict),
	}

	l := loader.New(loaderOpts...)

	var env map[string]string
	var err error

	baseEnv := l.PreservedEnv()
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
		mergeEnv(env, opts.globals)
	}

	for _, override := range opts.overrides {
		key, value, err := parseOverride(override)
		if err != nil {
			return nil, fmt.Errorf("invalid override %q: %w", override, err)
		}
		env[key] = value
	}

	return env, nil
}

func mergeEnv(dst map[string]string, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
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
