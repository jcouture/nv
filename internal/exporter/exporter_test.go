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

package exporter

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestWriteShellMasked(t *testing.T) {
	color.NoColor = true
	env := map[string]string{
		"API_KEY":   "secret",
		"DEBUG":     "true",
		"PLAINTEXT": "-----BEGIN PRIVATE KEY-----\nTEST\n-----END PRIVATE KEY-----",
		"NOTE":      "AIzaAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}

	var buf bytes.Buffer
	if err := Write(&buf, env, Options{Format: FormatShell}); err != nil {
		t.Fatalf("write shell: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "export API_KEY=\""+redactedValue+"\"") {
		t.Fatalf("expected API_KEY to be masked, got: %s", output)
	}
	if !strings.Contains(output, "export DEBUG=\"true\"") {
		t.Fatalf("expected DEBUG to be present, got: %s", output)
	}
	if !strings.Contains(output, "export PLAINTEXT=\""+redactedValue+"\"") {
		t.Fatalf("expected PLAINTEXT to be masked by value, got: %s", output)
	}
	if !strings.Contains(output, "export NOTE=\""+redactedValue+"\"") {
		t.Fatalf("expected NOTE to be masked by value pattern, got: %s", output)
	}
}

func TestWriteJSONUnredacted(t *testing.T) {
	color.NoColor = true
	env := map[string]string{
		"API_KEY": "secret",
		"DEBUG":   "true",
	}

	var buf bytes.Buffer
	if err := Write(&buf, env, Options{Format: FormatJSON, Unredacted: true}); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v", err)
	}

	if payload["API_KEY"] != "secret" {
		t.Fatalf("expected unredacted API_KEY, got: %s", payload["API_KEY"])
	}
	if payload["DEBUG"] != "true" {
		t.Fatalf("expected DEBUG to be true, got: %s", payload["DEBUG"])
	}
}

func TestWriteWithCustomMaskPattern(t *testing.T) {
	color.NoColor = true
	env := map[string]string{
		"NOTE": "FOO12345",
	}

	var buf bytes.Buffer
	if err := Write(&buf, env, Options{Format: FormatShell, MaskPatterns: []string{"FOO[0-9]+"}}); err != nil {
		t.Fatalf("write shell: %v", err)
	}

	if !strings.Contains(buf.String(), "export NOTE=\""+redactedValue+"\"") {
		t.Fatalf("expected NOTE to be masked by custom pattern, got: %s", buf.String())
	}
}

func TestWriteUnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, map[string]string{}, Options{Format: "unknown"})
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestWriteInvalidMaskPattern(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, map[string]string{}, Options{Format: FormatShell, MaskPatterns: []string{"("}})
	if err == nil {
		t.Fatal("expected error for invalid mask pattern")
	}
}

func TestWriteShellUnredacted(t *testing.T) {
	color.NoColor = true
	env := map[string]string{
		"API_KEY": "secret",
	}

	var buf bytes.Buffer
	if err := Write(&buf, env, Options{Format: FormatShell, Unredacted: true}); err != nil {
		t.Fatalf("write shell: %v", err)
	}

	if strings.Contains(buf.String(), redactedValue) {
		t.Fatalf("expected unredacted output, got: %s", buf.String())
	}
}

func TestWriteDefaultsToShellFormat(t *testing.T) {
	origNoColor := color.NoColor
	t.Cleanup(func() { color.NoColor = origNoColor })
	color.NoColor = true

	env := map[string]string{
		"FOO": "bar",
	}
	var buf bytes.Buffer
	if err := Write(&buf, env, Options{}); err != nil {
		t.Fatalf("write default format: %v", err)
	}
	if !strings.Contains(buf.String(), "export FOO=\"bar\"") {
		t.Fatalf("expected shell output, got: %s", buf.String())
	}
}

func TestWriteShellReturnsWriteError(t *testing.T) {
	writerErr := errWriter{err: errWriterSentinel}
	err := Write(&writerErr, map[string]string{"FOO": "bar"}, Options{Format: FormatShell})
	if err == nil {
		t.Fatal("expected write error")
	}
}

func TestWriteShellMaskedWithColor(t *testing.T) {
	origNoColor := color.NoColor
	t.Cleanup(func() { color.NoColor = origNoColor })
	color.NoColor = false

	env := map[string]string{
		"API_KEY": "secret",
	}
	var buf bytes.Buffer
	if err := Write(&buf, env, Options{Format: FormatShell}); err != nil {
		t.Fatalf("write shell: %v", err)
	}
	if !strings.Contains(buf.String(), redactedValue) {
		t.Fatalf("expected masked output, got: %s", buf.String())
	}
}

func TestEscapeShellValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "quotes and newline",
			input: "hello \"world\"\nnext",
			want:  "hello \\\"world\\\"\\nnext",
		},
		{
			name:  "backslash and dollar",
			input: `path\to\$HOME`,
			want:  `path\\to\\\$HOME`,
		},
		{
			name:  "backtick and tab",
			input: "foo`\tbar",
			want:  "foo\\`\\tbar",
		},
		{
			name:  "carriage return",
			input: "line\rend",
			want:  "line\\rend",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := escapeShellValue(tc.input)
			if got != tc.want {
				t.Fatalf("escapeShellValue = %q, want %q", got, tc.want)
			}
		})
	}
}

var errWriterSentinel = errors.New("write failed")

type errWriter struct {
	err error
}

func (w *errWriter) Write(p []byte) (int, error) {
	return 0, w.err
}
