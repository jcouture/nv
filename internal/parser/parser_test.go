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

package parser

import (
	"os"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		env      map[string]string
		strict   bool
		expected map[string]string
		wantErr  bool
	}{
		{name: "simple assignment", input: "FOO=bar", expected: map[string]string{"FOO": "bar"}},
		{name: "multiple assignments", input: "FOO=bar\nBAZ=qux", expected: map[string]string{"FOO": "bar", "BAZ": "qux"}},
		{name: "export prefix", input: "export FOO=bar", expected: map[string]string{"FOO": "bar"}},
		{name: "full line comment", input: "# comment\nFOO=bar", expected: map[string]string{"FOO": "bar"}},
		{name: "inline comment", input: "FOO=bar # comment", expected: map[string]string{"FOO": "bar"}},
		{name: "inline comment not in quotes", input: `FOO="bar # not a comment"`, expected: map[string]string{"FOO": "bar # not a comment"}},
		{name: "single quotes literal", input: `FOO='hello\nworld'`, expected: map[string]string{"FOO": `hello\nworld`}},
		{name: "double quotes with escape", input: "FOO=\"hello\\nworld\"", expected: map[string]string{"FOO": "hello\nworld"}},
		{name: "double quotes with escaped quote", input: "FOO=\"say \\\"hello\\\"\"", expected: map[string]string{"FOO": `say "hello"`}},
		{name: "double quotes unknown escape passthrough", input: "FOO=\"hi\\q\"", expected: map[string]string{"FOO": "hiq"}},
		{name: "double quotes with slash and control", input: "FOO=\"a\\\\b\\rc\\t\"", expected: map[string]string{"FOO": "a\\b\rc\t"}},
		{name: "multiline double quotes", input: "FOO=\"line1\nline2\"", expected: map[string]string{"FOO": "line1\nline2"}},
		{name: "multiline single quotes", input: "FOO='line1\nline2'", expected: map[string]string{"FOO": "line1\nline2"}},
		{name: "interpolation braces", input: "HOST=localhost\nURL=http://${HOST}", expected: map[string]string{"HOST": "localhost", "URL": "http://localhost"}},
		{name: "interpolation no braces", input: "HOST=localhost\nURL=http://$HOST", expected: map[string]string{"HOST": "localhost", "URL": "http://localhost"}},
		{name: "interpolation from existing env", input: "URL=http://${HOST}", env: map[string]string{"HOST": "example.com"}, expected: map[string]string{"URL": "http://example.com"}},
		{name: "no interpolation in single quotes", input: "HOST=localhost\nURL='http://${HOST}'", expected: map[string]string{"HOST": "localhost", "URL": "http://${HOST}"}},
		{name: "unresolved interpolation becomes empty", input: "URL=http://${UNDEFINED}", expected: map[string]string{"URL": "http://"}},
		{name: "path prepend", input: "PATH=./bin:${PATH}", env: map[string]string{"PATH": "/usr/bin"}, expected: map[string]string{"PATH": "./bin:/usr/bin"}},
		{name: "empty value", input: "FOO=", expected: map[string]string{"FOO": ""}},
		{name: "empty quoted value", input: `FOO=""`, expected: map[string]string{"FOO": ""}},
		{name: "value with equals sign", input: "FOO=bar=baz", expected: map[string]string{"FOO": "bar=baz"}},
		{name: "whitespace around equals", input: "FOO = bar", expected: map[string]string{"FOO": "bar"}},
		{name: "leading/trailing whitespace in value", input: `FOO="  spaced  "`, expected: map[string]string{"FOO": "  spaced  "}},
		{name: "strict interpolation error", input: "FOO=$MISSING", strict: true, wantErr: true},
		{name: "unterminated interpolation strict", input: "FOO=${BAR", strict: true, wantErr: true},
		{name: "missing equals", input: "FOO\nBAR=baz", wantErr: true},
		{name: "export without key", input: "export", wantErr: true},
		{name: "empty value followed by newline", input: "FOO=\nBAR=baz", expected: map[string]string{"FOO": "", "BAR": "baz"}},
		{name: "empty value with inline comment", input: "FOO=   # comment", expected: map[string]string{"FOO": ""}},
		{name: "unexpected token skipped", input: "=\nFOO=bar", expected: map[string]string{"FOO": "bar"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := []Option{}
			if tc.env != nil {
				opts = append(opts, WithExistingEnv(tc.env))
			}
			if tc.strict {
				opts = append(opts, WithStrictInterpolation())
			}
			p := NewParser(NewLexer(tc.input), opts...)
			got, err := p.Parse()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	t.Run("missing file returns error", func(t *testing.T) {
		_, err := ParseFile("testdata/does-not-exist.env")
		if err == nil {
			t.Fatalf("expected error for missing file")
		}
	})

	t.Run("parse file", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "parser-file-*.env")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmp.Name())

		content := "FOO=bar\nBAR=baz"
		if _, err := tmp.WriteString(content); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
		_ = tmp.Close()

		got, err := ParseFile(tmp.Name())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := map[string]string{"FOO": "bar", "BAR": "baz"}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})
}

func TestUnterminatedQuotes(t *testing.T) {
	_, err := NewParser(NewLexer(`FOO="unterminated`)).Parse()
	if err == nil {
		t.Fatalf("expected error for unterminated double quote")
	}

	_, err = NewParser(NewLexer("FOO='unterminated")).Parse()
	if err == nil {
		t.Fatalf("expected error for unterminated single quote")
	}
}

func TestInterpolateValueEdgeCases(t *testing.T) {
	got, err := interpolateValue("foo$", "KEY", modeUnquoted, map[string]string{}, map[string]string{}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "foo$" {
		t.Fatalf("expected foo$, got %q", got)
	}

	got, err = interpolateValue("${FOO", "KEY", modeDoubleQuoted, map[string]string{}, map[string]string{}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "${FOO" {
		t.Fatalf("expected literal ${FOO, got %q", got)
	}
}
