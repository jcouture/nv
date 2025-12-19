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

package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/color"
)

const (
	FormatShell = "shell"
	FormatJSON  = "json"
)

func writeShell(w io.Writer, env map[string]string, opts Options, patterns []*regexp.Regexp) error {
	keys := sortedKeys(env)
	maskColor := color.New(color.FgYellow, color.Faint)

	for _, key := range keys {
		value, masked := maskValue(key, env[key], opts.Unredacted, patterns)
		escaped := escapeShellValue(value)
		if masked && opts.Color {
			escaped = maskColor.Sprint(escaped)
		}
		if _, err := fmt.Fprintf(w, "export %s=\"%s\"\n", key, escaped); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(w io.Writer, env map[string]string, opts Options, patterns []*regexp.Regexp) error {
	payload := make(map[string]string, len(env))
	for key, val := range env {
		masked, _ := maskValue(key, val, opts.Unredacted, patterns)
		payload[key] = masked
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func escapeShellValue(value string) string {
	var builder strings.Builder
	for _, r := range value {
		switch r {
		case '\\':
			builder.WriteString(`\\`)
		case '"':
			builder.WriteString(`\"`)
		case '$':
			builder.WriteString(`\$`)
		case '`':
			builder.WriteString("\\`")
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func sortedKeys(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
