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
	"testing"
)

func TestSet(t *testing.T) {
	name := "FOO"
	expected := "BAR"

	vars := make(map[string]string)
	vars[name] = "BAR"
	Set(vars)

	result := os.Getenv(name)
	if result != expected {
		t.Errorf("Expected: %s, got: %s\n", expected, result)
	}
}

func TestJoin(t *testing.T) {
	base := make(map[string]string)
	base["FOO"] = "BAR"

	override := make(map[string]string)
	override["COLOR"] = "RED"

	result := Join(base, override)

	if len(result) != 2 {
		t.Errorf("Expected length: 2, got: %d\n", len(result))
	}

	if result["FOO"] != "BAR" {
		t.Errorf("Expected FOO == BAR, got FOO == %s\n", result["FOO"])
	}
}
