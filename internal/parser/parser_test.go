// Copyright 2015-2022 Jean-Philippe Couture
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	cases := []struct {
		description string
		input       string
		expected    bool
	}{
		{"File exists", "testdata/.env", true},
		{"File does not exist", "testdata/.env2", false},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			p := NewParser(tt.input)
			result := p.fileExists()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadFile(t *testing.T) {
	cases := []struct {
		description string
		input       string
		expected    string
	}{
		{"Standard .env file", "testdata/.env", "PORT=4200\nSECRET_KEY=1234567\nDATABASE_URL=postgres://simonprev:@localhost:5432/accent_playground_dev?pool=10\n"},
		{"File does not exist", "testdata/.env2", ""},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			p := NewParser(tt.input)
			result, err := p.readFile()
			if err != nil {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadLines(t *testing.T) {
	cases := []struct {
		description string
		input       string
		expected    []string
	}{
		{"A couple of lines", "LINE1\nLINE2", []string{"LINE1", "LINE2"}},
		{"Non empty lines with spaces", "  LINE1  \n  LINE2  ", []string{"LINE1", "LINE2"}},
		{"Empty lines", "\n\n", []string{}},
		{"Empty lines with spaces", "  \n  \n", []string{}},
		{"Empty lines with tabs", "\t\t\n\t\t\n", []string{}},
		{"Empty lines with spaces and tabs", "  \t\n  \t\n", []string{}},
		{"Lines, with a comment", "LINE1\n#LINE2\nLINE3", []string{"LINE1", "LINE3"}},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			result := readLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractVariables(t *testing.T) {
	cases := []struct {
		description string
		input       []string
		expected    map[string]string
	}{
		{"Key and value", []string{"PORT=4200"}, map[string]string{"PORT": "4200"}},
		{"Key and value with spaces", []string{"PORT = 4200"}, map[string]string{"PORT ": " 4200"}},
		{"Key and value with tabs", []string{"PORT\t=\t4200"}, map[string]string{"PORT\t": "\t4200"}},
		{"Key and value with spaces and tabs", []string{"PORT \t = \t 4200"}, map[string]string{"PORT \t ": " \t 4200"}},
		{"Key and value with extra equal sign", []string{"PORT=42=00"}, map[string]string{"PORT": "42=00"}},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			result := extractVariables(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
