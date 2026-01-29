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

package env

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Clear removes all environment variables except the provided list.
func Clear(except ...string) error {
	var errs []error
	for _, name := range Getnames(Getvars()) {
		if !contains(except, name) {
			if err := os.Unsetenv(name); err != nil {
				errs = append(errs, fmt.Errorf("unset %s: %w", name, err))
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Exists reports whether the variable is present in the environment.
func Exists(name string) bool {
	_, ok := os.LookupEnv(name)

	return ok
}

// Getvars returns the current environment as a map.
func Getvars() map[string]string {
	vars := make(map[string]string)

	for _, line := range os.Environ() {
		result := strings.Split(line, "=")
		if len(result) >= 2 {
			n := result[0]
			v := strings.Join(result[1:], "=")
			vars[n] = v
		}
	}

	return vars
}

// Getnames returns the list of environment variable names contained in the map.
func Getnames(vars map[string]string) []string {
	names := make([]string, 0)

	for n := range vars {
		if len(n) > 0 {
			names = append(names, n)
		}
	}

	return names
}

// Join merges the override map into base, returning the merged map.
func Join(base map[string]string, override map[string]string) map[string]string {
	if len(base) == 0 {
		return override
	}
	for k, v := range override {
		base[k] = v
	}

	return base
}

// Setvars sets environment variables from the provided map.
// Returns on the first error encountered to avoid partial silent failures.
func Setvars(vars map[string]string) error {
	for k, v := range vars {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
