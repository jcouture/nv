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

import "testing"

func TestExistsGetvarsGetnames(t *testing.T) {
	t.Setenv("NV_ENV_TEST", "value")

	if !Exists("NV_ENV_TEST") {
		t.Fatal("expected NV_ENV_TEST to exist")
	}
	if Exists("NV_ENV_MISSING") {
		t.Fatal("expected NV_ENV_MISSING to be absent")
	}

	vars := Getvars()
	if vars["NV_ENV_TEST"] != "value" {
		t.Fatalf("NV_ENV_TEST=%s want value", vars["NV_ENV_TEST"])
	}

	names := Getnames(map[string]string{"A": "1", "": "skip"})
	found := false
	for _, n := range names {
		if n == "A" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected names to contain A, got %v", names)
	}
}

func TestJoinOverrides(t *testing.T) {
	base := map[string]string{"A": "1"}
	override := map[string]string{"A": "2", "B": "3"}

	merged := Join(base, override)
	if merged["A"] != "2" || merged["B"] != "3" {
		t.Fatalf("merged=%v want A=2 B=3", merged)
	}
}

func TestSetvarsAndClear(t *testing.T) {
	vars := map[string]string{"NV_SET_TEST": "1", "NV_CLEAR_TEST": "2"}
	if err := Setvars(vars); err != nil {
		t.Fatalf("Setvars: %v", err)
	}

	cleared := Clear("NV_SET_TEST")
	if err := cleared; err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if !Exists("NV_SET_TEST") {
		t.Fatal("expected NV_SET_TEST to remain")
	}
	if Exists("NV_CLEAR_TEST") {
		t.Fatal("expected NV_CLEAR_TEST to be cleared")
	}
}
