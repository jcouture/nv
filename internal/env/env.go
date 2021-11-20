// Copyright 2015-2021 Jean-Philippe Couture
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
	"os"

	"github.com/jcouture/nv/internal/parser"
)

func Set(vars map[string]string) {
	for k, v := range vars {
		os.Setenv(k, v)
	}
}

func Join(base map[string]string, override map[string]string) map[string]string {
	if len(base) == 0 {
		return override
	}
	for k, v := range override {
		base[k] = v
	}

	return base
}

func Load(fn string) (map[string]string, error) {
	parser := parser.NewParser(fn)
	return parser.Parse()
}

func Clear() {
	// Clearing everything out the environment... except $PATH (we’re savages)!
	path := os.Getenv("PATH")
	os.Clearenv()
	os.Setenv("PATH", path)
}
