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

package config

import (
	"fmt"
	"strings"
)

func (c *Config) Validate() []error {
	var errs []error

	if err := ValidateGlobalsPriority(c.Globals.Priority); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateVerbosity(c.General.Verbosity); err != nil {
		errs = append(errs, err)
	}
	if envErrs := ValidateGlobalEnvKeys(c.Globals.Env); len(envErrs) > 0 {
		errs = append(errs, envErrs...)
	}

	return errs
}

func ValidateGlobalsPriority(priority string) error {
	switch priority {
	case GlobalsPriorityFirst, GlobalsPriorityLast:
		return nil
	default:
		return fmt.Errorf("invalid globals priority: %s", priority)
	}
}

func ValidateVerbosity(level int) error {
	if level < 0 || level > 2 {
		return fmt.Errorf("invalid verbosity level: %d", level)
	}
	return nil
}

func ValidateGlobalEnvKeys(env map[string]string) []error {
	var errs []error
	for key := range env {
		if strings.TrimSpace(key) == "" {
			errs = append(errs, fmt.Errorf("global env key cannot be empty"))
			continue
		}
		if strings.Contains(key, "=") {
			errs = append(errs, fmt.Errorf("global env key %q contains '='", key))
			continue
		}
		if strings.ContainsAny(key, " \t\n") {
			errs = append(errs, fmt.Errorf("global env key %q contains whitespace", key))
		}
	}
	return errs
}

func (c *Config) Fix() (*Config, []string) {
	fixed := *c
	var fields []string
	defaults := Default()

	if err := ValidateGlobalsPriority(fixed.Globals.Priority); err != nil {
		fixed.Globals.Priority = defaults.Globals.Priority
		fields = append(fields, "globals.priority")
	}
	if err := ValidateVerbosity(fixed.General.Verbosity); err != nil {
		fixed.General.Verbosity = defaults.General.Verbosity
		fields = append(fields, "general.verbosity")
	}

	if envErrs := ValidateGlobalEnvKeys(fixed.Globals.Env); len(envErrs) > 0 {
		for key := range fixed.Globals.Env {
			if len(ValidateGlobalEnvKeys(map[string]string{key: fixed.Globals.Env[key]})) > 0 {
				delete(fixed.Globals.Env, key)
				fields = append(fields, fmt.Sprintf("globals.env.%s", key))
			}
		}
	}

	if fixed.Globals.Env == nil {
		fixed.Globals.Env = map[string]string{}
	}

	return &fixed, fields
}
