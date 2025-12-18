// Copyright 2015-2023 Jean-Philippe Couture
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

package sys

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadVarsFromFile(t *testing.T) {
	cases := []struct {
		description string
		input       string
		expected    map[string]string
		wantErr     bool
	}{
		{"Simple file", "testdata/.env", map[string]string{"PORT": "4200", "SECRET_KEY": "1234567890"}, false},
		{"Empty file", "testdata/empty.env", map[string]string{}, false},
		{"Non-existent file", "testdata/non-existent.env", nil, true},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			result, err := ReadVarsFromFile(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCurrentEnv(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "captures plain variable",
			env:  map[string]string{"NV_TEST_CURRENT": "value"},
		},
		{
			name: "preserves equals in value",
			env:  map[string]string{"NV_TEST_EQUALS": "foo=bar=baz"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			got := currentEnv()
			for k, v := range tc.env {
				assert.Equal(t, v, got[k])
			}
		})
	}
}

func TestReadGlobalVars(t *testing.T) {
	cases := []struct {
		name        string
		fileContent string
		env         map[string]string
		expect      map[string]string
	}{
		{
			name:        "reads globals using HOME and existing env",
			fileContent: "FOO=bar\nFROM_ENV=${NV_HOME_VAR}",
			env:         map[string]string{"NV_HOME_VAR": "value-from-env"},
			expect:      map[string]string{"FOO": "bar", "FROM_ENV": "value-from-env"},
		},
		{
			name:   "missing file returns nil map",
			expect: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tempHome := t.TempDir()
			t.Setenv("HOME", tempHome)

			if tc.fileContent != "" {
				nvPath := tempHome + "/.nv"
				assert.NoError(t, os.WriteFile(nvPath, []byte(tc.fileContent), 0o600))
			}

			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			got := ReadGlobalVars()
			assert.Equal(t, tc.expect, got)
		})
	}
}
