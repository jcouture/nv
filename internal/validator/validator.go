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

package validator

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/jcouture/nv/internal/parser"
)

type Options struct {
	Strict bool
}

type Result struct {
	SchemaPath  string
	Missing     []string
	Extra       []string
	Required    []string
	EmptySchema bool
}

var requiredCommentPattern = regexp.MustCompile(`(?i)^\s*#\s*REQUIRED:\s*([A-Za-z_][A-Za-z0-9_]*)\s*$`)

func Validate(schemaPath string, env map[string]string, opts Options) (Result, error) {
	result := Result{SchemaPath: schemaPath}
	if schemaPath == "" {
		return result, fmt.Errorf("schema file path is required")
	}

	// #nosec G304 - schema file path is provided explicitly by the user/CLI.
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return result, fmt.Errorf("schema file '%s': %w", schemaPath, err)
	}

	requiredFromComments := parseRequiredComments(strings.NewReader(string(data)))
	schemaVars, err := parser.ParseFile(schemaPath)
	if err != nil {
		return result, fmt.Errorf("schema parse failed: %w", err)
	}

	requiredKeys := make(map[string]struct{})
	for key, val := range schemaVars {
		if val != "" {
			requiredKeys[key] = struct{}{}
		}
	}
	for _, key := range requiredFromComments {
		requiredKeys[key] = struct{}{}
	}

	if len(schemaVars) == 0 && len(requiredKeys) == 0 {
		result.EmptySchema = true
	}

	result.Required = sortedKeysFromSet(requiredKeys)

	for key := range requiredKeys {
		val, ok := env[key]
		if !ok || val == "" {
			result.Missing = append(result.Missing, key)
		}
	}
	sort.Strings(result.Missing)

	if opts.Strict {
		for key := range env {
			if _, ok := schemaVars[key]; ok {
				continue
			}
			if _, ok := requiredKeys[key]; ok {
				continue
			}
			result.Extra = append(result.Extra, key)
		}
		sort.Strings(result.Extra)
	}

	return result, nil
}

func parseRequiredComments(r io.Reader) []string {
	var keys []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		matches := requiredCommentPattern.FindStringSubmatch(line)
		if len(matches) == 2 {
			keys = append(keys, matches[1])
		}
	}
	return keys
}

func sortedKeysFromSet(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
