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

package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name        string
		schema      string
		env         map[string]string
		opts        Options
		wantErr     bool
		wantMissing []string
		wantExtra   []string
		wantEmpty   bool
	}{
		{
			name:        "missing required key with value and required comment",
			schema:      "DATABASE_URL=postgres://localhost\nOPTIONAL=\n# REQUIRED: API_KEY\n",
			env:         map[string]string{"DATABASE_URL": ""},
			wantMissing: []string{"API_KEY", "DATABASE_URL"},
		},
		{
			name:   "optional key absent does not count as missing",
			schema: "DATABASE_URL=postgres://localhost\nOPTIONAL=\n",
			env:    map[string]string{"DATABASE_URL": "postgres://localhost"},
		},
		{
			name:      "strict mode reports extra keys",
			schema:    "DATABASE_URL=postgres://localhost\n",
			env:       map[string]string{"DATABASE_URL": "postgres://localhost", "EXTRA": "1"},
			opts:      Options{Strict: true},
			wantExtra: []string{"EXTRA"},
		},
		{
			name:      "empty schema",
			schema:    "",
			env:       map[string]string{},
			wantEmpty: true,
		},
		{
			name:    "empty schema path returns error",
			wantErr: true,
		},
		{
			name:    "missing schema file returns error",
			wantErr: true,
		},
		{
			name:    "oversized schema line returns error",
			schema:  "API_KEY=" + strings.Repeat("A", maxSchemaLineSize+1) + "\n",
			wantErr: true,
		},
		{
			name:        "only required comments, no key=value lines",
			schema:      "# REQUIRED: FOO\n# REQUIRED: BAR\n",
			env:         map[string]string{"FOO": "set"},
			wantMissing: []string{"BAR"},
			wantEmpty:   false,
		},
		{
			name:   "strict mode: required key present in env is not reported as extra",
			schema: "# REQUIRED: FOO\n",
			env:    map[string]string{"FOO": "value"},
			opts:   Options{Strict: true},
		},
		{
			name:   "all required keys satisfied",
			schema: "FOO=default\n# REQUIRED: BAR\n",
			env:    map[string]string{"FOO": "val", "BAR": "val"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			schemaPath := ""

			if tc.name == "empty schema path returns error" {
				_, err := Validate("", map[string]string{}, Options{})
				if err == nil {
					t.Fatal("expected error for empty schema path")
				}
				return
			}

			if tc.name == "missing schema file returns error" {
				_, err := Validate("does-not-exist.env", map[string]string{}, Options{})
				if err == nil {
					t.Fatal("expected error for missing schema file")
				}
				return
			}

			tmpDir := t.TempDir()
			schemaPath = filepath.Join(tmpDir, ".env.example")
			if err := os.WriteFile(schemaPath, []byte(tc.schema), 0o644); err != nil {
				t.Fatalf("write schema: %v", err)
			}

			result, err := Validate(schemaPath, tc.env, tc.opts)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.EmptySchema != tc.wantEmpty {
				t.Fatalf("EmptySchema = %v, want %v", result.EmptySchema, tc.wantEmpty)
			}

			got := result.Missing
			if len(got) != len(tc.wantMissing) {
				t.Fatalf("Missing = %v, want %v", got, tc.wantMissing)
			}
			for i, k := range tc.wantMissing {
				if got[i] != k {
					t.Fatalf("Missing[%d] = %q, want %q", i, got[i], k)
				}
			}

			gotExtra := result.Extra
			if len(gotExtra) != len(tc.wantExtra) {
				t.Fatalf("Extra = %v, want %v", gotExtra, tc.wantExtra)
			}
			for i, k := range tc.wantExtra {
				if gotExtra[i] != k {
					t.Fatalf("Extra[%d] = %q, want %q", i, gotExtra[i], k)
				}
			}
		})
	}
}
