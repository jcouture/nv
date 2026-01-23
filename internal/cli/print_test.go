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

package cli

import (
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestRunPrintExactMatch(t *testing.T) {
	orig := color.NoColor
	t.Cleanup(func() {
		color.NoColor = orig
	})
	color.NoColor = true

	t.Setenv("NVX_PRINT_TEST_FOO", "bar")
	opts := &printOptions{}

	stdout, _ := captureOutput(t, func() {
		if err := runPrint(opts, []string{"NVX_PRINT_TEST_FOO"}); err != nil {
			t.Fatalf("runPrint error: %v", err)
		}
	})

	if !strings.Contains(stdout, "NVX_PRINT_TEST_FOO=bar") {
		t.Fatalf("unexpected output: %s", stdout)
	}
}

func TestRunPrintIgnoreCase(t *testing.T) {
	orig := color.NoColor
	t.Cleanup(func() {
		color.NoColor = orig
	})
	color.NoColor = true

	t.Setenv("NVX_PRINT_TEST_BAR", "baz")
	opts := &printOptions{ignoreCase: true}

	stdout, _ := captureOutput(t, func() {
		if err := runPrint(opts, []string{"nv_print_test_bar"}); err != nil {
			t.Fatalf("runPrint error: %v", err)
		}
	})

	if !strings.Contains(stdout, "NVX_PRINT_TEST_BAR=baz") {
		t.Fatalf("unexpected output: %s", stdout)
	}
}

func TestRunPrintSort(t *testing.T) {
	orig := color.NoColor
	t.Cleanup(func() {
		color.NoColor = orig
	})
	color.NoColor = true

	t.Setenv("NVX_PRINT_TEST_SORT_A", "1")
	t.Setenv("NVX_PRINT_TEST_SORT_B", "2")

	opts := &printOptions{sort: true}

	stdout, _ := captureOutput(t, func() {
		if err := runPrint(opts, []string{"NVX_PRINT_TEST_SORT_"}); err != nil {
			t.Fatalf("runPrint error: %v", err)
		}
	})

	idxA := strings.Index(stdout, "NVX_PRINT_TEST_SORT_A=1")
	idxB := strings.Index(stdout, "NVX_PRINT_TEST_SORT_B=2")
	if idxA == -1 || idxB == -1 {
		t.Fatalf("unexpected output: %s", stdout)
	}
	if idxA > idxB {
		t.Fatalf("expected sorted output, got: %s", stdout)
	}
}
